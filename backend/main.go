package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"kerpic-backend/handlers"
	"kerpic-backend/services"
)

func main() {
	// Initialize services
	photosDir := os.Getenv("PHOTOS_DIR")
	if photosDir == "" {
		photosDir = "/app/data/photos"
	}

	cacheDir := os.Getenv("CACHE_DIR")
	if cacheDir == "" {
		cacheDir = "/app/cache"
	}

	imageService := services.NewImageService(photosDir, cacheDir)
	folderService := services.NewFolderService(photosDir)

	// Initialize handlers
	folderHandler := handlers.NewFolderHandler(folderService)
	imageHandler := handlers.NewImageHandler(imageService)

	// Setup router
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Serve static files (SvelteKit build)
	r.Static("/_app", "./static/_app")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")
	r.StaticFile("/", "./static/index.html")
	
	// Fallback for SPA routing
	r.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// API routes
	api := r.Group("/api")
	{
		api.GET("/folders", folderHandler.ListFolders)
		api.GET("/folders/*path", folderHandler.GetFolderContents)
		api.GET("/images/:folder", imageHandler.ListImages)
		api.GET("/thumbnail/*path", imageHandler.GetThumbnail)
		api.POST("/batch", imageHandler.CreateBatch)
		api.POST("/save", imageHandler.SaveSelected)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Starting KerPIC server on port %s", port)
	log.Printf("Photos directory: %s", photosDir)
	log.Printf("Cache directory: %s", cacheDir)

	r.Run(":" + port)
}