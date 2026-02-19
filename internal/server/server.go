package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"GoApp/internal/database"
)

type Server struct {
	port int
	db database.Service
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

	NewServer := &Server{
		port: port,

		db: database.New(&database.DBConfig{
			Host:     cfg.DBHost,
			Port:     cfg.DBPort,
			Database: cfg.DBDatabase,
			Username: cfg.DBUsername,
			Password: cfg.DBPassword,
			Schema:   cfg.DBSchema,
		}),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(cfg),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := NewServer.db.Migrate(); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}	
	return server
}
