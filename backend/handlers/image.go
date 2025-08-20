package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"kerpic-backend/models"
	"kerpic-backend/services"
)

type ImageHandler struct {
	imageService *services.ImageService
}

func NewImageHandler(imageService *services.ImageService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

// ListImages returns images in a folder
func (ih *ImageHandler) ListImages(c *gin.Context) {
	folder := c.Param("folder")
	
	images, err := ih.imageService.ListImages(folder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, images)
}

// GetThumbnail serves a thumbnail for an image
func (ih *ImageHandler) GetThumbnail(c *gin.Context) {
	path := c.Param("path")
	// Remove leading slash from path parameter
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	// Get size parameter (default to 200)
	sizeStr := c.DefaultQuery("size", "200")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 || size > 1000 {
		size = 200
	}

	thumbnailPath, err := ih.imageService.GenerateThumbnail(path, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.File(thumbnailPath)
}

// CreateBatch creates a new comparison batch
func (ih *ImageHandler) CreateBatch(c *gin.Context) {
	var req models.BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	batchID, images, err := ih.imageService.CreateBatch(req.ImagePaths)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := models.BatchResponse{
		ID:     batchID,
		Images: images,
	}

	c.JSON(http.StatusOK, response)
}

// SaveSelected saves selected images to the target folder
func (ih *ImageHandler) SaveSelected(c *gin.Context) {
	var req models.SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := ih.imageService.SaveSelected(req.BatchID, req.SelectedPaths, req.TargetFolder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}