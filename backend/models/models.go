package models

import "time"

// ImageInfo represents metadata about an image file
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