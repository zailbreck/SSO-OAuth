package models

import (
	"time"

	"github.com/google/uuid"     // For UUID generation
	"golang.org/x/crypto/bcrypt" // For password hashing
)

// User represents the user data structure
type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`         // "-" hides this field from JSON output
	IsActive     bool      `json:"is_active"` // Indicates if the user account is active
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Role represents a role in the system
type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a specific permission
type Permission struct {
	ID          uuid.UUID  `json:"id"`
	SiteID      uuid.UUID  `json:"site_id"` // Foreign key to the sites table
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"` // Self-referencing foreign key for hierarchy. NULL for top-level permissions.
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UserRole is a junction table model for many-to-many relationship between User and Role
type UserRole struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
}

// RolePermission is a junction table model for many-to-many relationship between Role and Permission
type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
}

// RevokedToken represents a blacklisted JWT
type RevokedToken struct {
	JTI       string    `json:"jti"`        // JWT ID
	ExpiresAt time.Time `json:"expires_at"` // Original expiration time of the token
	RevokedAt time.Time `json:"revoked_at"` // When it was revoked
}

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a plain text password with a hashed password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
