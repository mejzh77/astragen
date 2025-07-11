package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mejzh77/astragen/internal/sync"
)

type WebService struct {
	syncService *sync.SyncService
}

func NewWebService(syncService *sync.SyncService) *WebService {
	return &WebService{
		syncService: syncService,
	}
}

func (s *WebService) RegisterRoutes(r *gin.Engine) {
	r.GET("/", s.IndexPage)
	r.GET("/tree", s.TreePage)
	r.POST("/api/sync", s.SyncData)
	r.GET("/api/tree-data", s.GetTreeData)
	r.GET("/api/details", s.getItemDetails)
}

func (s *WebService) IndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "ASTRA Dashboard",
	})
}

func (s *WebService) TreePage(c *gin.Context) {
	treeData, err := s.syncService.GetTreeData()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "tree.html", gin.H{
		"title": "Project Structure",
		"tree":  treeData,
	})
}

func (s *WebService) GetTreeData(c *gin.Context) {
	treeData, err := s.syncService.GetTreeData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get tree data",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, treeData)
}
func (s *WebService) getItemDetails(c *gin.Context) {
	itemType := c.Query("type")
	itemID := c.Query("id")

	var result interface{}
	var err error

	switch itemType {
	case "project":
		result, err = s.syncService.GetProjectDetails(itemID)
	case "system":
		result, err = s.syncService.GetSystemDetails(itemID)
	case "node":
		result, err = s.syncService.GetNodeDetails(itemID)
	case "product":
		result, err = s.syncService.GetProductDetails(itemID)
	case "functionblock":
		result, err = s.syncService.GetFunctionBlockDetails(itemID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
func (s *WebService) SyncData(c *gin.Context) {
	if err := s.syncService.RunFullSync(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Sync failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Database synchronized",
	})
}
