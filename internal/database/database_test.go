package database

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testCfg *DBConfig

func mustStartPostgresContainer() (func(context.Context, ...testcontainers.TerminateOption) error, error) {
	dbContainer, err := postgres.Run(
		context.Background(),
		"postgres:latest",
		postgres.WithDatabase("database"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	dbHost, err := dbContainer.Host(context.Background())
	if err != nil {
		return dbContainer.Terminate, err
	}

	dbPort, err := dbContainer.MappedPort(context.Background(), "5432/tcp")
	if err != nil {
		return dbContainer.Terminate, err
	}

	testCfg = &DBConfig{
		Host:     dbHost,
		Port:     dbPort.Port(),
		Database: "database",
		Username: "user",
		Password: "password",
		Schema:   "public",
	}
	return dbContainer.Terminate, nil
}

func TestMain(m *testing.M) {
	teardown, err := mustStartPostgresContainer()
	if err != nil {
		log.Fatalf("could not start postgres container: %v", err)
	}
	m.Run()
	if teardown != nil {
		if err := teardown(context.Background()); err != nil {
			log.Fatalf("could not teardown postgres container: %v", err)
		}
	}
}

func TestOpen(t *testing.T) {
	db := Open(testCfg)
	if db == nil {
		t.Fatal("Open() returned nil")
	}
	db.Close()
}

func TestMigrate(t *testing.T) {
	db := Open(testCfg)
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}
}

func TestHealth(t *testing.T) {
	db := Open(testCfg)
	defer db.Close()

	stats := Health(db)
	if stats["status"] != "up" {
		t.Fatalf("expected status up, got %s", stats["status"])
	}
	if _, ok := stats["error"]; ok {
		t.Fatalf("expected no error key in stats")
	}
	if stats["message"] != "It's healthy" {
		t.Fatalf("expected healthy message, got %s", stats["message"])
	}
}
