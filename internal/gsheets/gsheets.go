package gsheets

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"google.golang.org/api/sheets/v4"
	"github.com/mejzh77/astragen/pkg/models"
)

type GoogleSheetsService struct {
	service *sheets.Service
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

func (s *GoogleSheetsService) RunSync(ctx context.Context) error {
	// 1. Чтение данных из Google Sheets
	signals := []models.AI{}
	spreadsheetID := "1GAUwJRTtrBT4gr1y3ETsCSlHojrc7VCD2GlGDUM53kQ"
	readRange := "AI!A2:E"	
	rows, err := s.ReadSheet(spreadsheetID, readRange)
	if err != nil {
		return fmt.Errorf("failed to read sheet: %w", err)
	}
	// 2. Парсинг данных
	Unmarshal(rows, &signals)
	fmt.Println(signals)
	return nil
}
