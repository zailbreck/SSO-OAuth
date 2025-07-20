package repositories

import (
	"database/sql"
	"fmt"
	"sso-service/app/models"
	"time"

	"github.com/google/uuid"
)

// SiteRepository defines the interface for site data operations
type SiteRepository interface {
	CreateSite(site models.Site) (models.Site, error)
	GetAllSites() ([]models.Site, error)
	GetSiteByID(id uuid.UUID) (models.Site, error)
	GetSiteBySlug(slug string) (models.Site, error) // New helper
	UpdateSite(id uuid.UUID, site models.Site) (models.Site, error)
	DeleteSite(id uuid.UUID) error
}

// SiteRepositoryImpl is the implementation of SiteRepository
type SiteRepositoryImpl struct {
	db *sql.DB
}

// NewSiteRepository creates a new instance of SiteRepositoryImpl
func NewSiteRepository(db *sql.DB) SiteRepository {
	return &SiteRepositoryImpl{db: db}
}

// CreateSite inserts a new site into the database
func (r *SiteRepositoryImpl) CreateSite(site models.Site) (models.Site, error) {
	site.ID = uuid.New()
	site.CreatedAt = time.Now()
	site.UpdatedAt = time.Now()

	query := `INSERT INTO sites (id, site_name, site_slug, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.db.QueryRow(query, site.ID, site.SiteName, site.SiteSlug, site.CreatedAt, site.UpdatedAt).Scan(&site.ID)
	if err != nil {
		return models.Site{}, fmt.Errorf("failed to create site: %w", err)
	}
	return site, nil
}

// GetAllSites retrieves all sites from the database
func (r *SiteRepositoryImpl) GetAllSites() ([]models.Site, error) {
	query := `SELECT id, site_name, site_slug, created_at, updated_at FROM sites`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting all sites: %w", err)
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var site models.Site
		if err := rows.Scan(&site.ID, &site.SiteName, &site.SiteSlug, &site.CreatedAt, &site.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning site: %w", err)
		}
		sites = append(sites, site)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sites: %w", err)
	}
	return sites, nil
}

// GetSiteByID retrieves a site by its ID from the database
func (r *SiteRepositoryImpl) GetSiteByID(id uuid.UUID) (models.Site, error) {
	var site models.Site
	query := `SELECT id, site_name, site_slug, created_at, updated_at FROM sites WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&site.ID, &site.SiteName, &site.SiteSlug, &site.CreatedAt, &site.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Site{}, fmt.Errorf("site not found: %w", err)
		}
		return models.Site{}, fmt.Errorf("error getting site by ID: %w", err)
	}
	return site, nil
}

// GetSiteBySlug retrieves a site by its slug from the database
func (r *SiteRepositoryImpl) GetSiteBySlug(slug string) (models.Site, error) {
	var site models.Site
	query := `SELECT id, site_name, site_slug, created_at, updated_at FROM sites WHERE site_slug = $1`
	err := r.db.QueryRow(query, slug).Scan(&site.ID, &site.SiteName, &site.SiteSlug, &site.CreatedAt, &site.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Site{}, fmt.Errorf("site not found: %w", err)
		}
		return models.Site{}, fmt.Errorf("error getting site by slug: %w", err)
	}
	return site, nil
}

// UpdateSite updates an existing site in the database
func (r *SiteRepositoryImpl) UpdateSite(id uuid.UUID, site models.Site) (models.Site, error) {
	site.UpdatedAt = time.Now()
	query := `UPDATE sites SET site_name = $1, site_slug = $2, updated_at = $3 WHERE id = $4 RETURNING id`
	err := r.db.QueryRow(query, site.SiteName, site.SiteSlug, site.UpdatedAt, id).Scan(&site.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Site{}, fmt.Errorf("site not found for update: %w", err)
		}
		return models.Site{}, fmt.Errorf("failed to update site: %w", err)
	}
	return site, nil
}

// DeleteSite deletes a site from the database
func (r *SiteRepositoryImpl) DeleteSite(id uuid.UUID) error {
	query := `DELETE FROM sites WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete site: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("site not found for deletion")
	}
	return nil
}
