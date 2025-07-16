package repository

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mejzh77/astragen/pkg/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SignalRepository struct {
	db *gorm.DB
}

func NewSignalRepository(db *gorm.DB) *SignalRepository {
	return &SignalRepository{db: db}
}

func (r *SignalRepository) SaveSignals(signals []models.Signal, debug bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, signal := range signals {
			// Нормализация всех строковых полей
			signal.Tag = normalizeString(signal.Tag)
			signal.NodeRef = normalizeString(signal.NodeRef)

			if len(signal.NodeRef) > 255 {
				signal.NodeRef = truncateUTF8(signal.NodeRef, 255)
			}

			if debug {
				tx = tx.Debug()
			}
			result := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "tag"}},
				DoUpdates: clause.AssignmentColumns(getUpdateColumnsForSignalType(signal.SignalType)),
			}).Create(&signals[i])

			if result.Error != nil {
				log.Printf("Детали ошибки для сигнала %s:", signal.Tag)
				log.Printf("NodeRef (обрезано): %s", signal.NodeRef)
				log.Printf("NodeRef bytes: % x", []byte(signal.NodeRef))
				return fmt.Errorf("failed to save signal %s: %w", signal.Tag, result.Error)
			}
		}
		return nil
	})
}
func getUpdateColumnsForSignalType(signalType string) []string {
	// Базовые поля, общие для всех типов сигналов
	baseFields := []string{
		"system_id", "equipment", "name", "module", "channel",
		"crate", "place", "property", "address", "modbus_addr", "node_id",
		"node_ref", "fb", "check_status", "comment", "updated_at",
		"value", "product_id",
	}

	switch signalType {
	case "DI":
		return append(baseFields,
			"category", "inversion", "ton", "tof",
		)
	case "AI":
		return append(baseFields,
			"range_min", "range_max", "unit", "sign",
			"warning_low", "warning_high", "alarm_low", "alarm_high",
			"format", "filter",
		)
	case "DO":
		return baseFields
	case "AO":
		return append(baseFields,
			"range_min", "range_max", "unit",
		)
	default:
		return baseFields
	}
}

// getStructFields - альтернативный вариант с рефлексией (если предпочтете)
func getStructFields(signalType string) []string {
	s := models.Signal{}
	t := reflect.TypeOf(s)
	var fields []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Пропускаем поля, которые не должны обновляться
		if field.Name == "Model" || field.Name == "ID" || field.Name == "CreatedAt" || field.Name == "DeletedAt" {
			continue
		}
		// Пропускаем некоторые специфичные поля
		if field.Name == "Tag" {
			continue
		}
		// Для разных типов сигналов включаем/исключаем определенные поля
		switch signalType {
		case "DI":
			if field.Name == "RangeMin" || field.Name == "RangeMax" ||
				field.Name == "Unit" || field.Name == "Sign" ||
				field.Name == "WarningLow" || field.Name == "WarningHigh" ||
				field.Name == "AlarmLow" || field.Name == "AlarmHigh" ||
				field.Name == "Format" || field.Name == "Filter" {
				continue
			}
		case "AI":
			if field.Name == "Category" || field.Name == "Inversion" ||
				field.Name == "TON" || field.Name == "TOF" {
				continue
			}
		case "DO":
			if field.Name == "Category" || field.Name == "Inversion" ||
				field.Name == "TON" || field.Name == "TOF" ||
				field.Name == "RangeMin" || field.Name == "RangeMax" ||
				field.Name == "Unit" || field.Name == "Sign" ||
				field.Name == "WarningLow" || field.Name == "WarningHigh" ||
				field.Name == "AlarmLow" || field.Name == "AlarmHigh" ||
				field.Name == "Format" || field.Name == "Filter" {
				continue
			}
		case "AO":
			if field.Name == "Category" || field.Name == "Inversion" ||
				field.Name == "TON" || field.Name == "TOF" ||
				field.Name == "Sign" || field.Name == "WarningLow" ||
				field.Name == "WarningHigh" || field.Name == "AlarmLow" ||
				field.Name == "AlarmHigh" || field.Name == "Format" ||
				field.Name == "Filter" {
				continue
			}
		}
		// Преобразуем имя поля в snake_case для БД
		fields = append(fields, toSnakeCase(field.Name))
	}
	return fields
}

// toSnakeCase преобразует CamelCase в snake_case
// toSnakeCase правильно преобразует CamelCase в snake_case, учитывая аббревиатуры
func toSnakeCase(s string) string {
	var result []rune
	var prev, next rune

	for i, r := range s {
		if i > 0 {
			prev = rune(s[i-1])
			if i < len(s)-1 {
				next = rune(s[i+1])
			} else {
				next = 0
			}

			// Вставляем подчеркивание перед заглавной буквой, если:
			// 1. Это не первая буква И
			// 2. Предыдущий символ строчный ИЛИ
			// 3. Следующий символ строчный (для обработки аббревиатур типа "ID")
			if r >= 'A' && r <= 'Z' {
				if (prev >= 'a' && prev <= 'z') ||
					(next >= 'a' && next <= 'z' && next != 0) {
					result = append(result, '_')
				}
			}
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// Проверка существования узлов с возвратом списка отсутствующих
func verifyNodesExist(tx *gorm.DB, nodeIDs []uint) (missingIDs []uint, err error) {
	var existingIDs []uint
	if err := tx.Model(&models.Node{}).
		Where("id IN ?", nodeIDs).
		Pluck("id", &existingIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to verify nodes: %w", err)
	}

	// Создаем мапу для быстрого поиска
	existingMap := make(map[uint]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	// Находим отсутствующие
	for _, id := range nodeIDs {
		if !existingMap[id] {
			missingIDs = append(missingIDs, id)
		}
	}

	return missingIDs, nil
}
func (r *SignalRepository) UpdateSignalNodes(signals []models.Signal) error {
	// Сначала собираем все данные для массового обновления
	updates := make(map[uint]uint) // signalID -> nodeID

	for _, signal := range signals {
		if signal.NodeID != nil {
			updates[signal.ID] = *signal.NodeID // Пропускаем сигналы без NodeID
		}
	}

	if len(updates) == 0 {
		return nil // Нет чего обновлять
	}

	return r.db.Debug().Transaction(func(tx *gorm.DB) error {
		// Проверяем существование всех узлов
		var nodeIDs []uint
		for _, nodeID := range updates {
			nodeIDs = append(nodeIDs, nodeID)
		}

		missing, err := verifyNodesExist(tx, nodeIDs)
		if err != nil {
			return err
		}
		if len(missing) > 0 {
			return fmt.Errorf("nodes with IDs %v do not exist", missing)
		}

		// Массовое обновление одним запросом
		return tx.Model(&models.Signal{}).
			Where("id IN ?", getKeys(updates)).
			Update("node_id", gorm.Expr(
				"CASE id "+buildCaseWhen(updates)+" END",
			)).Error
	})
}

// Вспомогательная функция для построения CASE WHEN выражения
func buildCaseWhen(updates map[uint]uint) string {
	var builder strings.Builder
	for signalID, nodeID := range updates {
		builder.WriteString(fmt.Sprintf("WHEN %d THEN %d ", signalID, nodeID))
	}
	return builder.String()
}

// Вспомогательная функция для получения ключей мапы
func getKeys(m map[uint]uint) []uint {
	keys := make([]uint, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func normalizeString(s string) string {
	if s == "" {
		return s
	}

	s = strings.ToValidUTF8(s, "")

	var builder strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError {
			i++
			continue
		}
		builder.WriteRune(r)
		i += size
	}

	return builder.String()
}

func truncateUTF8(s string, n int) string {
	if len(s) <= n {
		return s
	}

	runes := []rune(s)
	if len(runes) <= n {
		return s
	}

	return string(runes[:n])
}
