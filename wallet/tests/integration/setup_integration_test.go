package integration

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"test-task/wallet/internal/app"
	"test-task/wallet/internal/config"
	"test-task/wallet/internal/domain/models"
	"test-task/wallet/internal/services"
	"test-task/wallet/internal/transport/http"
	"testing"
	"time"
)

var server *echo.Echo
var dbContainer testcontainers.Container
var rates = map[models.Currency]float64{
	models.USD: 1,
	models.EUR: 0.85,
	models.RUB: 0.1,
}

const apiPrefix = "/api/v1/"

func setupApp() error {

	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	storage, _, err := app.InitDB(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	exchanger := newExchangerClientMock(models.USD, rates)

	jwt, err := services.NewJwtService(services.JWTConfig{
		SecretKey: cfg.JwtSecret,
		Lifetime:  cfg.JwtLifetime,
		Issuer:    cfg.JwtIssuer,
		Audience:  cfg.JwtAudience,
	})
	if err != nil {
		return fmt.Errorf("failed to create jwt service: %w", err)
	}

	wallet := services.NewWalletService(storage, exchanger, redisMock{})
	auth := services.NewAuthService(jwt, storage)

	server = http.NewServer(http.Config{
		ServiceName:   "",
		JwtSecret:     cfg.JwtSecret,
		LaunchSwagger: false,
	}, wallet, auth)
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

	err := os.Chdir("../../") //project root to resolve correctly relative paths in code
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
