package models

import (
	"time"
)

// RevokedToken represents a blacklisted JWT
type RevokedToken struct {
	JTI       string    `json:"jti"`        // JWT ID
	ExpiresAt time.Time `json:"expires_at"` // Original expiration time of the token
	RevokedAt time.Time `json:"revoked_at"` // When it was revoked
}
