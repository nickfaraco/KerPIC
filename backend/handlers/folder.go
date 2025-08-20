package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"kerpic-backend/services"
)

type FolderHandler struct {
	folderService *services.FolderService
}

func NewFolderHandler(folderService *services.FolderService) *FolderHandler {
	return &FolderHandler{
		folderService: folderService,
	}
}

// ListFolders returns available root folders
func (fh *FolderHandler) ListFolders(c *gin.Context) {
	folders, err := fh.folderService.ListFolders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, folders)
}

// GetFolderContents returns contents of a specific folder
func (fh *FolderHandler) GetFolderContents(c *gin.Context) {
	path := c.Param("path")
	// Remove leading slash from path parameter
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	folder, err := fh.folderService.GetFolderContents(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, folder)
}