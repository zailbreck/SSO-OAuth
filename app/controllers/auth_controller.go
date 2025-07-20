package controllers

import (
	"net/http"
	"sso-service/app/services"
	"strings"

	"github.com/gin-gonic/gin"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthController defines the interface for authentication-related operations
type AuthController interface {
	Login(c *gin.Context)
	Refresh(c *gin.Context)
	Logout(c *gin.Context)
}

// AuthControllerImpl is the implementation of AuthController
type AuthControllerImpl struct {
	userService services.UserService
}

// NewAuthController creates a new instance of AuthControllerImpl
func NewAuthController(userService services.UserService) AuthController {
	return &AuthControllerImpl{
		userService: userService,
	}
}

// Login handles user login and returns JWT tokens with the desired custom format
func (ctrl *AuthControllerImpl) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"data":    nil,
			"message": err.Error(),
		})
		return
	}

	tokenResponse, err := ctrl.userService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"data":    nil,
			"message": err.Error(),
		})
		return
	}

	// Remove refresh_token from the data before sending
	responseData := gin.H{
		"access_token": tokenResponse.AccessToken,
		"token_type":   tokenResponse.TokenType,
		"expires_in":   tokenResponse.ExpiresIn,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    responseData,
		"message": "Login successful",
	})
}

// Refresh handles token refreshing using a refresh token from the Authorization header
func (ctrl *AuthControllerImpl) Refresh(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"data":    nil,
			"message": "Authorization header required",
		})
		return
	}

	refreshTokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if refreshTokenString == authHeader { // No "Bearer " prefix found
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"data":    nil,
			"message": "Bearer token required",
		})
		return
	}

	tokenResponse, err := ctrl.userService.RefreshToken(refreshTokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"data":    nil,
			"message": err.Error(),
		})
		return
	}

	// Return only the new refresh_token in the data field
	responseData := gin.H{
		"refresh_token": tokenResponse.RefreshToken,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    responseData,
		"message": "Token refreshed successfully",
	})
}

// Logout invalidates the user's session/token
func (ctrl *AuthControllerImpl) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"data":    nil,
			"message": "Authorization header required",
		})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader { // No "Bearer " prefix found
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"data":    nil,
			"message": "Bearer token required",
		})
		return
	}

	err := ctrl.userService.InvalidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"data":    nil,
			"message": "Failed to invalidate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    nil, // No data returned for logout
		"message": "Logged out successfully",
	})
}
