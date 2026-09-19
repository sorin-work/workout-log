package store

import (
	"context"
	"time"
)

// CreateSession stores a session token hash for a user.
func (s *Store) CreateSession(ctx context.Context, tokenHash []byte, userID int64, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		tokenHash, userID, expiresAt)
	return err
}

// UserBySession returns the user owning a session that has not expired.
func (s *Store) UserBySession(ctx context.Context, tokenHash []byte) (User, error) {
	u, err := scanUser(s.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.display_name, u.unit, u.timezone, u.e1rm_formula, u.password_hash, u.created_at
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = $1 AND s.expires_at > now()`, tokenHash))
	return u, notFound(err)
}

// DeleteSession removes a session, e.g. on logout.
func (s *Store) DeleteSession(ctx context.Context, tokenHash []byte) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}

// DeleteExpiredSessions removes sessions past their expiry.
func (s *Store) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	tag, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= now()`)
	return tag.RowsAffected(), err
}

// DeleteOtherSessions ends every session of the user except keepTokenHash
// and returns how many were removed.
func (s *Store) DeleteOtherSessions(ctx context.Context, userID int64, keepTokenHash []byte) (int64, error) {
	tag, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1 AND token_hash <> $2`, userID, keepTokenHash)
	return tag.RowsAffected(), err
}
