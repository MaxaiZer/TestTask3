package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"log/slog"
	nethttp "net/http"
	"os"
	"sync"
	"test-task/wallet/internal/clients"
	"test-task/wallet/internal/config"
	"test-task/wallet/internal/services"
	"test-task/wallet/internal/storage/postgres"
	"test-task/wallet/internal/storage/redis"
	"test-task/wallet/internal/tracing"
	"test-task/wallet/internal/transport/http"
)

type App struct {
	cfg       *config.Config
	server    *echo.Echo
	shutdowns []func(context.Context) error
}

func New() (*App, error) {

	var shutdowns []func(context.Context) error

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

	shutdowns = append(shutdowns, func(_ context.Context) error { connector.Close(); return nil })

	tracer, err := tracing.InitTracer(cfg.ServiceName, cfg.OtelEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	shutdowns = append(shutdowns, tracer.Shutdown)

	cache, err := redis.New(redis.Config{
		Address:  cfg.RedisAddress,
		Password: cfg.RedisPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	shutdowns = append(shutdowns, func(_ context.Context) error { return cache.Close() })

	exchanger, err := createExchangerClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exchanger client: %w", err)
	}

	jwt, err := services.NewJwtService(services.JWTConfig{
		SecretKey: cfg.JwtSecret,
		Lifetime:  cfg.JwtLifetime,
		Issuer:    cfg.JwtIssuer,
		Audience:  cfg.JwtAudience,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create jwt service: %w", err)
	}

	wallet := services.NewWalletService(storage, exchanger, cache)
	auth := services.NewAuthService(jwt, storage)

	server := startServer(cfg, auth, wallet)

	return &App{cfg: cfg, server: server, shutdowns: shutdowns}, nil
}

func (a *App) Run() {
	if err := a.server.Start(":" + a.cfg.Port); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}

func (a *App) Stop(ctx context.Context) {

	slog.Info("shutting down gracefully...")

	if err := a.server.Shutdown(ctx); err != nil {
		slog.Error("failed to gracefully shutdown server", "error", err)
	} else {
		slog.Info("HTTP server gracefully stopped")
	}

	wg := &sync.WaitGroup{}
	for _, shutdown := range a.shutdowns {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := shutdown(ctx); err != nil {
				slog.Error("failed to shutdown gracefully", "error", err)
			}
		}()
	}

	wg.Wait()
	slog.Info("application stopped")
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
	storage := postgres.NewStorage(connector.Pool)

	err = postgres.RunMigrations(connector.DB(), cfg.MigrationsPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return storage, connector, nil
}

func createExchangerClient(cfg *config.Config) (*clients.ExchangerClient, error) {
	if cfg.ConsulAddress != "" {
		slog.Info("using consul address", "address", cfg.ConsulAddress)
		return clients.NewExchangerClientWithConsul(cfg.ConsulAddress)
	}
	slog.Info("using exchanger url", "address", cfg.ExchangerUrl)
	return clients.NewExchangerClient(cfg.ExchangerUrl)
}

func startServer(cfg *config.Config, auth http.AuthService, wallet http.WalletService) *echo.Echo {

	serverConfig := http.Config{
		ServiceName:   cfg.ServiceName,
		JwtSecret:     cfg.JwtSecret,
		LaunchSwagger: cfg.Env == config.Development,
	}
	return http.NewServer(serverConfig, wallet, auth)
}
