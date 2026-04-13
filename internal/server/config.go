package server

import (
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
)

const (
	EnvDev        = "dev"
	EnvProduction = "production"
	EnvTest       = "test"
)

type Config struct {
	Port    string `validate:"required"`
	AppEnv  string `validate:"required,oneof=dev production test"`
	GinMode string `validate:"required,oneof=debug release test"`
	// Database
	DBHost     string `validate:"required"`
	DBPort     string `validate:"required,numeric"`
	DBDatabase string `validate:"required"`
	DBUsername string `validate:"required"`
	DBPassword string `validate:"required,min=8"`
	DBSchema   string `validate:"required"`
	// Pagination
	PageSize    int
	MaxPageSize int
	// mTLS cert paths (optional — defaults to Docker secrets paths)
	TLSCertPath string
	TLSKeyPath  string
	TLSCAPath   string
	TLSPort     string
}

func getEnvOrDefaultStr(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvOrDefaultInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
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
		PageSize:    getEnvOrDefaultInt("PAGE_SIZE", 20),
		MaxPageSize: getEnvOrDefaultInt("MAX_PAGE_SIZE", 100),
		TLSCertPath: getEnvOrDefaultStr("TLS_CERT_PATH", "/run/secrets/go_crt"),
		TLSKeyPath:  getEnvOrDefaultStr("TLS_KEY_PATH", "/run/secrets/go_key"),
		TLSCAPath:   getEnvOrDefaultStr("TLS_CA_PATH", "/run/secrets/backend_ca"),
		TLSPort:     getEnvOrDefaultStr("TLS_PORT", "8443"),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
