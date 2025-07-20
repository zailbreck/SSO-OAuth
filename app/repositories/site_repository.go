package repositories

import (
	"errors"
	"fmt"
	"sso-service/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SiteRepository defines the interface for site data operations
type SiteRepository interface {
	CreateSite(site models.Site) (models.Site, error)
	GetAllSites() ([]models.Site, error)
	GetSiteByID(id uuid.UUID) (models.Site, error)
	GetSiteBySlug(slug string) (models.Site, error)
	UpdateSite(site models.Site) (models.Site, error)
	DeleteSite(id uuid.UUID) error
}

// SiteRepositoryImpl is the implementation of SiteRepository
type SiteRepositoryImpl struct {
	db *gorm.DB // Diubah ke *gorm.DB
}

// NewSiteRepository creates a new instance of SiteRepositoryImpl
func NewSiteRepository(db *gorm.DB) SiteRepository {
	return &SiteRepositoryImpl{db: db}
}

// CreateSite inserts a new site into the database using GORM
func (r *SiteRepositoryImpl) CreateSite(site models.Site) (models.Site, error) {
	result := r.db.Create(&site)
	if result.Error != nil {
		return models.Site{}, fmt.Errorf("failed to create site: %w", result.Error)
	}
	return site, nil
}

// GetAllSites retrieves all sites from the database using GORM
func (r *SiteRepositoryImpl) GetAllSites() ([]models.Site, error) {
	var sites []models.Site
	result := r.db.Find(&sites)
	if result.Error != nil {
		return nil, fmt.Errorf("error getting all sites: %w", result.Error)
	}
	return sites, nil
}

// GetSiteByID retrieves a site by its ID from the database using GORM
func (r *SiteRepositoryImpl) GetSiteByID(id uuid.UUID) (models.Site, error) {
	var site models.Site
	result := r.db.First(&site, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Site{}, fmt.Errorf("site not found: %w", result.Error)
		}
		return models.Site{}, fmt.Errorf("error getting site by ID: %w", result.Error)
	}
	return site, nil
}

// GetSiteBySlug retrieves a site by its slug from the database using GORM
func (r *SiteRepositoryImpl) GetSiteBySlug(slug string) (models.Site, error) {
	var site models.Site
	result := r.db.Where("site_slug = ?", slug).First(&site)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Site{}, fmt.Errorf("site not found: %w", result.Error)
		}
		return models.Site{}, fmt.Errorf("error getting site by slug: %w", result.Error)
	}
	return site, nil
}

// UpdateSite updates an existing site in the database using GORM
func (r *SiteRepositoryImpl) UpdateSite(site models.Site) (models.Site, error) {
	if site.ID == uuid.Nil {
		return models.Site{}, errors.New("cannot update site without ID")
	}
	result := r.db.Save(&site)
	if result.Error != nil {
		return models.Site{}, fmt.Errorf("failed to update site: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.Site{}, gorm.ErrRecordNotFound
	}
	return site, nil
}

// DeleteSite deletes a site from the database using GORM
func (r *SiteRepositoryImpl) DeleteSite(id uuid.UUID) error {
	result := r.db.Delete(&models.Site{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete site: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("site not found for deletion")
	}
	return nil
}
