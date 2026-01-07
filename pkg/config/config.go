package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Environment string

const (
	Development Environment = "DEV"
	Production  Environment = "PROD"
)

type Config struct {
	Port           int         `env:"PORT,required"`
	DBConn         string      `env:"DB_CONNECT"`
	ServiceName    string      `env:"SERVICE_NAME"`
	JaegerEndpoint string      `env:"JAEGER_ENDPOINT"`
	Environment    Environment `env:"ENVIRONMENT"`
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	dbConn := os.Getenv("DB_CONNECT")

	return &Config{
		Port:   port,
		DBConn: dbConn,
	}, nil
}
