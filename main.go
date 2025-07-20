package main

import (
	"log"
	"os"
	"sso-service/app/controllers"
	"sso-service/app/repositories"
	"sso-service/app/services"
	"sso-service/database"
	"sso-service/router"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env. Using default environment variables.")
	}

	// Get API version from environment variable, default to "v1" if not set
	apiVersion := os.Getenv("API_VERSION")
	if apiVersion == "" {
		apiVersion = "v1"
	}

	// Get port from environment variable, default to "8080" if not set
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get JWT Secret from environment variable
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set. Please set it in .env or system environment.")
	}

	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB(db) // Ensure database connection is closed when main exits

	// Initialize Repositories
	userRepository := repositories.NewUserRepository(db)
	siteRepository := repositories.NewSiteRepository(db)
	roleRepository := repositories.NewRoleRepository(db)
	permissionRepository := repositories.NewPermissionRepository(db)
	revokedTokenRepository := repositories.NewRevokedTokenRepository(db)

	// Initialize Services
	// UserService needs other repositories for authorization helpers
	userService := services.NewUserService(jwtSecret, userRepository, roleRepository, permissionRepository, revokedTokenRepository)
	siteService := services.NewSiteService(siteRepository, userService)
	roleService := services.NewRoleService(roleRepository, userRepository, userService)                                   // RoleService needs UserRepository
	permissionService := services.NewPermissionService(permissionRepository, roleRepository, siteRepository, userService) // PermissionService needs Role/Site Repos

	// Initialize Controllers
	authController := controllers.NewAuthController(userService)
	userController := controllers.NewUserController(userService)
	siteController := controllers.NewSiteController(siteService)
	roleController := controllers.NewRoleController(roleService)
	permissionController := controllers.NewPermissionController(permissionService)

	// SetupRouter now receives initialized services and controllers directly
	r := router.SetupRouter(
		apiVersion,
		userService,
		siteController,
		roleController,
		permissionController,
		authController,
		userController,
	)

	// Running the server on the specified port
	log.Printf("Server started on :%s with API version /api/%s\n", port, apiVersion)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
