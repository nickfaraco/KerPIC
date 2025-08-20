package services

import (
	"os"
	"path/filepath"
	"strings"

	"kerpic-backend/models"
)

type FolderService struct {
	baseDir string
}

func NewFolderService(baseDir string) *FolderService {
	return &FolderService{
		baseDir: baseDir,
	}
}

// ListFolders returns the root folders available
func (fs *FolderService) ListFolders() ([]models.FolderInfo, error) {
	// Get the root directory contents
	rootContents, err := fs.GetFolderContents("")
	if err != nil {
		return nil, err
	}

	// Return just the root contents (includes both subfolders and images)
	return []models.FolderInfo{*rootContents}, nil
}

// GetFolderContents returns the contents of a specific folder
func (fs *FolderService) GetFolderContents(relativePath string) (*models.FolderInfo, error) {
	// Handle empty path (root directory)
	var cleanPath, fullPath string
	if relativePath == "" {
		cleanPath = ""
		fullPath = fs.baseDir
	} else {
		// Sanitize path to prevent directory traversal
		cleanPath = filepath.Clean(relativePath)
		if strings.Contains(cleanPath, "..") {
			return nil, os.ErrPermission
		}
		fullPath = filepath.Join(fs.baseDir, cleanPath)
	}

	// Check if path exists and is a directory
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrInvalid
	}

	// Handle folder name for root directory
	folderName := filepath.Base(cleanPath)
	if cleanPath == "" {
		folderName = "Root"
	}

	folder := &models.FolderInfo{
		Name: folderName,
		Path: cleanPath,
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			folder.Subfolders = append(folder.Subfolders, models.FolderInfo{
				Name: entry.Name(),
				Path: filepath.Join(cleanPath, entry.Name()),
			})
		} else if fs.isImageFile(entry.Name()) {
			// Basic image info, detailed info will be loaded on demand
			folder.Images = append(folder.Images, models.ImageInfo{
				Name: entry.Name(),
				Path: filepath.Join(cleanPath, entry.Name()),
			})
		}
	}

	return folder, nil
}

// isImageFile checks if a file is a supported image format
func (fs *FolderService) isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supportedExts := []string{".jpg", ".jpeg", ".png", ".webp", ".heic"}
	
	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			return true
		}
	}
	return false
}