package repository

import (
	"fmt"
	"log"
	"strings"
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
func (r *SignalRepository) SaveSignals(signals []models.Signal) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, signal := range signals {
			// Нормализация всех строковых полей
			signal.Tag = normalizeString(signal.Tag)
			signal.NodeRef = normalizeString(signal.NodeRef)

			if len(signal.NodeRef) > 255 {
				signal.NodeRef = truncateUTF8(signal.NodeRef, 255)
			}

			// Включаем логирование SQL запроса
			tx = tx.Debug()

			result := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "tag"}},
				DoUpdates: clause.AssignmentColumns(getUpdateColumns(signal.SignalType)),
			}).Create(&signal)

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

func normalizeString(s string) string {
	if s == "" {
		return s
	}

	// Удаляем невалидные UTF-8 последовательности
	s = strings.ToValidUTF8(s, "")

	// Дополнительная проверка на одиночные байты кириллицы
	var builder strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError {
			// Пропускаем невалидные символы
			i++
			continue
		}
		builder.WriteRune(r)
		i += size
	}

	return builder.String()
}

// truncateUTF8 безопасно обрезает строку до n символов, не разрывая UTF-8 последовательности
func truncateUTF8(s string, n int) string {
	if len(s) <= n {
		return s
	}

	// Преобразуем в руны для корректного подсчета символов
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}

	return string(runes[:n])
}

//func (r *SignalRepository) SaveSignals(signals []models.Signal) error {
//return r.db.Transaction(func(tx *gorm.DB) error {
//for _, signal := range signals {
//// Upsert операция - обновляем если существует запись с таким же Tag
//if len(signal.NodeRef) > 50 {
//signal.NodeRef = signal.NodeRef[:50]
//}
//result := tx.Clauses(clause.OnConflict{
//Columns:   []clause.Column{{Name: "tag"}},
//DoUpdates: clause.AssignmentColumns(getUpdateColumns(signal.SignalType)),
//}).Create(&signal)

//if result.Error != nil {

// return fmt.Errorf("failed to save signal %s: %w", signal.Tag, result.Error)
// }
// }
// return nil
// })
// }
func getUpdateColumns(signalType string) []string {
	base := []string{
		"system", "equipment", "name", "module", "channel",
		"crate", "place", "property", "address", "modbus_addr",
		"node_ref", "fb", "check_status", "comment", "updated_at",
	}

	switch signalType {
	case "DI":
		return append(base, "category", "inversion", "ton", "tof")
	case "AI":
		return append(base, "range_min", "range_max", "unit", "sign",
			"warning_low", "warning_high", "alarm_low", "alarm_high",
			"format", "filter")
	case "DO":
		return base
	case "AO":
		return append(base, "range_min", "range_max", "unit")
	default:
		return base
	}
}
