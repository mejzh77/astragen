package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mejzh77/astragen/internal/api"
	"github.com/mejzh77/astragen/internal/database"
	"github.com/mejzh77/astragen/internal/sync"
)

func main() {
	// Инициализация SyncService (как в основном приложении)
	db := database.InitDB("your-dsn")
	syncService := sync.NewSyncService(
		nil, // gsheets service если нужно
		db,
	)

	// Создание веб-сервиса
	webService := api.NewWebService(syncService)

	// Настройка роутера
	r := gin.Default()
	webService.RegisterRoutes(r)

	// Запуск сервера
	log.Fatal(r.Run(":8080"))
}
