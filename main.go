package main

import (
	"fmt"
	"log"

	"github.com/mejzh77/astragen/config"
	"github.com/mejzh77/astragen/data"
	"github.com/mejzh77/astragen/generation"
	"github.com/mejzh77/astragen/parsing"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Обновление данных из Google Sheets, если требуется
	if cfg.Update {
		err = data.UpdateFromGoogle(cfg.GSID, "src")
		if err != nil {
			log.Fatalf("Failed to update data from Google Sheets: %v", err)
		}
		fmt.Println("Завершено чтение из таблицы Google")
	}

	// Чтение и парсинг CSV файла
	inputFile := "src/ITF.csv"
	itfs, err := parsing.ParseCSV(inputFile)
	if err != nil {
		log.Fatalf("Failed to parse CSV: %v", err)
	}
	fmt.Println("Завершено чтение CSV")

	// Группировка интерфейсов по системам
	systems, err := parsing.GroupInterfacesBySystem(itfs)
	if err != nil {
		log.Fatalf("Failed to group interfaces by system: %v", err)
	}

	// Обработка каждой системы
	for sysName, sys := range systems {
		err = data.ProcessSystem(&sys, cfg)
		if err != nil {
			log.Fatalf("Failed to process system %s: %v", sysName, err)
		}
		systems[sysName] = sys
	}

	// Генерация выходных данных для целевой системы
	targetSystem := systems[cfg.System]
	generation.GenerateVariables(targetSystem, cfg)
	generation.GenerateFunctionalBlocks(targetSystem, cfg)
	generation.GenerateOMX(targetSystem, cfg)
	generation.GenerateOPC(targetSystem, cfg)
}
