package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"GoApp/internal/database"
)

type DB interface {
	Health() map[string]string
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
}

type sqlDB struct {
	raw     *sql.DB
	queries *database.Queries
}

func (s *sqlDB) Health() map[string]string {
	return database.Health(s.raw)
}

func (s *sqlDB) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return s.queries.CreateUser(ctx, arg)
}

type Server struct {
	port int
	db   DB
}

func NewServer() *http.Server {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("invalid config %v", err)
	}

	port, err := strconv.Atoi(cfg.Port)
	if err != nil {
		log.Fatalf("invalid PORT value %v", err)
	}

	dbCfg := &database.DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		Database: cfg.DBDatabase,
		Username: cfg.DBUsername,
		Password: cfg.DBPassword,
		Schema:   cfg.DBSchema,
	}

	raw := database.Open(dbCfg)

	if err := database.Migrate(raw); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	s := &Server{
		port: port,
		db: &sqlDB{
			raw:     raw,
			queries: database.New(raw),
		},
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.RegisterRoutes(cfg),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}