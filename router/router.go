package router

import (
	// "database/sql" // Removed: db is no longer directly passed here
	"sso-service/app/controllers"
	"sso-service/app/middleware"
	"sso-service/app/services"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes API routes
func SetupRouter(
	apiVersion string,
	userService services.UserService, // Passed for middleware
	siteController controllers.SiteController,
	roleController controllers.RoleController,
	permissionController controllers.PermissionController,
	authController controllers.AuthController,
	userController controllers.UserController,
) *gin.Engine {
	r := gin.Default() // Create Gin router instance

	// Group routes for /api/<version>
	apiGroup := r.Group("/api/" + apiVersion)
	{
		// Public routes (no authentication required)
		apiGroup.POST("/login", authController.Login)

		// Protected routes (authentication required)
		protected := apiGroup.Group("/")
		protected.Use(middleware.AuthMiddleware(userService)) // Apply AuthMiddleware to all routes in this group
		{
			// User Profile (authenticated, but potentially accessible by all roles)
			protected.GET("/me", userController.GetMe)
			protected.POST("/refresh", authController.Refresh)
			protected.POST("/logout", authController.Logout)

			// User Management (CRUD) - Requires specific permissions
			usersGroup := protected.Group("/users")
			{
				usersGroup.POST("/", middleware.PermissionMiddleware(userService, "user:create"), userController.CreateUser)
				usersGroup.GET("/", middleware.PermissionMiddleware(userService, "user:read_all"), userController.GetUsers)
				usersGroup.GET("/:id", middleware.PermissionMiddleware(userService, "user:read_all", "user:read_own"), userController.GetUserByID)
				usersGroup.PUT("/:id", middleware.PermissionMiddleware(userService, "user:update_all", "user:update_own"), userController.UpdateUser)
				usersGroup.DELETE("/:id", middleware.PermissionMiddleware(userService, "user:delete"), userController.DeleteUser)
			}

			// Site Management (CRUD) - Requires specific permissions
			sitesGroup := protected.Group("/sites")
			{
				sitesGroup.POST("/", middleware.PermissionMiddleware(userService, "site:create"), siteController.CreateSite)
				sitesGroup.GET("/", middleware.PermissionMiddleware(userService, "site:read_all"), siteController.GetAllSites)
				sitesGroup.GET("/:id", middleware.PermissionMiddleware(userService, "site:read_all"), siteController.GetSiteByID)
				sitesGroup.PUT("/:id", middleware.PermissionMiddleware(userService, "site:update"), siteController.UpdateSite)
				sitesGroup.DELETE("/:id", middleware.PermissionMiddleware(userService, "site:delete"), siteController.DeleteSite)
			}

			// Role Management (CRUD & Assignment) - Requires specific permissions
			rolesGroup := protected.Group("/roles")
			{
				rolesGroup.POST("/", middleware.PermissionMiddleware(userService, "role:create"), roleController.CreateRole)
				rolesGroup.GET("/", middleware.PermissionMiddleware(userService, "role:read_all"), roleController.GetAllRoles)
				rolesGroup.GET("/:id", middleware.PermissionMiddleware(userService, "role:read_all"), roleController.GetRoleByID)
				rolesGroup.PUT("/:id", middleware.PermissionMiddleware(userService, "role:update"), roleController.UpdateRole)
				rolesGroup.DELETE("/:id", middleware.PermissionMiddleware(userService, "role:delete"), roleController.DeleteRole)
				rolesGroup.POST("/assign", middleware.PermissionMiddleware(userService, "role:assign"), roleController.AssignRoleToUser)
				rolesGroup.POST("/remove", middleware.PermissionMiddleware(userService, "role:assign"), roleController.RemoveRoleFromUser) // Re-use assign permission for remove
			}

			// Permission Management (CRUD & Assignment) - Requires specific permissions
			permissionsGroup := protected.Group("/permissions")
			{
				permissionsGroup.POST("/", middleware.PermissionMiddleware(userService, "permission:create"), permissionController.CreatePermission)
				permissionsGroup.GET("/", middleware.PermissionMiddleware(userService, "permission:read_all"), permissionController.GetAllPermissions)
				permissionsGroup.GET("/:id", middleware.PermissionMiddleware(userService, "permission:read_all"), permissionController.GetPermissionByID)
				permissionsGroup.PUT("/:id", middleware.PermissionMiddleware(userService, "permission:update"), permissionController.UpdatePermission)
				permissionsGroup.DELETE("/:id", middleware.PermissionMiddleware(userService, "permission:delete"), permissionController.DeletePermission)
				permissionsGroup.POST("/assign", middleware.PermissionMiddleware(userService, "permission:assign"), permissionController.AssignPermissionToRole)
				permissionsGroup.POST("/remove", middleware.PermissionMiddleware(userService, "permission:assign"), permissionController.RemovePermissionFromRole) // Re-use assign permission for remove
			}
		}
	}

	return r
}
