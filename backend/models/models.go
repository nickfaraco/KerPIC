package models

import "time"

// GPSInfo represents GPS coordinates from EXIF data
type GPSInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude,omitempty"`
}

// PhotoMetadata represents enhanced metadata about a photo file
type PhotoMetadata struct {
	Path              string    `json:"path"`         // Actual file location
	Name              string    `json:"name"`
	Size              int64     `json:"size"`
	ModTime           time.Time `json:"modTime"`
	DateTaken         time.Time `json:"dateTaken"`    // From EXIF
	Location          *GPSInfo  `json:"location"`     // From EXIF GPS
	Width             int       `json:"width"`
	Height            int       `json:"height"`
	Orientation       int       `json:"orientation"`
	Rating            int       `json:"rating"`       // Stored in .json sidecar
	Tags              []string  `json:"tags"`         // Stored in .json sidecar
	Albums            []string  `json:"albums"`       // Which album folders link to this
	ThumbnailURL      string    `json:"thumbnailUrl"`
	MarkedForDeletion bool      `json:"markedForDeletion"`
}

// ImageInfo represents metadata about an image file (legacy, for backward compatibility)
type ImageInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"modTime"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Orientation  int       `json:"orientation"`
	ThumbnailURL string    `json:"thumbnailUrl"`
}

// FolderInfo represents a folder with its contents
type FolderInfo struct {
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Images    []ImageInfo  `json:"images"`
	Subfolders []FolderInfo `json:"subfolders"`
}

// BatchRequest represents a request to create a comparison batch
type BatchRequest struct {
	ImagePaths []string `json:"imagePaths" binding:"required"`
}

// BatchResponse represents the created batch
type BatchResponse struct {
	ID     string      `json:"id"`
	Images []ImageInfo `json:"images"`
}

// SaveRequest represents a request to save selected images
type SaveRequest struct {
	BatchID       string   `json:"batchId" binding:"required"`
	SelectedPaths []string `json:"selectedPaths" binding:"required"`
	TargetFolder  string   `json:"targetFolder"`
}

// SaveResponse represents the result of saving images
type SaveResponse struct {
	Success       []string `json:"success"`
	Failed        []string `json:"failed"`
	Conflicts     []string `json:"conflicts"`
	TargetFolder  string   `json:"targetFolder"`
}

// Album represents a photo album
type Album struct {
	ID          string    `json:"id"`
	Path        string    `json:"path"`         // Filesystem folder path
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CoverPhoto  string    `json:"coverPhoto"`
	PhotoCount  int       `json:"photoCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Tags        []string  `json:"tags"`
}

// AlbumCreateRequest represents a request to create a new album
type AlbumCreateRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	PhotoPaths  []string `json:"photoPaths"`
	Tags        []string `json:"tags"`
}

// AlbumUpdateRequest represents a request to update an album
type AlbumUpdateRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CoverPhoto  string   `json:"coverPhoto"`
	Tags        []string `json:"tags"`
}

// GalleryDashboard represents dashboard data
type GalleryDashboard struct {
	Stats        GalleryStats    `json:"stats"`
	RecentPhotos []PhotoMetadata `json:"recentPhotos"`
	RecentAlbums []Album         `json:"recentAlbums"`
}

// GalleryStats represents gallery statistics
type GalleryStats struct {
	TotalPhotos int    `json:"totalPhotos"`
	TotalAlbums int    `json:"totalAlbums"`
	TotalSize   string `json:"totalSize"`
}

// TimelineGroup represents photos grouped by time period
type TimelineGroup struct {
	Date        time.Time       `json:"date"`
	Title       string          `json:"title"`         // e.g., "August 2025", "August 12, 2025"
	PhotoCount  int             `json:"photoCount"`
	Photos      []PhotoMetadata `json:"photos"`
}

// SearchRequest represents a photo search request
type SearchRequest struct {
	Query       string    `json:"query"`
	Tags        []string  `json:"tags"`
	DateFrom    time.Time `json:"dateFrom"`
	DateTo      time.Time `json:"dateTo"`
	Location    string    `json:"location"`
	Albums      []string  `json:"albums"`
	Limit       int       `json:"limit"`
	Offset      int       `json:"offset"`
}