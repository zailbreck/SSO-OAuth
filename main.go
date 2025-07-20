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
	// Impor GORM
)

func main() {
	// ... (kode untuk env, port, jwtSecret tetap sama) ...
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env. Using default environment variables.")
	}

	apiVersion := os.Getenv("API_VERSION")
	if apiVersion == "" {
		apiVersion = "v1"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set. Please set it in .env or system environment.")
	}

	// PERUBAHAN DI SINI: Inisialisasi database GORM
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB(db)

	userRepository := repositories.NewUserRepository(db)
	siteRepository := repositories.NewSiteRepository(db)
	roleRepository := repositories.NewRoleRepository(db)
	permissionRepository := repositories.NewPermissionRepository(db)
	revokedTokenRepository := repositories.NewRevokedTokenRepository(db) // Masih menggunakan sql.DB

	// ... (sisa kode untuk inisialisasi services, controllers, dan router tetap sama) ...
	// Initialize Services
	userService := services.NewUserService(jwtSecret, userRepository, roleRepository, permissionRepository, revokedTokenRepository)
	siteService := services.NewSiteService(siteRepository, userService)
	roleService := services.NewRoleService(roleRepository, userRepository, userService)
	permissionService := services.NewPermissionService(permissionRepository, roleRepository, siteRepository, userService)

	// Initialize Controllers
	authController := controllers.NewAuthController(userService)
	userController := controllers.NewUserController(userService)
	siteController := controllers.NewSiteController(siteService)
	roleController := controllers.NewRoleController(roleService)
	permissionController := controllers.NewPermissionController(permissionService)

	// SetupRouter
	r := router.SetupRouter(
		apiVersion,
		userService,
		siteController,
		roleController,
		permissionController,
		authController,
		userController,
	)

	// Running the server
	log.Printf("Server started on :%s with API version /api/%s\n", port, apiVersion)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
