package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"GoApp/internal/database"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
)

type DB interface {
	Health() map[string]string
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error)
	CreateSession(ctx context.Context, arg database.CreateSessionParams) (database.Session, error)
	GetSessionByToken(ctx context.Context, token string) (database.Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteExpiredSessions(ctx context.Context) error
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

func (s *sqlDB) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return s.queries.GetUserByEmail(ctx, email)
}

func (s *sqlDB) GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error) {
	return s.queries.GetUserByID(ctx, id)
}

func (s *sqlDB) CreateSession(ctx context.Context, arg database.CreateSessionParams) (database.Session, error) {
	return s.queries.CreateSession(ctx, arg)
}

func (s *sqlDB) GetSessionByToken(ctx context.Context, token string) (database.Session, error) {
	return s.queries.GetSessionByToken(ctx, token)
}

func (s *sqlDB) DeleteSession(ctx context.Context, token string) error {
	return s.queries.DeleteSession(ctx, token)
}

func (s *sqlDB) DeleteExpiredSessions(ctx context.Context) error {
	return s.queries.DeleteExpiredSessions(ctx)
}

type Server struct {
	port int
	db   DB
	cfg  *Config
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
		cfg: cfg,
	}

	s.StartSessionCleanup(context.Background(), 1*time.Hour)
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.RegisterRoutes(cfg),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
