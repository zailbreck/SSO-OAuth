package controllers

import (
	"net/http"
	"sso-service/app/models"
	"sso-service/app/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserController defines the interface for user profile management and CRUD (authenticated requests)
type UserController interface {
	GetMe(c *gin.Context)
	CreateUser(c *gin.Context)
	GetUsers(c *gin.Context)
	GetUserByID(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
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
	// The user ID and claims are set in the context by the AuthMiddleware
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"data":    nil,
			"message": "User ID not found in context",
		})
		return
	}

	parsedUserID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"data":    nil,
			"message": "Invalid user ID format",
		})
		return
	}

	// Get claims from context for authorization checks within service if needed
	_, claimsExist := c.Get("claims")
	if !claimsExist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"data":    nil,
			"message": "User claims not found in context",
		})
		return
	}
	user, err := ctrl.userService.GetUserProfile(parsedUserID) // This method doesn't need claims for now
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"data":    nil,
			"message": "User profile not found",
		})
		return
	}

	// Get roles and permissions from context (set by AuthMiddleware)
	roles, _ := c.Get("roles")
	permissions, _ := c.Get("permissions")

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
		"roles":       roles,       // Already []string from middleware
		"permissions": permissions, // Already []string from middleware
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    responseData,
		"message": "User profile retrieved successfully",
	})
}

// CreateUser handles creating a new user. Requires 'user:create' permission.
func (ctrl *UserControllerImpl) CreateUser(c *gin.Context) {
	var req models.UserCreateRequest
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

	createdUser, err := ctrl.userService.CreateUser(jwtClaims, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "forbidden") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "exists") {
			statusCode = http.StatusConflict // 409 Conflict for resource already exists
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
		"data":    createdUser,
		"message": "User created successfully",
	})
}

// GetUsers handles retrieving all users. Requires 'user:read_all' permission.
func (ctrl *UserControllerImpl) GetUsers(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	users, err := ctrl.userService.GetAllUsers(jwtClaims)
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
		"data":    users,
		"message": "Users retrieved successfully",
	})
}

// GetUserByID handles retrieving a single user by ID. Requires 'user:read_all' or 'user:read_own'.
func (ctrl *UserControllerImpl) GetUserByID(c *gin.Context) {
	userIDParam := c.Param("id")
	parsedUserID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid user ID format",
		})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	user, err := ctrl.userService.GetUserByID(jwtClaims, parsedUserID)
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
		"data":    user,
		"message": "User retrieved successfully",
	})
}

// UpdateUser handles updating an existing user. Requires 'user:update_all' or 'user:update_own'.
func (ctrl *UserControllerImpl) UpdateUser(c *gin.Context) {
	userIDParam := c.Param("id")
	parsedUserID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid user ID format",
		})
		return
	}

	var req models.UserUpdateRequest
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

	updatedUser, err := ctrl.userService.UpdateUser(jwtClaims, parsedUserID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "forbidden") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "exists") {
			statusCode = http.StatusConflict // For username/email conflict
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
		"data":    updatedUser,
		"message": "User updated successfully",
	})
}

// DeleteUser handles deleting a user. Requires 'user:delete' permission.
func (ctrl *UserControllerImpl) DeleteUser(c *gin.Context) {
	userIDParam := c.Param("id")
	parsedUserID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": "Invalid user ID format",
		})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "data": nil, "message": "User claims not found"})
		return
	}
	jwtClaims := claims.(*services.JWTClaims)

	err = ctrl.userService.DeleteUser(jwtClaims, parsedUserID)
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
		"message": "User deleted successfully",
	})
}
