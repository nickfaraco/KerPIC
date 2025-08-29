package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kerpic-backend/models"

	"github.com/google/uuid"
)

type AlbumService struct {
	baseDir string
	albumsDir string
	db      *DatabaseService
}

func NewAlbumService(baseDir string, db *DatabaseService) *AlbumService {
	albumsDir := filepath.Join(baseDir, "albums")
	os.MkdirAll(albumsDir, 0755)
	
	return &AlbumService{
		baseDir:   baseDir,
		albumsDir: albumsDir,
		db:        db, // Can be nil
	}
}

// Album metadata file operations
func (as *AlbumService) getMetadataPath(albumPath string) string {
	return filepath.Join(albumPath, ".album-metadata.json")
}

func (as *AlbumService) saveMetadata(album *models.Album) error {
	metadataPath := as.getMetadataPath(album.Path)
	
	metadata := struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		CoverPhoto  string    `json:"coverPhoto"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
		Tags        []string  `json:"tags"`
	}{
		Name:        album.Name,
		Description: album.Description,
		CoverPhoto:  album.CoverPhoto,
		CreatedAt:   album.CreatedAt,
		UpdatedAt:   album.UpdatedAt,
		Tags:        album.Tags,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

func (as *AlbumService) loadMetadata(albumPath string) (*models.Album, error) {
	metadataPath := as.getMetadataPath(albumPath)
	
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		CoverPhoto  string    `json:"coverPhoto"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
		Tags        []string  `json:"tags"`
	}

	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	// Count photos in album directory
	photoCount, err := as.countPhotosInDirectory(albumPath)
	if err != nil {
		photoCount = 0
	}

	album := &models.Album{
		ID:          filepath.Base(albumPath),
		Path:        albumPath,
		Name:        metadata.Name,
		Description: metadata.Description,
		CoverPhoto:  metadata.CoverPhoto,
		PhotoCount:  photoCount,
		CreatedAt:   metadata.CreatedAt,
		UpdatedAt:   metadata.UpdatedAt,
		Tags:        metadata.Tags,
	}

	return album, nil
}

func (as *AlbumService) countPhotosInDirectory(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && isImageFile(entry.Name()) {
			count++
		}
	}

	return count, nil
}

// Create a new album
func (as *AlbumService) CreateAlbum(req *models.AlbumCreateRequest) (*models.Album, error) {
	// Generate unique ID and create directory
	albumID := uuid.New().String()
	albumPath := filepath.Join(as.albumsDir, albumID)
	
	if err := os.MkdirAll(albumPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create album directory: %w", err)
	}

	now := time.Now()
	album := &models.Album{
		ID:          albumID,
		Path:        albumPath,
		Name:        req.Name,
		Description: req.Description,
		PhotoCount:  0,
		CreatedAt:   now,
		UpdatedAt:   now,
		Tags:        req.Tags,
	}

	// Save metadata to filesystem
	if err := as.saveMetadata(album); err != nil {
		os.RemoveAll(albumPath) // Cleanup on error
		return nil, fmt.Errorf("failed to save album metadata: %w", err)
	}

	// Cache in database
	if as.db != nil {
		if err := as.db.CreateAlbum(album); err != nil {
			// Don't fail if database caching fails, filesystem is source of truth
			fmt.Printf("Warning: failed to cache album in database: %v\n", err)
		}
	}

	// Add photos if provided
	if len(req.PhotoPaths) > 0 {
		for _, photoPath := range req.PhotoPaths {
			if err := as.AddPhotoToAlbum(albumID, photoPath); err != nil {
				fmt.Printf("Warning: failed to add photo %s to album: %v\n", photoPath, err)
			}
		}
		
		// Update photo count and set cover if not set
		album.PhotoCount = len(req.PhotoPaths)
		if album.CoverPhoto == "" && len(req.PhotoPaths) > 0 {
			album.CoverPhoto = req.PhotoPaths[0]
		}
		
		album.UpdatedAt = time.Now()
		as.saveMetadata(album)
		if as.db != nil {
			as.db.CreateAlbum(album) // Update cache
		}
	}

	return album, nil
}

// Add a photo to an album via symlink
func (as *AlbumService) AddPhotoToAlbum(albumID, photoPath string) error {
	albumPath := filepath.Join(as.albumsDir, albumID)
	
	// Check if album exists
	if _, err := os.Stat(albumPath); os.IsNotExist(err) {
		return fmt.Errorf("album not found")
	}

	// Check if photo exists
	if _, err := os.Stat(photoPath); os.IsNotExist(err) {
		return fmt.Errorf("photo not found: %s", photoPath)
	}

	// Create symlink in album directory
	photoName := filepath.Base(photoPath)
	linkPath := filepath.Join(albumPath, photoName)
	
	// Handle naming conflicts
	counter := 1
	originalLinkPath := linkPath
	for {
		if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
			break // Path is available
		}
		
		// Check if existing symlink points to the same file
		if target, err := os.Readlink(linkPath); err == nil {
			if filepath.Clean(target) == filepath.Clean(photoPath) {
				// Already linked
				as.db.AddPhotoToAlbum(albumID, photoPath)
				return nil
			}
		}
		
		// Generate new name
		ext := filepath.Ext(originalLinkPath)
		base := strings.TrimSuffix(filepath.Base(originalLinkPath), ext)
		linkPath = filepath.Join(albumPath, fmt.Sprintf("%s_%d%s", base, counter, ext))
		counter++
	}

	// Create relative symlink
	relPath, err := filepath.Rel(albumPath, photoPath)
	if err != nil {
		return fmt.Errorf("failed to create relative path: %w", err)
	}

	if err := os.Symlink(relPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Update database cache
	if as.db != nil {
		as.db.AddPhotoToAlbum(albumID, photoPath)
	}

	return nil
}

// Remove photo from album
func (as *AlbumService) RemovePhotoFromAlbum(albumID, photoPath string) error {
	albumPath := filepath.Join(as.albumsDir, albumID)
	
	// Find and remove symlink
	entries, err := os.ReadDir(albumPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == ".album-metadata.json" {
			continue
		}

		linkPath := filepath.Join(albumPath, entry.Name())
		if target, err := os.Readlink(linkPath); err == nil {
			// Resolve to absolute path for comparison
			if absTarget, err := filepath.Abs(filepath.Join(albumPath, target)); err == nil {
				if filepath.Clean(absTarget) == filepath.Clean(photoPath) {
					os.Remove(linkPath)
					if as.db != nil {
						as.db.RemovePhotoFromAlbum(albumID, photoPath)
					}
					return nil
				}
			}
		}
	}

	return fmt.Errorf("photo not found in album")
}

// List all albums
func (as *AlbumService) ListAlbums() ([]models.Album, error) {
	var albums []models.Album

	// First try database cache
	if as.db != nil {
		if cachedAlbums, err := as.db.ListAlbums(100, 0); err == nil && len(cachedAlbums) > 0 {
			return cachedAlbums, nil
		}
	}

	// Fallback to filesystem scan
	entries, err := os.ReadDir(as.albumsDir)
	if err != nil {
		return albums, nil // Return empty list if albums dir doesn't exist
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		albumPath := filepath.Join(as.albumsDir, entry.Name())
		if album, err := as.loadMetadata(albumPath); err == nil {
			albums = append(albums, *album)
			// Re-cache in database
			if as.db != nil {
				as.db.CreateAlbum(album)
			}
		}
	}

	return albums, nil
}

// Get album by ID
func (as *AlbumService) GetAlbum(albumID string) (*models.Album, error) {
	// Try database cache first
	if as.db != nil {
		if album, err := as.db.GetAlbum(albumID); err == nil {
			return album, nil
		}
	}

	// Fallback to filesystem
	albumPath := filepath.Join(as.albumsDir, albumID)
	album, err := as.loadMetadata(albumPath)
	if err != nil {
		return nil, fmt.Errorf("album not found")
	}

	// Re-cache in database
	if as.db != nil {
		as.db.CreateAlbum(album)
	}

	return album, nil
}

// Get photos in an album
func (as *AlbumService) GetAlbumPhotos(albumID string) ([]models.PhotoMetadata, error) {
	albumPath := filepath.Join(as.albumsDir, albumID)
	
	entries, err := os.ReadDir(albumPath)
	if err != nil {
		return nil, err
	}

	var photos []models.PhotoMetadata
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == ".album-metadata.json" {
			continue
		}

		linkPath := filepath.Join(albumPath, entry.Name())
		
		// Resolve symlink to actual photo path
		var actualPath string
		if target, err := os.Readlink(linkPath); err == nil {
			if filepath.IsAbs(target) {
				actualPath = target
			} else {
				actualPath = filepath.Join(albumPath, target)
			}
		} else {
			// Not a symlink, use as-is
			actualPath = linkPath
		}

		// Clean path
		actualPath = filepath.Clean(actualPath)

		// Check if cached metadata exists
		var foundCached bool
		if as.db != nil {
			if metadata, err := as.db.GetCachedPhotoMetadata(actualPath); err == nil {
				photos = append(photos, *metadata)
				foundCached = true
			}
		}
		
		if !foundCached {
			// Create basic metadata if not cached
			if stat, err := os.Stat(actualPath); err == nil && isImageFile(entry.Name()) {
				photo := models.PhotoMetadata{
					Path:    actualPath,
					Name:    entry.Name(),
					Size:    stat.Size(),
					ModTime: stat.ModTime(),
				}
				photos = append(photos, photo)
			}
		}
	}

	return photos, nil
}

// Update album metadata
func (as *AlbumService) UpdateAlbum(albumID string, req *models.AlbumUpdateRequest) (*models.Album, error) {
	album, err := as.GetAlbum(albumID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		album.Name = req.Name
	}
	if req.Description != "" {
		album.Description = req.Description
	}
	if req.CoverPhoto != "" {
		album.CoverPhoto = req.CoverPhoto
	}
	if req.Tags != nil {
		album.Tags = req.Tags
	}

	album.UpdatedAt = time.Now()

	// Save to filesystem
	if err := as.saveMetadata(album); err != nil {
		return nil, err
	}

	// Update database cache
	if as.db != nil {
		as.db.CreateAlbum(album)
	}

	return album, nil
}

// Delete album
func (as *AlbumService) DeleteAlbum(albumID string) error {
	albumPath := filepath.Join(as.albumsDir, albumID)
	
	// Remove from filesystem (this removes symlinks, not actual photos)
	if err := os.RemoveAll(albumPath); err != nil {
		return fmt.Errorf("failed to delete album directory: %w", err)
	}

	// Remove from database cache
	// Note: CASCADE will handle album_photos cleanup
	if as.db != nil {
		as.db.db.Exec("DELETE FROM albums WHERE id = ?", albumID)
	}

	return nil
}

// Helper function to check if file is an image
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" || ext == ".heic"
}