package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mejzh77/astragen/pkg/models"
	"gorm.io/gorm"
)

type TreeAPI struct {
	db *gorm.DB
}

func NewTreeAPI(db *gorm.DB) *TreeAPI {
	return &TreeAPI{db: db}
}

// GetProjects возвращает список проектов
func (api *TreeAPI) GetProjects(c *gin.Context) {
	var projects []models.Project
	if err := api.db.Find(&projects).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, projects)
}

// GetSystems возвращает системы проекта
func (api *TreeAPI) GetSystems(c *gin.Context) {
	projectID := c.Param("project_id")
	var systems []models.System
	if err := api.db.Where("project_id = ?", projectID).Find(&systems).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, systems)
}

// GetSystemChildren возвращает содержимое системы
func (api *TreeAPI) GetSystemChildren(c *gin.Context) {
	systemID := c.Param("system_id")
	systemType := c.Query("type") // "product" или "node"

	if systemType == "product" {
		var products []models.Product
		if err := api.db.Where("system_id = ?", systemID).Find(&products).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"products": products})
	} else {
		var nodes []models.Node
		if err := api.db.Where("system_id = ?", systemID).Find(&nodes).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"nodes": nodes})
	}
}

// GetProductSignals возвращает сигналы изделия
func (api *TreeAPI) GetProductSignals(c *gin.Context) {
	productID := c.Param("product_id")
	var signals []models.Signal
	if err := api.db.Where("product_id = ?", productID).Find(&signals).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, signals)
}

// GetNodeFunctionBlocks возвращает ФБ узла
func (api *TreeAPI) GetNodeFunctionBlocks(c *gin.Context) {
	nodeID := c.Param("node_id")
	var fbs []models.FunctionBlock
	if err := api.db.Where("node_id = ?", nodeID).Find(&fbs).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, fbs)
}

// GetFBVariables возвращает переменные ФБ
func (api *TreeAPI) GetFBVariables(c *gin.Context) {
	fbID := c.Param("fb_id")
	var variables []models.FBVariable
	if err := api.db.Where("fb_id = ?", fbID).Find(&variables).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, variables)
}
