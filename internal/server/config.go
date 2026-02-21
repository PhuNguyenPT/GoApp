package server

import (
	"os"

	"github.com/go-playground/validator/v10"
)

const (
	EnvDev        = "dev"
	EnvProduction = "production"
	EnvTest       = "test"
)

type Config struct {
	Port       string `validate:"required"`
	AppEnv     string `validate:"required,oneof=dev production test"`
	GinMode    string `validate:"required,oneof=debug release test"`
	DBHost     string `validate:"required"`
	DBPort     string `validate:"required,numeric"`
	DBDatabase string `validate:"required"`
	DBUsername string `validate:"required"`
	DBPassword string `validate:"required,min=8"`
	DBSchema   string `validate:"required"`
	// mTLS cert paths (optional — defaults to Docker secrets paths)
	TLSCertPath string
	TLSKeyPath  string
	TLSCAPath   string
	TLSPort     string
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:        os.Getenv("PORT"),
		AppEnv:      os.Getenv("APP_ENV"),
		GinMode:     os.Getenv("GIN_MODE"),
		DBHost:      os.Getenv("POSTGRES_HOST"),
		DBPort:      os.Getenv("POSTGRES_PORT"),
		DBDatabase:  os.Getenv("POSTGRES_DATABASE"),
		DBUsername:  os.Getenv("POSTGRES_USERNAME"),
		DBPassword:  os.Getenv("POSTGRES_PASSWORD"),
		DBSchema:    os.Getenv("POSTGRES_SCHEMA"),
		TLSCertPath: getEnvOrDefault("TLS_CERT_PATH", "/run/secrets/go_crt"),
		TLSKeyPath:  getEnvOrDefault("TLS_KEY_PATH", "/run/secrets/go_key"),
		TLSCAPath:   getEnvOrDefault("TLS_CA_PATH", "/run/secrets/backend_ca"),
		TLSPort:     getEnvOrDefault("TLS_PORT", "8443"),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
