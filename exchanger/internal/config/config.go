package config

import (
	"flag"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"os"
)

type Environment string

const (
	Production  Environment = "production"
	Development Environment = "development"
)

type Config struct {
	Env            Environment
	ServiceName    string `validate:"required"`
	Port           string `validate:"required"`
	DbUrl          string `validate:"required"`
	MigrationsPath string `validate:"required"`
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

	cfg := Config{
		Env:            getEnvironment(),
		ServiceName:    os.Getenv("SERVICE_NAME"),
		Port:           os.Getenv("PORT"),
		DbUrl:          os.Getenv("DB_URL"),
		MigrationsPath: os.Getenv("MIGRATIONS_PATH"),
		OtelEndpoint:   os.Getenv("OTEL_ENDPOINT"),
		ConsulAddress:  os.Getenv("CONSUL_ADDRESS"),
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
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
