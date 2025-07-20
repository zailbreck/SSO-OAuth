package main

import (
	"log"
	"sso-service/app/controllers"
	"sso-service/app/repositories"
	"sso-service/app/services"
	"sso-service/config" // Impor paket config
	"sso-service/database"
	"sso-service/router"
)

func main() {
	// 1. Muat semua konfigurasi dalam satu langkah.
	cfg := config.LoadConfig()

	// 2. Inisialisasi database dengan meneruskan konfigurasi.
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB(db)

	// 3. Inisialisasi semua komponen lain dengan nilai dari struct config.
	// Repositories
	userRepository := repositories.NewUserRepository(db)
	siteRepository := repositories.NewSiteRepository(db)
	roleRepository := repositories.NewRoleRepository(db)
	permissionRepository := repositories.NewPermissionRepository(db)
	revokedTokenRepository := repositories.NewRevokedTokenRepository(db)

	// Services
	userService := services.NewUserService(cfg.JWTSecret, userRepository, roleRepository, permissionRepository, revokedTokenRepository)
	siteService := services.NewSiteService(siteRepository, userService)
	roleService := services.NewRoleService(roleRepository, userRepository, userService)
	permissionService := services.NewPermissionService(permissionRepository, roleRepository, siteRepository, userService)

	// Controllers
	authController := controllers.NewAuthController(userService)
	userController := controllers.NewUserController(userService)
	siteController := controllers.NewSiteController(siteService)
	roleController := controllers.NewRoleController(roleService)
	permissionController := controllers.NewPermissionController(permissionService)

	// SetupRouter
	r := router.SetupRouter(
		cfg.APIVersion,
		userService,
		siteController,
		roleController,
		permissionController,
		authController,
		userController,
	)

	// 4. Jalankan server dengan port dari config.
	log.Printf("Server started on :%s with API version /api/%s\n", cfg.Port, cfg.APIVersion)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
