package services

import (
	"crypto/md5"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"kerpic-backend/models"
)

type ImageService struct {
	baseDir      string
	cacheDir     string
	batches      map[string][]models.ImageInfo
	batchesMutex sync.RWMutex
	cache        map[string]models.ImageInfo
	cacheMutex   sync.RWMutex
}

func NewImageService(baseDir, cacheDir string) *ImageService {
	// Ensure cache directory exists
	os.MkdirAll(cacheDir, 0755)
	
	return &ImageService{
		baseDir:  baseDir,
		cacheDir: cacheDir,
		batches:  make(map[string][]models.ImageInfo),
		cache:    make(map[string]models.ImageInfo),
	}
}

// GetImageInfo returns detailed information about an image
func (is *ImageService) GetImageInfo(relativePath string) (*models.ImageInfo, error) {
	is.cacheMutex.RLock()
	if cached, exists := is.cache[relativePath]; exists {
		is.cacheMutex.RUnlock()
		return &cached, nil
	}
	is.cacheMutex.RUnlock()

	// Sanitize path
	cleanPath := filepath.Clean(relativePath)
	if strings.Contains(cleanPath, "..") {
		return nil, os.ErrPermission
	}

	fullPath := filepath.Join(is.baseDir, cleanPath)

	// Get file info
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	imageInfo := models.ImageInfo{
		Name:    filepath.Base(cleanPath),
		Path:    cleanPath,
		Size:    fileInfo.Size(),
		ModTime: fileInfo.ModTime(),
	}

	// Try to get image dimensions and EXIF data
	file, err := os.Open(fullPath)
	if err == nil {
		defer file.Close()

		// Decode image to get dimensions
		img, _, err := image.DecodeConfig(file)
		if err == nil {
			imageInfo.Width = img.Width
			imageInfo.Height = img.Height
		}

		// Reset file position for EXIF reading
		file.Seek(0, 0)

		// Try to read EXIF data
		if exifData, err := exif.Decode(file); err == nil {
			if orient, err := exifData.Get(exif.Orientation); err == nil {
				if orientVal, err := orient.Int(0); err == nil {
					imageInfo.Orientation = orientVal
				}
			}
		}
	}

	// Generate thumbnail URL
	imageInfo.ThumbnailURL = fmt.Sprintf("/api/thumbnail/%s", cleanPath)

	// Cache the result
	is.cacheMutex.Lock()
	is.cache[relativePath] = imageInfo
	is.cacheMutex.Unlock()

	return &imageInfo, nil
}

// ListImages returns detailed info for images in a folder
func (is *ImageService) ListImages(folderPath string) ([]models.ImageInfo, error) {
	cleanPath := filepath.Clean(folderPath)
	if strings.Contains(cleanPath, "..") {
		return nil, os.ErrPermission
	}

	fullPath := filepath.Join(is.baseDir, cleanPath)
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var images []models.ImageInfo
	for _, entry := range entries {
		if !entry.IsDir() && is.isImageFile(entry.Name()) {
			imagePath := filepath.Join(cleanPath, entry.Name())
			if imageInfo, err := is.GetImageInfo(imagePath); err == nil {
				images = append(images, *imageInfo)
			}
		}
	}

	return images, nil
}

// GenerateThumbnail creates a thumbnail for an image
func (is *ImageService) GenerateThumbnail(relativePath string, size int) (string, error) {
	// Create cache filename
	hash := fmt.Sprintf("%x", md5.Sum([]byte(relativePath)))
	cacheFilename := fmt.Sprintf("%s_%d.jpg", hash, size)
	cachePath := filepath.Join(is.cacheDir, cacheFilename)

	// Check if thumbnail already exists
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	// Sanitize path
	cleanPath := filepath.Clean(relativePath)
	if strings.Contains(cleanPath, "..") {
		return "", os.ErrPermission
	}

	fullPath := filepath.Join(is.baseDir, cleanPath)

	// Open and decode image
	src, err := imaging.Open(fullPath)
	if err != nil {
		return "", err
	}

	// Apply EXIF orientation
	if imageInfo, err := is.GetImageInfo(relativePath); err == nil {
		src = is.applyOrientation(src, imageInfo.Orientation)
	}

	// Create thumbnail
	thumbnail := imaging.Thumbnail(src, size, size, imaging.Lanczos)

	// Save thumbnail
	err = imaging.Save(thumbnail, cachePath, imaging.JPEGQuality(80))
	if err != nil {
		return "", err
	}

	return cachePath, nil
}

// CreateBatch creates a new comparison batch
func (is *ImageService) CreateBatch(imagePaths []string) (string, []models.ImageInfo, error) {
	batchID := fmt.Sprintf("batch_%d", len(is.batches))
	
	var images []models.ImageInfo
	for _, path := range imagePaths {
		if imageInfo, err := is.GetImageInfo(path); err == nil {
			images = append(images, *imageInfo)
		}
	}

	is.batchesMutex.Lock()
	is.batches[batchID] = images
	is.batchesMutex.Unlock()

	return batchID, images, nil
}

// SaveSelected moves selected images to the saved folder
func (is *ImageService) SaveSelected(batchID string, selectedPaths []string, targetFolder string) (*models.SaveResponse, error) {
	if targetFolder == "" {
		targetFolder = "saved"
	}

	response := &models.SaveResponse{
		Success:      []string{},
		Failed:       []string{},
		Conflicts:    []string{},
		TargetFolder: targetFolder,
	}

	for _, imagePath := range selectedPaths {
		cleanPath := filepath.Clean(imagePath)
		if strings.Contains(cleanPath, "..") {
			response.Failed = append(response.Failed, imagePath)
			continue
		}

		sourcePath := filepath.Join(is.baseDir, cleanPath)
		
		// Create target directory
		targetDir := filepath.Join(filepath.Dir(sourcePath), targetFolder)
		os.MkdirAll(targetDir, 0755)

		// Determine target filename with conflict resolution
		filename := filepath.Base(sourcePath)
		targetPath := filepath.Join(targetDir, filename)
		
		// Handle filename conflicts
		counter := 1
		for {
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				break
			}
			
			// Check if it's the same file (same size and mod time)
			if is.isSameFile(sourcePath, targetPath) {
				response.Conflicts = append(response.Conflicts, imagePath)
				goto nextImage
			}

			// Generate new filename with counter
			ext := filepath.Ext(filename)
			nameWithoutExt := strings.TrimSuffix(filename, ext)
			newFilename := fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext)
			targetPath = filepath.Join(targetDir, newFilename)
			counter++
		}

		// Move the file
		if err := os.Rename(sourcePath, targetPath); err != nil {
			response.Failed = append(response.Failed, imagePath)
		} else {
			response.Success = append(response.Success, imagePath)
		}

		nextImage:
	}

	return response, nil
}

// Helper methods

func (is *ImageService) isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supportedExts := []string{".jpg", ".jpeg", ".png", ".webp", ".heic"}
	
	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			return true
		}
	}
	return false
}

func (is *ImageService) applyOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Transpose(img)
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Transverse(img)
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}

func (is *ImageService) isSameFile(path1, path2 string) bool {
	info1, err1 := os.Stat(path1)
	info2, err2 := os.Stat(path2)
	
	if err1 != nil || err2 != nil {
		return false
	}

	return info1.Size() == info2.Size() && info1.ModTime().Equal(info2.ModTime())
}