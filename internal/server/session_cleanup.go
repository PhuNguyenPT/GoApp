package server

import (
	"context"
	"log"
	"time"
)

func (s *Server) StartSessionCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	s.runSessionCleanup(ctx, ticker.C)
}

func (s *Server) runSessionCleanup(ctx context.Context, tick <-chan time.Time) {
	go func() {
		for {
			select {
			case <-tick:
				if err := s.db.DeleteExpiredSessions(ctx); err != nil {
					log.Printf("session cleanup error: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
