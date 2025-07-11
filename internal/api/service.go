package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mejzh77/astragen/internal/repository"
)

type WebService struct {
	projectRepo *repository.ProjectRepository
	systemRepo  *repository.SystemRepository
	// ... другие репозитории
}

func NewWebService(
	projectRepo *repository.ProjectRepository,
	// ... другие зависимости
) *WebService {
	return &WebService{
		projectRepo: projectRepo,
		// ...
	}
}

func (s *WebService) RegisterRoutes(r *gin.Engine) {
	// Статические файлы
	r.Static("/static", "./internal/api/static")

	// API endpoints
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/tree", s.GetTreeData)
		apiGroup.POST("/sync", s.SyncData)
	}

	// HTML страницы
	r.GET("/", s.IndexPage)
	r.GET("/tree", s.TreePage)
}
