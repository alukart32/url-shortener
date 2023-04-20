// Package pingpgx provides postgres ping functionality.
package pingpgx

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// pinger defines the postgres pinger.
type pinger struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

// Pinger returns a new postgres pinger.
func Pinger(pg *pgxpool.Pool, timeout time.Duration) *pinger {
	return &pinger{
		pool:    pg,
		timeout: timeout,
	}
}

// Ping pings the postgres instance.
func (s *pinger) Ping() error {
	if s.pool != nil {
		ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
		defer cancel()

		return s.pool.Ping(ctx)
	}
	return errors.New("corrupted ping")
}
