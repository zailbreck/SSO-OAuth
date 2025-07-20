package controllers

import (
	"net/http"
	"sso-service/app/models"
	"sso-service/app/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RoleController defines the interface for role management operations
type RoleController interface {
	CreateRole(c *gin.Context)
	GetAllRoles(c *gin.Context)
	GetRoleByID(c *gin.Context)
	UpdateRole(c *gin.Context)
	DeleteRole(c *gin.Context)
	AssignRoleToUser(c *gin.Context)
	RemoveRoleFromUser(c *gin.Context)
}

// RoleControllerImpl is the implementation of RoleController
type RoleControllerImpl struct {
	roleService services.RoleService
}

// NewRoleController creates a new instance of RoleControllerImpl
func NewRoleController(roleService services.RoleService) RoleController {
	return &RoleControllerImpl{
		roleService: roleService,
	}
}

// CreateRole handles creating a new role. Requires 'role:create' permission.
func (ctrl *RoleControllerImpl) CreateRole(c *gin.Context) {
	var req models.RoleCreateRequest
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

	createdRole, err := ctrl.roleService.CreateRole(jwtClaims, req)
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
		"data":    createdRole,
		"message": "Role created successfully",
	})
}

// GetAllRoles handles retrieving all roles. Requires 'role:read_all' permission.
func (ctrl *RoleControllerImpl) GetAllRoles(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	roles, err := ctrl.roleService.GetAllRoles(jwtClaims)
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
		"data":    roles,
		"message": "Roles retrieved successfully",
	})
}

// GetRoleByID handles retrieving a single role by ID. Requires 'role:read_all' permission.
func (ctrl *RoleControllerImpl) GetRoleByID(c *gin.Context) {
	roleIDParam := c.Param("id")
	parsedRoleID, err := uuid.Parse(roleIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid role ID format",
		})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	role, err := ctrl.roleService.GetRoleByID(jwtClaims, parsedRoleID)
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
		"data":    role,
		"message": "Role retrieved successfully",
	})
}

// UpdateRole handles updating an existing role. Requires 'role:update' permission.
func (ctrl *RoleControllerImpl) UpdateRole(c *gin.Context) {
	roleIDParam := c.Param("id")
	parsedRoleID, err := uuid.Parse(roleIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid role ID format",
		})
		return
	}

	var req models.RoleUpdateRequest
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

	updatedRole, err := ctrl.roleService.UpdateRole(jwtClaims, parsedRoleID, req)
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
		"data":    updatedRole,
		"message": "Role updated successfully",
	})
}

// DeleteRole handles deleting a role. Requires 'role:delete' permission.
func (ctrl *RoleControllerImpl) DeleteRole(c *gin.Context) {
	roleIDParam := c.Param("id")
	parsedRoleID, err := uuid.Parse(roleIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid role ID format",
		})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	err = ctrl.roleService.DeleteRole(jwtClaims, parsedRoleID)
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
		"message": "Role deleted successfully",
	})
}

// AssignRoleToUser handles assigning a role to a user. Requires 'role:assign' permission.
func (ctrl *RoleControllerImpl) AssignRoleToUser(c *gin.Context) {
	var req models.AssignRoleToUserRequest
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

	err := ctrl.roleService.AssignRoleToUser(jwtClaims, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "forbidden") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "already has this role") {
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
		"message": "Role assigned to user successfully",
	})
}

// RemoveRoleFromUser handles removing a role from a user. Requires 'role:assign' permission.
func (ctrl *RoleControllerImpl) RemoveRoleFromUser(c *gin.Context) {
	var req models.AssignRoleToUserRequest
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

	err := ctrl.roleService.RemoveRoleFromUser(jwtClaims, req)
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
		"message": "Role removed from user successfully",
	})
}
