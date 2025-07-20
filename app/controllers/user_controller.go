package controllers

import (
	"net/http"
	"sso-service/app/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserController defines the interface for user profile management (authenticated requests)
type UserController interface {
	GetMe(c *gin.Context)
}

// UserControllerImpl is the implementation of UserController
type UserControllerImpl struct {
	userService services.UserService
}

// NewUserController creates a new instance of UserControllerImpl
func NewUserController(userService services.UserService) UserController {
	return &UserControllerImpl{
		userService: userService,
	}
}

// GetMe returns the current authenticated user's profile and JWT claims with the desired custom format
func (ctrl *UserControllerImpl) GetMe(c *gin.Context) {
	// The user ID is set in the context by the AuthMiddleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"data":    nil,
			"message": "User ID not found in context",
		})
		return
	}

	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"data":    nil,
			"message": "Invalid user ID format",
		})
		return
	}

	user, err := ctrl.userService.GetUserProfile(parsedUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"data":    nil,
			"message": "User profile not found",
		})
		return
	}

	// Get roles from context (set by AuthMiddleware)
	roles, rolesExist := c.Get("roles")
	var userRoles []string
	if rolesExist {
		if r, ok := roles.([]string); ok {
			userRoles = r
		}
	}

	// Get permissions from context (set by AuthMiddleware)
	permissions, permissionsExist := c.Get("permissions")
	var userPermissions []string
	if permissionsExist {
		if perms, ok := permissions.([]string); ok {
			userPermissions = perms
		}
	}

	// Get the current access token from the Authorization header (no longer returned in data)
	authHeader := c.GetHeader("Authorization")
	_ = strings.TrimPrefix(authHeader, "Bearer ") // Still extract, but not use in response

	// Construct the profile data payload
	profileData := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	// Construct the main data payload with separate keys
	responseData := gin.H{
		"profile":     profileData,
		"roles":       userRoles,
		"permissions": userPermissions,
		// access_token is no longer included here as requested
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    responseData,
		"message": "User profile retrieved successfully",
	})
}
