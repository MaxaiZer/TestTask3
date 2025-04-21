package config

import (
	"flag"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type Environment string

const (
	Production  Environment = "production"
	Development Environment = "development"
)

type Config struct {
	Env            Environment
	Port           string        `validate:"required"`
	ServiceName    string        `validate:"required"`
	JwtSecret      string        `validate:"required"`
	JwtLifetime    time.Duration `validate:"required,gt=0"`
	JwtIssuer      string        `validate:"required"`
	JwtAudience    string        `validate:"required"`
	ExchangerUrl   string
	DbUrl          string `validate:"required"`
	MigrationsPath string `validate:"required"`
	RedisAddress   string `validate:"required"`
	RedisPassword  string `validate:"required"`
	OtelEndpoint   string `validate:"required"`
	ConsulAddress  string
}

var flagSet = false

func Get() (*Config, error) {

	if path := getConfigPath(); path != "" {
		if err := godotenv.Load(path); err != nil {
			return nil, err
		}
	}

	jwtLifetime, err := strconv.Atoi(os.Getenv("JWT_LIFETIME"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert jwt lifetime: %w", err)
	}

	cfg := Config{
		Env:            getEnvironment(),
		Port:           os.Getenv("PORT"),
		ServiceName:    os.Getenv("SERVICE_NAME"),
		JwtSecret:      os.Getenv("JWT_SECRET"),
		JwtLifetime:    time.Duration(jwtLifetime) * time.Second,
		JwtIssuer:      os.Getenv("JWT_ISSUER"),
		JwtAudience:    os.Getenv("JWT_AUDIENCE"),
		ExchangerUrl:   os.Getenv("EXCHANGER_URL"),
		DbUrl:          os.Getenv("DB_URL"),
		MigrationsPath: os.Getenv("MIGRATIONS_PATH"),
		RedisAddress:   os.Getenv("REDIS_ADDRESS"),
		RedisPassword:  os.Getenv("REDIS_PASSWORD"),
		OtelEndpoint:   os.Getenv("OTEL_ENDPOINT"),
		ConsulAddress:  os.Getenv("CONSUL_ADDRESS"),
	}

	validate := validator.New()
	if err = validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	if cfg.ConsulAddress == "" && cfg.ExchangerUrl == "" {
		return nil, fmt.Errorf("exchangerUrl is required when consulAddress is empty")
	}

	return &cfg, nil
}

func getConfigPath() string {

	var path string

	if !flagSet {
		flag.StringVar(&path, "config", "", "path to config file")
		flag.Parse()
	}
	flagSet = true

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	return path
}

func getEnvironment() Environment {
	env := Environment(os.Getenv("ENV"))
	switch env {
	case Production, Development:
		return env
	default:
		return Development
	}
}
