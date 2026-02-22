package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mime"
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
	UpdateUserName(ctx context.Context, arg database.UpdateUserNameParams) (database.User, error)
	UpdateUserPassword(ctx context.Context, arg database.UpdateUserPasswordParams) error
	DeleteUser(ctx context.Context, id uuid.UUID) error

	CreateSession(ctx context.Context, arg database.CreateSessionParams) (database.Session, error)
	GetSessionByToken(ctx context.Context, token string) (database.Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteExpiredSessions(ctx context.Context) error
	GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]database.Session, error)
	DeleteSessionByID(ctx context.Context, arg database.DeleteSessionByIDParams) error

	CreateContact(ctx context.Context, arg database.CreateContactParams) (database.Contact, error)
	CountContactsByIPToday(ctx context.Context, ipAddress string) (int64, error)
	CountContactsByEmailToday(ctx context.Context, email string) (int64, error)
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

func (s *sqlDB) UpdateUserName(ctx context.Context, arg database.UpdateUserNameParams) (database.User, error) {
	return s.queries.UpdateUserName(ctx, arg)
}

func (s *sqlDB) UpdateUserPassword(ctx context.Context, arg database.UpdateUserPasswordParams) error {
	return s.queries.UpdateUserPassword(ctx, arg)
}

func (s *sqlDB) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeleteUser(ctx, id)
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

func (s *sqlDB) GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]database.Session, error) {
	return s.queries.GetActiveSessionsByUserID(ctx, userID)
}

func (s *sqlDB) DeleteSessionByID(ctx context.Context, arg database.DeleteSessionByIDParams) error {
	return s.queries.DeleteSessionByID(ctx, arg)
}

func (s *sqlDB) CreateContact(ctx context.Context, arg database.CreateContactParams) (database.Contact, error) {
	return s.queries.CreateContact(ctx, arg)
}

func (s *sqlDB) CountContactsByIPToday(ctx context.Context, ipAddress string) (int64, error) {
	return s.queries.CountContactsByIPToday(ctx, ipAddress)
}

func (s *sqlDB) CountContactsByEmailToday(ctx context.Context, email string) (int64, error) {
	return s.queries.CountContactsByEmailToday(ctx, email)
}

type Server struct {
	port int
	db   DB
	cfg  *Config
}

func init() {
	if err := mime.AddExtensionType(".webmanifest", "application/manifest+json"); err != nil {
		log.Fatalf("failed to register .webmanifest MIME type: %v", err)
	}
}

func NewServer(cfg *Config) *http.Server {
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
		Addr:         fmt.Sprintf(":%s", cfg.TLSPort),
		Handler:      s.RegisterRoutes(cfg),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
