package app

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	consulapi "github.com/hashicorp/consul/api"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
	"os"
	"strconv"
	"sync"
	"test-task/api/gen/grpc/exchange"
	"test-task/exchanger/internal/config"
	"test-task/exchanger/internal/services"
	"test-task/exchanger/internal/storage/postgres"
	mygrpc "test-task/exchanger/internal/transport/grpc"
	"time"
)

type shutdownTask struct {
	name     string
	shutdown func(context.Context) error
}

type App struct {
	cfg       *config.Config
	server    *grpc.Server
	listener  net.Listener
	shutdowns []shutdownTask
}

func New() (*App, error) {

	var shutdowns []shutdownTask

	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	slog.Info("current environment", "env", cfg.Env)

	initLogger(cfg)

	storage, connector, err := InitDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	shutdowns = append(shutdowns, shutdownTask{
		name:     "db connection",
		shutdown: connector.Close,
	})

	tracer, err := initTracer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	shutdowns = append(shutdowns, shutdownTask{
		name:     "tracer",
		shutdown: tracer.Shutdown,
	})

	server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	reflection.Register(server)

	exchangeService := services.NewExchangeService(storage)
	exchangeServer := mygrpc.NewExchangeServer(exchangeService)
	exchange.RegisterExchangeServer(server, exchangeServer)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthgrpc.HealthCheckResponse_SERVING)
	healthgrpc.RegisterHealthServer(server, healthServer)

	consulShutdown, err := registerInConsul(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to register in consul: %w", err)
	}
	shutdowns = append(shutdowns, shutdownTask{
		name:     "consul",
		shutdown: consulShutdown,
	})

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to create port listener: %w", err)
	}
	return &App{cfg: cfg, server: server, listener: listener, shutdowns: shutdowns}, nil
}

func (a *App) Run() {
	if err := a.server.Serve(a.listener); err != nil {
		slog.Error("failed to serve grpc server", "error", err)
		os.Exit(1)
	}
}

func (a *App) Stop(ctx context.Context) {

	slog.Info("shutting down gracefully...")

	a.server.GracefulStop()
	slog.Info("grpc server gracefully stopped")

	wg := &sync.WaitGroup{}
	for _, task := range a.shutdowns {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("shutting down task", "name", task.name)
			if err := task.shutdown(ctx); err != nil {
				slog.Error("failed to shutdown gracefully", "task", task.name, "error", err)
			} else {
				slog.Info("task gracefully stopped", "name", task.name)
			}
		}()
	}

	wg.Wait()
}

func initLogger(cfg *config.Config) {
	var handler slog.Handler

	if cfg.Env == config.Production {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func InitDB(cfg *config.Config) (*postgres.Storage, *postgres.Connector, error) {
	connector, err := postgres.NewConnector(context.Background(), cfg.DbUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize postgres connector: %w", err)
	}

	err = postgres.RunMigrations(connector.DB(), cfg.MigrationsPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	storage, err := postgres.NewStorage(connector.Pool)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return storage, connector, nil
}

func initTracer(cfg *config.Config) (*trace.TracerProvider, error) {
	ctx := context.Background()

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OtelEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otel exporter: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
		)),
	)

	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}

func registerInConsul(cfg *config.Config) (func(ctx context.Context) error, error) {

	if cfg.ConsulAddress == "" {
		return nil, nil
	}

	client, err := consulapi.NewClient(&consulapi.Config{
		Address: cfg.ConsulAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	port, _ := strconv.Atoi(cfg.Port)
	serviceID := fmt.Sprintf("%s-%s", cfg.ServiceName, uuid.NewString())
	reg := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    cfg.ServiceName,
		Port:    port,
		Address: "exchanger",
		Check: &consulapi.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("exchanger:%s", cfg.Port),
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}

	err = client.Agent().ServiceRegister(reg)
	if err != nil {
		return nil, fmt.Errorf("failed to register service in consul: %w", err)
	}

	slog.Info("registered service in consul", "service", reg.Name, "port", reg.Port)

	shutdown := func(_ context.Context) error {
		slog.Info("deregistering service from consul", "serviceID", serviceID)
		return client.Agent().ServiceDeregister(serviceID)
	}
	return shutdown, nil
}
