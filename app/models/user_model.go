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

// UserCreateRequest represents the request body for creating a new user
type UserCreateRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	IsActive *bool  `json:"is_active"` // Pointer to allow optional and distinguish between false and unset
}

// UserUpdateRequest represents the request body for updating an existing user
type UserUpdateRequest struct {
	Username *string `json:"username"` // Pointer to allow partial updates
	Email    *string `json:"email" binding:"omitempty,email"`
	Password *string `json:"password"`
	IsActive *bool   `json:"is_active"`
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
