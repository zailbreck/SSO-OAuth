package repositories

import (
	"database/sql"
	"fmt" // Import core_models for RevokedToken
	"time"
)

// RevokedTokenRepository defines the interface for revoked token data operations
type RevokedTokenRepository interface {
	AddRevokedToken(jti string, expiresAt time.Time) error
	IsTokenRevoked(jti string) (bool, error)
}

// RevokedTokenRepositoryImpl is the implementation of RevokedTokenRepository
type RevokedTokenRepositoryImpl struct {
	db *sql.DB
}

// NewRevokedTokenRepository creates a new instance of RevokedTokenRepositoryImpl
func NewRevokedTokenRepository(db *sql.DB) RevokedTokenRepository {
	return &RevokedTokenRepositoryImpl{db: db}
}

// AddRevokedToken adds a token's JTI to the revoked_tokens table
func (r *RevokedTokenRepositoryImpl) AddRevokedToken(jti string, expiresAt time.Time) error {
	query := `INSERT INTO revoked_tokens (jti, expires_at, revoked_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, jti, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}
	return nil
}

// IsTokenRevoked checks if a token's JTI exists in the revoked_tokens table
func (r *RevokedTokenRepositoryImpl) IsTokenRevoked(jti string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM revoked_tokens WHERE jti = $1)`
	err := r.db.QueryRow(query, jti).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if token is revoked: %w", err)
	}
	return exists, nil // Return true if exists, false otherwise
}
