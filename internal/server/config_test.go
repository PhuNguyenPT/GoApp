package server

import (
	"maps"
	"os"
	"testing"
)

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for k, v := range env {
		os.Setenv(k, v)
	}
	t.Cleanup(func() {
		for k := range env {
			os.Unsetenv(k)
		}
	})
}

var validEnv = map[string]string{
	"PORT":              "8080",
	"APP_ENV":           "test",
	"GIN_MODE":          "debug",
	"POSTGRES_HOST":     "localhost",
	"POSTGRES_PORT":     "5432",
	"POSTGRES_DATABASE": "mydb",
	"POSTGRES_USERNAME": "myuser",
	"POSTGRES_PASSWORD": "mypassword",
	"POSTGRES_SCHEMA":   "public",
}

func TestLoadConfig_Valid(t *testing.T) {
	setEnv(t, validEnv)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %v", cfg.Port)
	}
	if cfg.GinMode != "debug" {
		t.Errorf("expected gin mode debug, got %v", cfg.GinMode)
	}
}

func TestLoadConfig_MissingRequired(t *testing.T) {
	// Start with valid env then remove one required field
	setEnv(t, validEnv)
	os.Unsetenv("PORT")

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for missing PORT, got nil")
	}
}

func TestLoadConfig_InvalidGinMode(t *testing.T) {
	env := maps.Clone(validEnv)
	env["GIN_MODE"] = "invalid"
	setEnv(t, env)

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for invalid GIN_MODE, got nil")
	}
}

func TestLoadConfig_ShortPassword(t *testing.T) {
	env := maps.Clone(validEnv)
	env["POSTGRES_PASSWORD"] = "short"
	setEnv(t, env)

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for password shorter than 8 chars, got nil")
	}
}

func TestLoadConfig_NonNumericDBPort(t *testing.T) {
	env := maps.Clone(validEnv)
	env["POSTGRES_PORT"] = "notaport"
	setEnv(t, env)

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for non-numeric POSTGRES_PORT, got nil")
	}
}
