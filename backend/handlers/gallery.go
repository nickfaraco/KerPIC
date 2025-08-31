package handlers

import (
	"net/http"
	"strconv"

	"kerpic-backend/models"
	"kerpic-backend/services"

	"github.com/gin-gonic/gin"
)

type GalleryHandler struct {
	imageService  *services.ImageService
	folderService *services.FolderService
	albumService  *services.AlbumService
	db           *services.DatabaseService
}

func NewGalleryHandler(imageService *services.ImageService, folderService *services.FolderService, albumService *services.AlbumService, db *services.DatabaseService) *GalleryHandler {
	return &GalleryHandler{
		imageService:  imageService,
		folderService: folderService,
		albumService:  albumService,
		db:           db,
	}
}

// GET /api/gallery - Get dashboard data
func (gh *GalleryHandler) GetDashboard(c *gin.Context) {
	// Get statistics
	var stats *models.GalleryStats
	var err error
	
	if gh.db != nil {
		stats, err = gh.db.GetGalleryStats()
	}
	
	if gh.db == nil || err != nil {
		stats = &models.GalleryStats{
			TotalPhotos: 0,
			TotalAlbums: 0,
			TotalSize:   "0 MB",
		}
	}

	// Get recent photos (limit 20)
	var recentPhotos []models.PhotoMetadata = []models.PhotoMetadata{}
	
	// Always scan filesystem to catch new photos and cache them
	if folderPhotos, err := gh.scanFilesystemForPhotos(20); err == nil && folderPhotos != nil {
		recentPhotos = folderPhotos
	}
	
	// If filesystem scan failed, fallback to database
	if len(recentPhotos) == 0 && gh.db != nil {
		searchReq := &models.SearchRequest{
			Limit: 20,
		}
		if dbPhotos, err := gh.db.SearchPhotos(searchReq); err == nil && dbPhotos != nil {
			recentPhotos = dbPhotos
		}
	}

	// Add thumbnail URLs to photos
	for i := range recentPhotos {
		recentPhotos[i].ThumbnailURL = "/api/thumbnail/" + recentPhotos[i].Path
	}

	// Get recent albums (limit 6)
	recentAlbums, err := gh.albumService.ListAlbums()
	if err != nil {
		recentAlbums = []models.Album{}
	}
	
	// Limit to 6 most recent
	if len(recentAlbums) > 6 {
		recentAlbums = recentAlbums[:6]
	}

	dashboard := &models.GalleryDashboard{
		Stats:        *stats,
		RecentPhotos: recentPhotos,
		RecentAlbums: recentAlbums,
	}

	c.JSON(http.StatusOK, dashboard)
}

// GET /api/photos - Get photos with pagination and filtering
func (gh *GalleryHandler) GetPhotos(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 200 {
		limit = 50
	}

	searchReq := &models.SearchRequest{
		Limit:  limit,
		Offset: offset,
	}

	// Parse date filters if provided
	if dateFrom := c.Query("dateFrom"); dateFrom != "" {
		// TODO: Parse date string
	}
	
	if dateTo := c.Query("dateTo"); dateTo != "" {
		// TODO: Parse date string  
	}

	// Parse tag filters
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		searchReq.Tags = tags
	}

	var photos []models.PhotoMetadata = []models.PhotoMetadata{}
	
	// Always scan filesystem to catch new photos and cache them
	if folderPhotos, err := gh.scanFilesystemForPhotos(limit); err == nil && folderPhotos != nil {
		photos = folderPhotos
	}
	
	// If filesystem scan failed, fallback to database
	if len(photos) == 0 && gh.db != nil {
		if dbPhotos, err := gh.db.SearchPhotos(searchReq); err == nil && dbPhotos != nil {
			photos = dbPhotos
		}
	}

	// Add thumbnail URLs
	for i := range photos {
		photos[i].ThumbnailURL = "/api/thumbnail/" + photos[i].Path
	}

	c.JSON(http.StatusOK, gin.H{
		"photos": photos,
		"total":  len(photos), // TODO: Get actual total count
		"limit":  limit,
		"offset": offset,
	})
}

// Album endpoints

// GET /api/albums - List all albums
func (gh *GalleryHandler) ListAlbums(c *gin.Context) {
	albums, err := gh.albumService.ListAlbums()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list albums"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"albums": albums})
}

// POST /api/albums - Create new album
func (gh *GalleryHandler) CreateAlbum(c *gin.Context) {
	var req models.AlbumCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	album, err := gh.albumService.CreateAlbum(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create album"})
		return
	}

	c.JSON(http.StatusCreated, album)
}

// GET /api/albums/:id - Get album by ID
func (gh *GalleryHandler) GetAlbum(c *gin.Context) {
	albumID := c.Param("id")
	
	album, err := gh.albumService.GetAlbum(albumID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	c.JSON(http.StatusOK, album)
}

// PUT /api/albums/:id - Update album
func (gh *GalleryHandler) UpdateAlbum(c *gin.Context) {
	albumID := c.Param("id")
	
	var req models.AlbumUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	album, err := gh.albumService.UpdateAlbum(albumID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update album"})
		return
	}

	c.JSON(http.StatusOK, album)
}

// DELETE /api/albums/:id - Delete album
func (gh *GalleryHandler) DeleteAlbum(c *gin.Context) {
	albumID := c.Param("id")
	
	err := gh.albumService.DeleteAlbum(albumID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete album"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Album deleted successfully"})
}

// GET /api/albums/:id/photos - Get photos in album
func (gh *GalleryHandler) GetAlbumPhotos(c *gin.Context) {
	albumID := c.Param("id")
	
	photos, err := gh.albumService.GetAlbumPhotos(albumID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get album photos"})
		return
	}

	// Add thumbnail URLs
	for i := range photos {
		photos[i].ThumbnailURL = "/api/thumbnail/" + photos[i].Path
	}

	c.JSON(http.StatusOK, gin.H{"photos": photos})
}

// POST /api/albums/:id/photos - Add photo to album
func (gh *GalleryHandler) AddPhotoToAlbum(c *gin.Context) {
	albumID := c.Param("id")
	
	var req struct {
		PhotoPath string `json:"photoPath" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := gh.albumService.AddPhotoToAlbum(albumID, req.PhotoPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add photo to album"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Photo added to album"})
}

// DELETE /api/albums/:id/photos - Remove photo from album
func (gh *GalleryHandler) RemovePhotoFromAlbum(c *gin.Context) {
	albumID := c.Param("id")
	photoPath := c.Query("photoPath")
	
	if photoPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "photoPath query parameter is required"})
		return
	}

	err := gh.albumService.RemovePhotoFromAlbum(albumID, photoPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove photo from album"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Photo removed from album"})
}

// Search endpoints

// GET /api/search - Search photos
func (gh *GalleryHandler) SearchPhotos(c *gin.Context) {
	var req models.SearchRequest
	
	// Parse query parameters
	req.Query = c.Query("q")
	req.Tags = c.QueryArray("tags")
	req.Location = c.Query("location")
	req.Albums = c.QueryArray("albums")
	
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	
	req.Limit, _ = strconv.Atoi(limitStr)
	req.Offset, _ = strconv.Atoi(offsetStr)

	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}

	// TODO: Parse date range parameters
	
	var photos []models.PhotoMetadata = []models.PhotoMetadata{}
	
	// Always scan filesystem to catch new photos and cache them
	if folderPhotos, err := gh.scanFilesystemForPhotos(req.Limit); err == nil && folderPhotos != nil {
		photos = folderPhotos
	}
	
	// If filesystem scan failed, fallback to database
	if len(photos) == 0 && gh.db != nil {
		if dbPhotos, err := gh.db.SearchPhotos(&req); err == nil && dbPhotos != nil {
			photos = dbPhotos
		}
	}

	// Add thumbnail URLs
	for i := range photos {
		photos[i].ThumbnailURL = "/api/thumbnail/" + photos[i].Path
	}

	c.JSON(http.StatusOK, gin.H{
		"photos": photos,
		"total":  len(photos), // TODO: Get actual total count
		"query":  req,
	})
}

// Timeline endpoints

// GET /api/timeline - Get photos grouped by time
func (gh *GalleryHandler) GetTimeline(c *gin.Context) {
	groupBy := c.DefaultQuery("groupBy", "month") // month, day, year
	year := c.Query("year")
	
	// TODO: Implement timeline grouping logic
	// For now, return basic structure
	
	var groups []models.TimelineGroup
	
	// Placeholder implementation
	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"groupBy": groupBy,
		"year": year,
	})
}

// Helper method to scan filesystem for photos when database is unavailable
func (gh *GalleryHandler) scanFilesystemForPhotos(limit int) ([]models.PhotoMetadata, error) {
	folderContents, err := gh.folderService.GetFolderContents("")
	if err != nil {
		return nil, err
	}

	var photos []models.PhotoMetadata
	count := 0

	// Convert legacy ImageInfo to PhotoMetadata
	for _, img := range folderContents.Images {
		if count >= limit {
			break
		}

		// Try to get existing metadata from database first
		var photo models.PhotoMetadata
		var fromDB bool = false
		
		if gh.db != nil {
			if existingPhoto, err := gh.db.GetCachedPhotoMetadata(img.Path); err == nil {
				photo = *existingPhoto
				fromDB = true
			}
		}
		
		// If not in database or database unavailable, create fresh metadata
		if !fromDB {
			photo = models.PhotoMetadata{
				Path:    img.Path,
				Name:    img.Name,
				Size:    img.Size,
				ModTime: img.ModTime,
				Width:   img.Width,
				Height:  img.Height,
				Orientation: img.Orientation,
			}
			
			// Cache photo metadata in database for future operations (like deletion)
			if gh.db != nil {
				gh.db.CachePhotoMetadata(&photo)
			}
		}
		
		photos = append(photos, photo)
		count++
	}

	// Also scan subfolders for more photos
	for _, subfolder := range folderContents.Subfolders {
		if count >= limit {
			break
		}

		subContents, err := gh.folderService.GetFolderContents(subfolder.Path)
		if err != nil {
			continue
		}

		for _, img := range subContents.Images {
			if count >= limit {
				break
			}

			// Try to get existing metadata from database first
			var photo models.PhotoMetadata
			var fromDB bool = false
			
			if gh.db != nil {
				if existingPhoto, err := gh.db.GetCachedPhotoMetadata(img.Path); err == nil {
					photo = *existingPhoto
					fromDB = true
				}
			}
			
			// If not in database or database unavailable, create fresh metadata
			if !fromDB {
				photo = models.PhotoMetadata{
					Path:    img.Path,
					Name:    img.Name,
					Size:    img.Size,
					ModTime: img.ModTime,
					Width:   img.Width,
					Height:  img.Height,
					Orientation: img.Orientation,
				}
				
				// Cache photo metadata in database for future operations (like deletion)
				if gh.db != nil {
					gh.db.CachePhotoMetadata(&photo)
				}
			}
			
			photos = append(photos, photo)
			count++
		}
	}

	return photos, nil
}

// Photo deletion endpoints

// POST /api/photos/mark-deletion - Mark photos for deletion
func (gh *GalleryHandler) MarkPhotosForDeletion(c *gin.Context) {
	var req struct {
		PhotoPaths []string `json:"photoPaths" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.PhotoPaths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No photo paths provided"})
		return
	}

	if gh.db != nil {
		if err := gh.db.MarkPhotosForDeletion(req.PhotoPaths); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark photos for deletion"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Photos marked for deletion successfully", 
		"count": len(req.PhotoPaths),
	})
}

// POST /api/photos/unmark-deletion - Unmark photos for deletion
func (gh *GalleryHandler) UnmarkPhotosForDeletion(c *gin.Context) {
	var req struct {
		PhotoPaths []string `json:"photoPaths" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.PhotoPaths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No photo paths provided"})
		return
	}

	if gh.db != nil {
		if err := gh.db.UnmarkPhotosForDeletion(req.PhotoPaths); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmark photos for deletion"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Photos unmarked for deletion successfully",
		"count": len(req.PhotoPaths),
	})
}

// DELETE /api/photos/delete-marked - Permanently delete marked photos
func (gh *GalleryHandler) DeleteMarkedPhotos(c *gin.Context) {
	if gh.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	deletedPhotos, err := gh.db.DeleteMarkedPhotos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete marked photos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Marked photos deleted successfully",
		"deletedPhotos": deletedPhotos,
		"count": len(deletedPhotos),
	})
}

// GET /api/photos/marked - Get list of photos marked for deletion
func (gh *GalleryHandler) GetMarkedPhotos(c *gin.Context) {
	if gh.db == nil {
		c.JSON(http.StatusOK, gin.H{"markedPhotos": []string{}})
		return
	}

	markedPhotos, err := gh.db.GetMarkedPhotos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get marked photos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"markedPhotos": markedPhotos,
		"count": len(markedPhotos),
	})
}