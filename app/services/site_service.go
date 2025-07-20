package services

import (
	"errors"
	"fmt"
	"sso-service/app/models"
	"sso-service/app/repositories"

	"github.com/google/uuid"
)

// SiteService defines the interface for site-related services
type SiteService interface {
	CreateSite(claims *JWTClaims, req models.SiteCreateRequest) (models.Site, error)
	GetAllSites(claims *JWTClaims) ([]models.Site, error)
	GetSiteByID(claims *JWTClaims, siteID uuid.UUID) (models.Site, error)
	UpdateSite(claims *JWTClaims, siteID uuid.UUID, req models.SiteUpdateRequest) (models.Site, error)
	DeleteSite(claims *JWTClaims, siteID uuid.UUID) error
}

// SiteServiceImpl is the implementation of SiteService
type SiteServiceImpl struct {
	siteRepository repositories.SiteRepository
	userService    UserService // To use authorization helpers
}

// NewSiteService creates a new instance of SiteServiceImpl
func NewSiteService(siteRepository repositories.SiteRepository, userService UserService) SiteService {
	return &SiteServiceImpl{
		siteRepository: siteRepository,
		userService:    userService,
	}
}

// CreateSite creates a new site. Requires 'site:create' permission.
func (s *SiteServiceImpl) CreateSite(claims *JWTClaims, req models.SiteCreateRequest) (models.Site, error) {
	if !s.userService.HasPermission(claims, "site:create") {
		return models.Site{}, errors.New("forbidden: insufficient permissions")
	}

	// Check if site slug already exists
	_, err := s.siteRepository.GetSiteBySlug(req.SiteSlug)
	if err == nil {
		return models.Site{}, errors.New("site slug already exists")
	}

	newSite := models.Site{
		SiteName: req.SiteName,
		SiteSlug: req.SiteSlug,
	}

	createdSite, err := s.siteRepository.CreateSite(newSite)
	if err != nil {
		return models.Site{}, fmt.Errorf("failed to create site: %w", err)
	}
	return createdSite, nil
}

// GetAllSites retrieves all sites. Requires 'site:read_all' permission.
func (s *SiteServiceImpl) GetAllSites(claims *JWTClaims) ([]models.Site, error) {
	if !s.userService.HasPermission(claims, "site:read_all") {
		return nil, errors.New("forbidden: insufficient permissions")
	}

	sites, err := s.siteRepository.GetAllSites()
	if err != nil {
		return nil, fmt.Errorf("failed to get all sites: %w", err)
	}
	return sites, nil
}

// GetSiteByID retrieves a site by ID. Requires 'site:read_all' permission.
func (s *SiteServiceImpl) GetSiteByID(claims *JWTClaims, siteID uuid.UUID) (models.Site, error) {
	if !s.userService.HasPermission(claims, "site:read_all") {
		return models.Site{}, errors.New("forbidden: insufficient permissions")
	}

	site, err := s.siteRepository.GetSiteByID(siteID)
	if err != nil {
		return models.Site{}, fmt.Errorf("failed to get site by ID: %w", err)
	}
	return site, nil
}

// UpdateSite updates an existing site. Requires 'site:update' permission.
func (s *SiteServiceImpl) UpdateSite(claims *JWTClaims, siteID uuid.UUID, req models.SiteUpdateRequest) (models.Site, error) {
	if !s.userService.HasPermission(claims, "site:update") {
		return models.Site{}, errors.New("forbidden: insufficient permissions")
	}

	existingSite, err := s.siteRepository.GetSiteByID(siteID)
	if err != nil {
		return models.Site{}, errors.New("site not found")
	}

	if req.SiteName != nil {
		existingSite.SiteName = *req.SiteName
	}
	if req.SiteSlug != nil {
		// Check for slug uniqueness if changed
		if *req.SiteSlug != existingSite.SiteSlug {
			_, err := s.siteRepository.GetSiteBySlug(*req.SiteSlug)
			if err == nil {
				return models.Site{}, errors.New("site slug already exists")
			}
		}
		existingSite.SiteSlug = *req.SiteSlug
	}

	updatedSite, err := s.siteRepository.UpdateSite(existingSite)
	if err != nil {
		return models.Site{}, fmt.Errorf("failed to update site: %w", err)
	}
	return updatedSite, nil
}

// DeleteSite deletes a site. Requires 'site:delete' permission.
func (s *SiteServiceImpl) DeleteSite(claims *JWTClaims, siteID uuid.UUID) error {
	if !s.userService.HasPermission(claims, "site:delete") {
		return errors.New("forbidden: insufficient permissions")
	}

	err := s.siteRepository.DeleteSite(siteID)
	if err != nil {
		return fmt.Errorf("failed to delete site: %w", err)
	}
	return nil
}
