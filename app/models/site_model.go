package models

import (
	"time"

	"github.com/google/uuid"
)

// Site represents a site/application in the system
type Site struct {
	ID        uuid.UUID `json:"id"`
	SiteName  string    `json:"site_name"`
	SiteSlug  string    `json:"site_slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SiteCreateRequest represents the request body for creating a new site
type SiteCreateRequest struct {
	SiteName string `json:"site_name" binding:"required"`
	SiteSlug string `json:"site_slug" binding:"required"`
}

// SiteUpdateRequest represents the request body for updating an existing site
type SiteUpdateRequest struct {
	SiteName *string `json:"site_name"`
	SiteSlug *string `json:"site_slug"`
}
