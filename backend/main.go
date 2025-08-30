package main

import (
	"log"
	"os"
	"path/filepath"

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

	// Database for metadata caching (optional)
	dbPath := filepath.Join(cacheDir, "kerpic.db")
	db, err := services.NewDatabaseService(dbPath, photosDir)
	if err != nil {
		log.Printf("Warning: Failed to initialize database (running without cache): %v", err)
		// Create a nil database service for graceful degradation
		db = nil
	} else {
		defer db.Close()
	}

	// Core services
	imageService := services.NewImageService(photosDir, cacheDir)
	folderService := services.NewFolderService(photosDir)
	albumService := services.NewAlbumService(photosDir, db)

	// Initialize handlers
	folderHandler := handlers.NewFolderHandler(folderService)
	imageHandler := handlers.NewImageHandler(imageService)
	galleryHandler := handlers.NewGalleryHandler(imageService, folderService, albumService, db)

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
		// Legacy routes (for comparison tool)
		api.GET("/folders", folderHandler.ListFolders)
		api.GET("/folders/*path", folderHandler.GetFolderContents)
		api.GET("/images/:folder", imageHandler.ListImages)
		api.GET("/thumbnail/*path", imageHandler.GetThumbnail)
		api.POST("/batch", imageHandler.CreateBatch)
		api.POST("/save", imageHandler.SaveSelected)

		// New gallery routes
		api.GET("/gallery", galleryHandler.GetDashboard)
		api.GET("/photos", galleryHandler.GetPhotos)
		api.GET("/search", galleryHandler.SearchPhotos)
		api.GET("/timeline", galleryHandler.GetTimeline)

		// Photo deletion routes
		photos := api.Group("/photos")
		{
			photos.POST("/mark-deletion", galleryHandler.MarkPhotosForDeletion)
			photos.POST("/unmark-deletion", galleryHandler.UnmarkPhotosForDeletion)
			photos.DELETE("/delete-marked", galleryHandler.DeleteMarkedPhotos)
			photos.GET("/marked", galleryHandler.GetMarkedPhotos)
		}

		// Album routes
		albums := api.Group("/albums")
		{
			albums.GET("", galleryHandler.ListAlbums)
			albums.POST("", galleryHandler.CreateAlbum)
			albums.GET("/:id", galleryHandler.GetAlbum)
			albums.PUT("/:id", galleryHandler.UpdateAlbum)
			albums.DELETE("/:id", galleryHandler.DeleteAlbum)
			albums.GET("/:id/photos", galleryHandler.GetAlbumPhotos)
			albums.POST("/:id/photos", galleryHandler.AddPhotoToAlbum)
			albums.DELETE("/:id/photos", galleryHandler.RemovePhotoFromAlbum)
		}
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