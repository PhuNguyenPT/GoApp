package server

import (
	"os"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	Port       string `validate:"required"`
	AppEnv     string `validate:"required"`
	GinMode    string `validate:"required,oneof=debug release test"`
	DBHost     string `validate:"required"`
	DBPort     string `validate:"required,numeric"`
	DBDatabase string `validate:"required"`
	DBUsername string `validate:"required"`
	DBPassword string `validate:"required,min=8"`
	DBSchema   string `validate:"required"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:       os.Getenv("PORT"),
		AppEnv:     os.Getenv("APP_ENV"),
		GinMode:    os.Getenv("GIN_MODE"),
		DBHost:     os.Getenv("POSTGRES_HOST"),
		DBPort:     os.Getenv("POSTGRES_PORT"),
		DBDatabase: os.Getenv("POSTGRES_DATABASE"),
		DBUsername: os.Getenv("POSTGRES_USERNAME"),
		DBPassword: os.Getenv("POSTGRES_PASSWORD"),
		DBSchema:   os.Getenv("POSTGRES_SCHEMA"),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
