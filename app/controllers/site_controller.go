package controllers

import (
	"net/http"
	"sso-service/app/models"
	"sso-service/app/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SiteController defines the interface for site management operations
type SiteController interface {
	CreateSite(c *gin.Context)
	GetAllSites(c *gin.Context)
	GetSiteByID(c *gin.Context)
	UpdateSite(c *gin.Context)
	DeleteSite(c *gin.Context)
}

// SiteControllerImpl is the implementation of SiteController
type SiteControllerImpl struct {
	siteService services.SiteService
}

// NewSiteController creates a new instance of SiteControllerImpl
func NewSiteController(siteService services.SiteService) SiteController {
	return &SiteControllerImpl{
		siteService: siteService,
	}
}

// CreateSite handles creating a new site. Requires 'site:create' permission.
func (ctrl *SiteControllerImpl) CreateSite(c *gin.Context) {
	var req models.SiteCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": err.Error(),
		})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	createdSite, err := ctrl.siteService.CreateSite(jwtClaims, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "forbidden") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "exists") {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{
			"status":  statusCode,
			"data":    nil,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"data":    createdSite,
		"message": "Site created successfully",
	})
}

// GetAllSites handles retrieving all sites. Requires 'site:read_all' permission.
func (ctrl *SiteControllerImpl) GetAllSites(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	sites, err := ctrl.siteService.GetAllSites(jwtClaims)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "forbidden") {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{
			"status":  statusCode,
			"data":    nil,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    sites,
		"message": "Sites retrieved successfully",
	})
}

// GetSiteByID handles retrieving a single site by ID. Requires 'site:read_all' permission.
func (ctrl *SiteControllerImpl) GetSiteByID(c *gin.Context) {
	siteIDParam := c.Param("id")
	parsedSiteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid site ID format",
		})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	site, err := ctrl.siteService.GetSiteByID(jwtClaims, parsedSiteID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "forbidden") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{
			"status":  statusCode,
			"data":    nil,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    site,
		"message": "Site retrieved successfully",
	})
}

// UpdateSite handles updating an existing site. Requires 'site:update' permission.
func (ctrl *SiteControllerImpl) UpdateSite(c *gin.Context) {
	siteIDParam := c.Param("id")
	parsedSiteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid site ID format",
		})
		return
	}

	var req models.SiteUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": err.Error(),
		})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	updatedSite, err := ctrl.siteService.UpdateSite(jwtClaims, parsedSiteID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "forbidden") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "exists") {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{
			"status":  statusCode,
			"data":    nil,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    updatedSite,
		"message": "Site updated successfully",
	})
}

// DeleteSite handles deleting a site. Requires 'site:delete' permission.
func (ctrl *SiteControllerImpl) DeleteSite(c *gin.Context) {
	siteIDParam := c.Param("id")
	parsedSiteID, err := uuid.Parse(siteIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid site ID format",
		})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	err = ctrl.siteService.DeleteSite(jwtClaims, parsedSiteID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "forbidden") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{
			"status":  statusCode,
			"data":    nil,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    nil,
		"message": "Site deleted successfully",
	})
}
