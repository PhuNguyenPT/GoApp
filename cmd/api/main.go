package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"GoApp/internal/server"
)

func loadMTLSConfig(cfg *server.Config) (*tls.Config, error) {
	caCert, err := os.ReadFile(cfg.TLSCAPath)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)
	cert, err := tls.LoadX509KeyPair(cfg.TLSCertPath, cfg.TLSKeyPath)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}
	log.Println("Server exiting")
	done <- true
}

func main() {
	cfg, err := server.LoadConfig()
	if err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	srv := server.NewServer(cfg)
	done := make(chan bool, 1)
	go gracefulShutdown(srv, done)

	if cfg.TLSCertPath != "" {
		tlsCfg, tlsErr := loadMTLSConfig(cfg)
		if tlsErr != nil {
			log.Fatalf("failed to load mTLS config: %v", tlsErr)
		}
		srv.TLSConfig = tlsCfg
		log.Printf("mTLS configured, starting HTTPS on %s", cfg.TLSPort)
		err = srv.ListenAndServeTLS(cfg.TLSCertPath, cfg.TLSKeyPath)
	} else {
		log.Printf("starting HTTP on %s", cfg.Port)
		err = srv.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}
	<-done
	log.Println("Graceful shutdown complete.")
}
