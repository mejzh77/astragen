package main

import (
	"context"
	"fmt"
	"strconv"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsService struct {
	service *sheets.Service
}

type SyncService struct {
	sheetsService *gsheets.GoogleSheetsService
	userRepo      *repository.UserRepository
}

func NewService(ctx context.Context, credentialsJSON []byte) (*GoogleSheetsService, error) {
	client, err := google.JWTConfigFromJSON(
		credentialsJSON,
		sheets.SpreadsheetsReadonlyScope,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT config: %w", err)
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client.Client(ctx)))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &GoogleSheetsService{service: srv}, nil
}

func (s *GoogleSheetsService) ReadSheet(spreadsheetID, readRange string) ([][]interface{}, error) {
	resp, err := s.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet: %w", err)
	}
	return resp.Values, nil
}

func ParseUsers(rows [][]interface{}) []models.User {
	var users []models.User

	for i, row := range rows {
		// Пропускаем заголовок
		if i == 0 {
			continue
		}

		user := models.User{
			GoogleID: fmt.Sprint(row[0]),
			Name:     fmt.Sprint(row[1]),
			Email:    fmt.Sprint(row[2]),
		}

		// Обработка чисел
		if age, err := strconv.Atoi(fmt.Sprint(row[3])); err == nil {
			user.Age = age
		}

		users = append(users, user)
	}
	return users
}

func (s *SyncService) RunSync(ctx context.Context) error {
	// 1. Чтение данных из Google Sheets
	spreadsheetID := "your-spreadsheet-id"
	readRange := "Users!A2:E"
	rows, err := s.sheetsService.ReadSheet(spreadsheetID, readRange)
	if err != nil {
		return fmt.Errorf("failed to read sheet: %w", err)
	}

	// 2. Парсинг данных
	users := gsheets.ParseUsers(rows)

	return nil
}
