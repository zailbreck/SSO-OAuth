package controllers

import (
	"net/http"
	"sso-service/app/models"
	"sso-service/app/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionController defines the interface for permission management operations
type PermissionController interface {
	CreatePermission(c *gin.Context)
	GetAllPermissions(c *gin.Context)
	GetPermissionByID(c *gin.Context)
	UpdatePermission(c *gin.Context)
	DeletePermission(c *gin.Context)
	AssignPermissionToRole(c *gin.Context)
	RemovePermissionFromRole(c *gin.Context)
}

// PermissionControllerImpl is the implementation of PermissionController
type PermissionControllerImpl struct {
	permissionService services.PermissionService
}

// NewPermissionController creates a new instance of PermissionControllerImpl
func NewPermissionController(permissionService services.PermissionService) PermissionController {
	return &PermissionControllerImpl{
		permissionService: permissionService,
	}
}

// CreatePermission handles creating a new permission. Requires 'permission:create' permission.
func (ctrl *PermissionControllerImpl) CreatePermission(c *gin.Context) {
	var req models.PermissionCreateRequest
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

	createdPermission, err := ctrl.permissionService.CreatePermission(jwtClaims, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "forbidden") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "not found") { // For site/parent not found
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

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"data":    createdPermission,
		"message": "Permission created successfully",
	})
}

// GetAllPermissions handles retrieving all permissions. Requires 'permission:read_all' permission.
func (ctrl *PermissionControllerImpl) GetAllPermissions(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	permissions, err := ctrl.permissionService.GetAllPermissions(jwtClaims)
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
		"data":    permissions,
		"message": "Permissions retrieved successfully",
	})
}

// GetPermissionByID handles retrieving a single permission by ID. Requires 'permission:read_all' permission.
func (ctrl *PermissionControllerImpl) GetPermissionByID(c *gin.Context) {
	permissionIDParam := c.Param("id")
	parsedPermissionID, err := uuid.Parse(permissionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid permission ID format",
		})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	permission, err := ctrl.permissionService.GetPermissionByID(jwtClaims, parsedPermissionID)
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
		"data":    permission,
		"message": "Permission retrieved successfully",
	})
}

// UpdatePermission handles updating an existing permission. Requires 'permission:update' permission.
func (ctrl *PermissionControllerImpl) UpdatePermission(c *gin.Context) {
	permissionIDParam := c.Param("id")
	parsedPermissionID, err := uuid.Parse(permissionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid permission ID format",
		})
		return
	}

	var req models.PermissionUpdateRequest
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

	updatedPermission, err := ctrl.permissionService.UpdatePermission(jwtClaims, parsedPermissionID, req)
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
		"data":    updatedPermission,
		"message": "Permission updated successfully",
	})
}

// DeletePermission handles deleting a permission. Requires 'permission:delete' permission.
func (ctrl *PermissionControllerImpl) DeletePermission(c *gin.Context) {
	permissionIDParam := c.Param("id")
	parsedPermissionID, err := uuid.Parse(permissionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid permission ID format",
		})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	err = ctrl.permissionService.DeletePermission(jwtClaims, parsedPermissionID)
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
		"message": "Permission deleted successfully",
	})
}

// AssignPermissionToRole handles assigning a permission to a role. Requires 'permission:assign' permission.
func (ctrl *PermissionControllerImpl) AssignPermissionToRole(c *gin.Context) {
	var req models.AssignPermissionToRoleRequest
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

	err := ctrl.permissionService.AssignPermissionToRole(jwtClaims, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "forbidden") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "already has this permission") {
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
		"data":    nil,
		"message": "Permission assigned to role successfully",
	})
}

// RemovePermissionFromRole handles removing a permission from a role. Requires 'permission:assign' permission.
func (ctrl *PermissionControllerImpl) RemovePermissionFromRole(c *gin.Context) {
	var req models.AssignPermissionToRoleRequest
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

	err := ctrl.permissionService.RemovePermissionFromRole(jwtClaims, req)
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
		"message": "Permission removed from role successfully",
	})
}
