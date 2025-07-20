package models

import (
	"time"
)

// RevokedToken represents a blacklisted JWT
type RevokedToken struct {
	JTI       string    `gorm:"primaryKey"` // Tandai sebagai primary key
	ExpiresAt time.Time `json:"expires_at"`
	RevokedAt time.Time `json:"revoked_at"`
}

// TableName secara eksplisit memberi tahu GORM nama tabel yang harus digunakan.
func (RevokedToken) TableName() string {
	return "revoked_tokens"
}
