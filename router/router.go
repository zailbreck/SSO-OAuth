package router

import (
	"database/sql" // Import database/sql
	"net/http"
	"sso-service/app/controllers"
	"sso-service/app/repositories" // Import the new repositories package
	"sso-service/app/services"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifies JWT token from Authorization header
func AuthMiddleware(userService services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"data":    nil,
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // No "Bearer " prefix found
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"data":    nil,
				"message": "Bearer token required",
			})
			c.Abort()
			return
		}

		claims, err := userService.VerifyAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"data":    nil,
				"message": "Invalid or expired token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Set user ID and claims in context for subsequent handlers
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)

		c.Next() // Proceed to the next handler (the actual route handler)
	}
}

// SetupRouter initializes API routes
func SetupRouter(apiVersion string, jwtSecret string, db *sql.DB) *gin.Engine { // Added db parameter
	r := gin.Default() // Create Gin router instance

	// Initialize Repository
	userRepository := repositories.NewUserRepository(db) // Pass db to UserRepository

	// Initialize Service
	userService := services.NewUserService(jwtSecret, userRepository) // Pass userRepository to UserService

	// Initialize Controllers
	authController := controllers.NewAuthController(userService) // AuthController for login, refresh, logout
	userController := controllers.NewUserController(userService) // User Controller for /me (and other user profile actions)

	// Group routes for /api/<version>
	apiGroup := r.Group("/api/" + apiVersion)
	{
		// Public routes (no authentication required)
		apiGroup.POST("/login", authController.Login)

		// Protected routes (authentication required)
		// All routes defined within this 'protected' group will automatically use AuthMiddleware
		protected := apiGroup.Group("/")
		protected.Use(AuthMiddleware(userService)) // Apply AuthMiddleware to all routes in this group
		{
			protected.GET("/me", userController.GetMe)
			apiGroup.POST("/refresh", authController.Refresh)
			protected.POST("/logout", authController.Logout) // Logout is now handled by AuthController
			// Add other protected routes here that require authentication
			// e.g., protected.GET("/users/:id", userController.GetUserByID) // Example: if you want to get other user profiles
			// e.g., protected.GET("/content", contentController.GetContent)
		}
	}

	return r
}
