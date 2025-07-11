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
