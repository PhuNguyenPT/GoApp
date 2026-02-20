package server

import (
	"context"
	"testing"
	"time"
)

func TestSessionCleanup(t *testing.T) {
	t.Run("calls DeleteExpiredSessions on tick", func(t *testing.T) {
		db := &mockDB{}
		s := &Server{db: db}
		tick := make(chan time.Time)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s.runSessionCleanup(ctx, tick)

		tick <- time.Now()
		tick <- time.Now() // blocks until goroutine processes first tick

		// by the time second tick is accepted, first is guaranteed processed
		time.Sleep(10 * time.Millisecond)

		if db.deleteExpiredSessionsCalled != 2 {
			t.Errorf("expected 2 calls, got %d", db.deleteExpiredSessionsCalled)
		}
	})

	t.Run("stops on context cancel", func(t *testing.T) {
		s := &Server{db: &mockDB{}}
		tick := make(chan time.Time)
		ctx, cancel := context.WithCancel(context.Background())

		s.runSessionCleanup(ctx, tick)
		cancel() // should cause goroutine to exit cleanly
		time.Sleep(10 * time.Millisecond)
	})
}
