// Package store contains the PostgreSQL queries used by the API.
package store

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrNotFound is returned when a row does not exist or belongs to
	// another user.
	ErrNotFound = errors.New("not found")
	// ErrConflict is returned when a unique constraint is violated.
	ErrConflict = errors.New("conflict")
)

// Store wraps the connection pool.
type Store struct {
	pool *pgxpool.Pool
}

// New returns a store backed by pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func itoa(n int) string { return strconv.Itoa(n) }

// Ping checks that the database answers.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
