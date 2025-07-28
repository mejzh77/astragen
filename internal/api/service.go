package api

import (
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/internal/sync"
	"html/template"
	"log"
	"net/http"
	"strings"
	stdsync "sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все origin (в продакшене нужно ограничить)
	},
}

type WebService struct {
	syncService  *sync.SyncService
	clients      map[*websocket.Conn]bool
	clientsMutex stdsync.Mutex
	router       *gin.Engine
}

func NewWebService(syncService *sync.SyncService) *WebService {
	return &WebService{
		syncService: syncService,
		clients:     make(map[*websocket.Conn]bool),
		router:      gin.Default(),
	}
}

//	func (s *WebService) SetupTemplates() {
//		tpl := template.New("base").Funcs(template.FuncMap{
//			"hasChildren": func(item interface{}) bool {
//				m, ok := item.(map[string]interface{})
//				if !ok {
//					return false
//				}
//				return m["systems"] != nil || m["nodes"] != nil ||
//					m["products"] != nil || m["functionBlocks"] != nil
//			},
//		})
//
//		// Сначала layout, потом остальные
//		files := []string{
//			"templates/layout.html",
//			"templates/index.html",
//			"templates/config.html",
//			"templates/generate.html",
//			"templates/tree.html",
//		}
//
//		tpl = template.Must(tpl.ParseFiles(files...))
//
//		// Логгируем загруженные шаблоны
//		for _, t := range tpl.Templates() {
//			log.Printf("Template: %s (defined: %v)", t.Name(), t.DefinedTemplates())
//		}
//
//		s.router.HTMLRender = &TemplRender{Template: tpl}
//	}
//
//// Кастомный рендерер для Gin
//type TemplRender struct {
//	Template *template.Template
//}
//
//func (t *TemplRender) Instance(name string, data interface{}) render.Render {
//	log.Printf("Rendering template: %s (defined: %s)", name, t.Template.DefinedTemplates())
//	return render.HTML{
//		Template: t.Template,
//		Name:     name,
//		Data:     data,
//	}
//}

func (s *WebService) Run(addr string) {
	s.router.HTMLRender = ginview.New(goview.Config{
		Root:      "templates",
		Extension: ".html",
		Master:    "layout",
		Partials:  []string{},
		Funcs: template.FuncMap{
			"hasChildren": func(item interface{}) bool {
				m, ok := item.(map[string]interface{})
				if !ok {
					return false
				}
				return m["systems"] != nil || m["nodes"] != nil ||
					m["products"] != nil || m["functionBlocks"] != nil
			},
		},
		DisableCache: true,
	})
	s.router.Static("/static", "./static")
	err := s.router.Run(addr)
	if err != nil {
		log.Fatal(err)
	}
}
func (s *WebService) RegisterRoutes() {
	s.router.GET("/", s.IndexPage)
	s.router.GET("/tree", s.TreePage)
	s.router.POST("/api/sync", s.SyncData)
	s.router.GET("/api/tree-data", s.GetTreeData)
	s.router.GET("/api/details", s.getItemDetails)
	s.router.GET("/api/config", s.GetConfig)
	s.router.POST("/api/config", s.UpdateConfig)
	s.router.GET("/config", s.ConfigPage)
	s.router.GET("/ws", s.handleWebSocket)
	s.router.GET("/generate", s.GenerateImportPage)
	s.router.POST("/api/generate-import", s.GenerateImportFile)
	s.router.POST("/api/regenerate-import-files", s.RegenerateAllImportFiles)
	s.router.GET("/api/nodes", s.GetNodesBySystem)

}

// Обработчик WebSocket
func (s *WebService) handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Добавляем клиента
	s.clientsMutex.Lock()
	s.clients[conn] = true
	s.clientsMutex.Unlock()

	defer func() {
		// Удаляем клиента при отключении
		s.clientsMutex.Lock()
		delete(s.clients, conn)
		s.clientsMutex.Unlock()
		conn.Close()
	}()

	// Просто слушаем соединение (сообщения не обрабатываем)
	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
}
func (s *WebService) IndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index", gin.H{
		"title": "ПТК AstraRegul",
	})
}
func (s *WebService) ConfigPage(c *gin.Context) {
	c.HTML(http.StatusOK, "config", gin.H{
		"title": "Редактор конфигурации",
	})
}
func (s *WebService) GenerateImportPage(c *gin.Context) {
	systems, _ := s.syncService.GetAllSystems()
	cdsTypes, _ := s.syncService.GetAllCDSTypes()

	c.HTML(http.StatusOK, "generate", gin.H{
		"title":    "Generate Import Files",
		"systems":  systems,
		"cdsTypes": cdsTypes,
		"fbTypes":  getAvailableFBTypes(), // из конфига
	})
}

func getAvailableFBTypes() []string {
	var types []string
	for k := range config.Cfg.FunctionBlocks {
		types = append(types, k)
	}
	return types
}
func (s *WebService) GenerateImportFile(c *gin.Context) {
	var request struct {
		System   string `json:"system"`
		CdsType  string `json:"cdsType"`
		Node     string `json:"node"`
		FileType string `json:"fileType"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fbs, err := s.syncService.GetFilteredFunctionBlocks(request.System, request.CdsType, request.Node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var content strings.Builder
	for _, fb := range fbs {
		switch request.FileType {
		case "STDecl":
			if fb.Call != "" {
				content.WriteString(fb.Declaration + "\n\n")
			}
		case "ST":
			if fb.Call != "" {
				content.WriteString(fb.Call + "\n\n")
			}
		case "OMX":
			if fb.OMX != "" {
				content.WriteString(fb.OMX + "\n\n")
			}
		case "OPC":
			if fb.OPC != "" {
				content.WriteString(fb.OPC + "\n\n")
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"content": content.String(),
		"count":   len(fbs),
	})
}

// Добавить в service.go
func (s *WebService) GetConfig(c *gin.Context) {
	config, err := s.syncService.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get config",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, config)
}
func (s *WebService) GetNodesBySystem(c *gin.Context) {
	system := c.Query("system")
	nodes, err := s.syncService.GetNodesBySystem(system)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nodes)
}
func (s *WebService) UpdateConfig(c *gin.Context) {
	var updates map[string]interface{}
	if err := c.BindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if err := s.syncService.UpdateConfig(updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update config",
			"details": err.Error(),
		})
		return
	}
	// Рассылаем уведомление всем клиентам
	s.clientsMutex.Lock()
	for client := range s.clients {
		if err := client.WriteMessage(websocket.TextMessage, []byte("config_updated")); err != nil {
			log.Printf("Failed to send WS message: %v", err)
			delete(s.clients, client)
			client.Close()
		}
	}
	s.clientsMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (s *WebService) TreePage(c *gin.Context) {
	treeData, err := s.syncService.GetTreeData()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "tree", gin.H{
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
func (s *WebService) RegenerateAllImportFiles(c *gin.Context) {
	content, err := s.syncService.RegenerateAllImportFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": content,
		"count":   len(content),
	})
}
