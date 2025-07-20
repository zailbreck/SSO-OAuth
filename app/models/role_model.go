package models

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a role in the system
type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleCreateRequest represents the request body for creating a new role
type RoleCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"` // Optional
}

// RoleUpdateRequest represents the request body for updating an existing role
type RoleUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// AssignRoleToUserRequest represents the request body for assigning a role to a user
type AssignRoleToUserRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}
