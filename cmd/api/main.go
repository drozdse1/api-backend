package main

import (
	"log"

	"api-backend/internal/database"
	"api-backend/internal/handlers"
	"api-backend/internal/middleware"
	"api-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.AllowedOrigins))

	healthHandler := handlers.NewHealthHandler(db)
	userHandler := handlers.NewUserHandler(db)
	radarHandler := handlers.NewRadarHandler(db)

	api := router.Group("/api/v1")
	{
		api.GET("/health", healthHandler.Check)

		users := api.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.GetByID)
			users.DELETE("/:id", userHandler.Delete)
		}

		radar := api.Group("/radar")
		{
			radar.POST("/location", radarHandler.UpdateLocation)
			radar.GET("/nearby", radarHandler.GetNearbyUsers)
		}
	}

	log.Printf("Server starting on port %s in %s mode", cfg.Port, cfg.Environment)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
