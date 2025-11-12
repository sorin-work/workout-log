package store

import (
	"context"
	"time"
)

// User is an account that owns workouts.
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"displayName"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

const userColumns = `id, email, display_name, password_hash, created_at`

func scanUser(row interface{ Scan(...any) error }) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

// CreateUser inserts a user. It returns ErrConflict if the email is taken.
func (s *Store) CreateUser(ctx context.Context, email, displayName, passwordHash string) (User, error) {
	u, err := scanUser(s.pool.QueryRow(ctx,
		`INSERT INTO users (email, display_name, password_hash) VALUES ($1, $2, $3)
		 RETURNING `+userColumns, email, displayName, passwordHash))
	if isUniqueViolation(err) {
		return User{}, ErrConflict
	}
	return u, err
}

// UserByEmail looks a user up by email, ignoring case.
func (s *Store) UserByEmail(ctx context.Context, email string) (User, error) {
	u, err := scanUser(s.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE lower(email) = lower($1)`, email))
	return u, notFound(err)
}

// UserByID returns a user by id.
func (s *Store) UserByID(ctx context.Context, id int64) (User, error) {
	u, err := scanUser(s.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id))
	return u, notFound(err)
}
