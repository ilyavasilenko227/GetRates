package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"DEBUG"`

	AppHost string `env:"APP_HOST" envDefault:"0.0.0.0"`
	AppPort string `env:"APP_PORT" envDefault:"8080"`

	DbHost     string `env:"POSTGRES_HOST"`
	DbPort     string `env:"POSTGRES_PORT"`
	DbUser     string `env:"POSTGRES_USER"`
	DbName     string `env:"POSTGRES_DB"`
	DbPassword string `env:"POSTGRES_PASSWORD"`

	OTELExporterOTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"http://localhost:4318"`

	PrometheusHost string `env:"PROMETHEUS_HOST" envDefault:"0.0.0.0"`
	PrometheusPort string `env:"PROMETHEUS_PORT" envDefault:"8081"`
}

func ReadConfig() (*Config, error) {
	dbHost := flag.String("host", "", "Postgres host")
	dbPort := flag.String("port", "", "Postgres port")
	dbUser := flag.String("user", "", "Postgres user")
	dbName := flag.String("dbname", "", "Postgres database name")
	dbPassword := flag.String("password", "", "Postgres password")
	flag.Parse()

	config := Config{}

	err := env.Parse(&config)
	if err != nil {
		return nil, fmt.Errorf("read config error: %w", err)
	}
	// Переопределение значений из флагов командной строки, если они заданы
	if *dbHost != "" {
		config.DbHost = *dbHost
	}
	if *dbPort != "" {
		config.DbPort = *dbPort
	}
	if *dbUser != "" {
		config.DbUser = *dbUser
	}
	if *dbName != "" {
		config.DbName = *dbName
	}
	if *dbPassword != "" {
		config.DbPassword = *dbPassword
	}

	return &config, err
}
