package config

import (
	"errors"
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
	if portStr == "" {
		return nil, errors.New("PORT environment variable not set")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	dbConn := os.Getenv("DB_CONNECT")
	if dbConn == "" {
		return nil, errors.New("DB_CONNECT environment variable not set")
	}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		return nil, errors.New("SERVICE_NAME environment variable not set")
	}

	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		return nil, errors.New("JAEGER_ENDPOINT environment variable not set")
	}

	environment := Environment(os.Getenv("ENVIRONMENT"))
	if environment == "" {
		return nil, errors.New("ENVIRONMENT environment variable not set")
	}

	return &Config{
		Port:           port,
		DBConn:         dbConn,
		ServiceName:    serviceName,
		JaegerEndpoint: jaegerEndpoint,
		Environment:    environment,
	}, nil
}
