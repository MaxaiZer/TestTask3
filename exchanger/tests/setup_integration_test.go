package integration

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"log/slog"
	"net"
	"os"
	"test-task/api/gen/grpc/exchange"
	"test-task/exchanger/internal/app"
	"test-task/exchanger/internal/config"
	"test-task/exchanger/internal/models"
	"test-task/exchanger/internal/services"
	mygrpc "test-task/exchanger/internal/transport/grpc"
	"testing"
	"time"
)

var cfg *config.Config
var dbContainer testcontainers.Container
var rates = map[models.Currency]float64{
	models.USD: 1,
	models.EUR: 0.85,
	models.RUB: 0.1,
}
var exchangeClient exchange.ExchangeClient

func setupApp() error {

	var err error
	cfg, err = config.Get()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	storage, _, err := app.InitDB(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	reflection.Register(server)

	exchangeService := services.NewExchangeService(storage)
	exchangeServer := mygrpc.NewExchangeServer(exchangeService)

	exchange.RegisterExchangeServer(server, exchangeServer)

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to create port listener: %w", err)
	}

	go func() {
		if err = server.Serve(listener); err != nil {
			slog.Error("failed to serve grpc server", "error", err)
			os.Exit(1)
		}
	}()

	conn, err := grpc.NewClient(
		"localhost:"+cfg.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("failed to create grpc client: %w", err)
	}

	exchangeClient = exchange.NewExchangeClient(conn)
	return nil
}

func upEnvironment() {

	ctx := context.Background()

	db := "test_db"
	user := "postgres"
	password := "postgres"

	req := testcontainers.ContainerRequest{
		Image:        "postgres:17-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     user,
			"POSTGRES_PASSWORD": password,
			"POSTGRES_DB":       db,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5 * time.Second),
	}

	var err error
	dbContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		log.Fatalf("could not start PostgreSQL container: %s", err)
	}

	port, err := dbContainer.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("could not get port for PostgreSQL container: %s", err)
	}

	conn := fmt.Sprintf("host=localhost dbname=%s port=%d user=%s password=%s sslmode=disable",
		db, port.Int(), user, password)
	err = os.Setenv("DB_URL", conn)

	if err != nil {
		log.Fatalf("could not set environment variable DB_CONNECTION_STRING: %s", err)
	}
}

func downEnvironment() {
	ctx := context.Background()
	if err := dbContainer.Terminate(ctx); err != nil {
		fmt.Printf("Could not terminate PostgreSQL container: %s", err)
	}
}

func TestMain(m *testing.M) {

	err := os.Chdir("../") //project root to resolve correctly relative paths in code
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("CONFIG_PATH", "configs/config.env")

	upEnvironment()

	err = setupApp()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	downEnvironment()

	os.Exit(code)
}
