package repositories

import (
	"errors"
	"fmt"
	"sso-service/app/models" // Impor models
	"time"

	"gorm.io/gorm" // Impor GORM
)

// RevokedTokenRepository defines the interface for revoked token data operations
type RevokedTokenRepository interface {
	AddRevokedToken(jti string, expiresAt time.Time) error
	IsTokenRevoked(jti string) (bool, error)
}

// RevokedTokenRepositoryImpl is the implementation of RevokedTokenRepository
type RevokedTokenRepositoryImpl struct {
	db *gorm.DB // Diubah ke *gorm.DB
}

// NewRevokedTokenRepository creates a new instance of RevokedTokenRepositoryImpl
func NewRevokedTokenRepository(db *gorm.DB) RevokedTokenRepository {
	return &RevokedTokenRepositoryImpl{db: db}
}

// AddRevokedToken adds a token's JTI to the revoked_tokens table using GORM
func (r *RevokedTokenRepositoryImpl) AddRevokedToken(jti string, expiresAt time.Time) error {
	revokedToken := models.RevokedToken{
		JTI:       jti,
		ExpiresAt: expiresAt,
		RevokedAt: time.Now(),
	}
	result := r.db.Create(&revokedToken)
	if result.Error != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", result.Error)
	}
	return nil
}

// IsTokenRevoked checks if a token's JTI exists in the revoked_tokens table using GORM
func (r *RevokedTokenRepositoryImpl) IsTokenRevoked(jti string) (bool, error) {
	var token models.RevokedToken
	err := r.db.First(&token, "jti = ?", jti).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Jika token tidak ditemukan, berarti tidak dicabut (not revoked)
			return false, nil
		}
		// Untuk error lainnya
		return false, fmt.Errorf("error checking if token is revoked: %w", err)
	}

	// Jika token ditemukan tanpa error, berarti token tersebut sudah dicabut (revoked)
	return true, nil
}
