package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mejzh77/astragen/internal/repository"
)

type TreeHandler struct {
	projectRepo *repository.ProjectRepository
}

func NewTreeHandler(pr *repository.ProjectRepository) *TreeHandler {
	return &TreeHandler{projectRepo: pr}
}

func (h *TreeHandler) GetTreeData(c *gin.Context) {
	projects, err := h.projectRepo.GetAllWithHierarchy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, projects)
}

func (h *TreeHandler) TreePage(c *gin.Context) {
	c.HTML(http.StatusOK, "tree.html", gin.H{
		"title": "Древовидная структура",
	})
}
