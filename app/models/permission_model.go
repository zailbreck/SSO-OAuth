package models

import (
	"time"

	"github.com/google/uuid"
)

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

// PermissionCreateRequest represents the request body for creating a new permission
type PermissionCreateRequest struct {
	SiteID      uuid.UUID  `json:"site_id" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description *string    `json:"description"` // Optional
	ParentID    *uuid.UUID `json:"parent_id"`   // Optional
}

// PermissionUpdateRequest represents the request body for updating an existing permission
type PermissionUpdateRequest struct {
	SiteID      *uuid.UUID `json:"site_id"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"` // Can be set to null to remove parent
}

// AssignPermissionToRoleRequest represents the request body for assigning a permission to a role
type AssignPermissionToRoleRequest struct {
	RoleID       uuid.UUID `json:"role_id" binding:"required"`
	PermissionID uuid.UUID `json:"permission_id" binding:"required"`
}
