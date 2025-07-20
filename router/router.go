package router

import (
	"database/sql"
	"sso-service/app/controllers"
	"sso-service/app/middleware" // Import the new middleware package
	"sso-service/app/repositories"
	"sso-service/app/services"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes API routes
func SetupRouter(apiVersion string, jwtSecret string, db *sql.DB) *gin.Engine {
	r := gin.Default() // Create Gin router instance

	// Initialize Repository
	userRepository := repositories.NewUserRepository(db)

	// Initialize Service
	userService := services.NewUserService(jwtSecret, userRepository)

	// Initialize Controllers
	authController := controllers.NewAuthController(userService)
	userController := controllers.NewUserController(userService)

	// Group routes for /api/<version>
	apiGroup := r.Group("/api/" + apiVersion)
	{
		// Public routes (no authentication required)
		apiGroup.POST("/login", authController.Login)
		apiGroup.POST("/refresh", authController.Refresh)

		// Protected routes (authentication required)
		protected := apiGroup.Group("/")
		// Use middleware from the new package
		protected.Use(middleware.AuthMiddleware(userService)) // Apply AuthMiddleware to all routes in this group
		{
			// User Profile (authenticated, but potentially accessible by all roles)
			protected.GET("/me", userController.GetMe)
			protected.POST("/logout", authController.Logout)

			// User Management (CRUD) - Requires specific permissions
			usersGroup := protected.Group("/users")
			{
				// Create User: Requires 'user:create' permission
				usersGroup.POST("/", middleware.PermissionMiddleware(userService, "user:create"), userController.CreateUser)
				// Get All Users: Requires 'user:read_all' permission
				usersGroup.GET("/", middleware.PermissionMiddleware(userService, "user:read_all"), userController.GetUsers)
				// Get User by ID: Requires 'user:read_all' or 'user:read_own' (handled in service)
				usersGroup.GET("/:id", middleware.PermissionMiddleware(userService, "user:read_all", "user:read_own"), userController.GetUserByID)
				// Update User: Requires 'user:update_all' or 'user:update_own' (handled in service)
				usersGroup.PUT("/:id", middleware.PermissionMiddleware(userService, "user:update_all", "user:update_own"), userController.UpdateUser)
				// Delete User: Requires 'user:delete' permission
				usersGroup.DELETE("/:id", middleware.PermissionMiddleware(userService, "user:delete"), userController.DeleteUser)
			}

			// Add other protected routes here
		}
	}

	return r
}
