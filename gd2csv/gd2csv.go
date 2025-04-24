package gd2csv

import (
	"bytes"
	"context"
	"encoding/csv"
	"log"
	"net/http"
	"os"

	"google.golang.org/api/option"
	"gopkg.in/Iwark/spreadsheet.v2"
)

// Addslashes добавляет экранирование символов в строку
func Addslashes(str string) string {
	var buf bytes.Buffer
	for _, char := range str {
		switch char {
		case '\'', '"', '\\':
			buf.WriteRune('\\')
		case '/':
			continue
		}
		buf.WriteRune(char)
	}
	return buf.String()
}

// checkError обрабатывает ошибки
func checkError(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// getClientJWT создает HTTP-клиент с использованием JWT
func getClientJWT(credentialsFile string) *http.Client {
	ctx := context.Background()

	// Загрузка учетных данных сервисного аккаунта
	client, err := option.WithCredentialsFile(credentialsFile).Client(ctx)
	checkError(err)

	return client
}

// Update обновляет данные из Google Sheets и сохраняет их в CSV файлы
func Update(spreadsheetID, dir string, credentialsFile string) {
	// Создание клиента с использованием JWT
	client := getClientJWT(credentialsFile)

	// Инициализация сервиса Google Sheets
	service := spreadsheet.NewServiceWithClient(client)
	spreadsheet, err := service.FetchSpreadsheet(spreadsheetID)
	checkError(err)

	// Обработка всех листов в таблице
	sheets := spreadsheet.Sheets
	for _, sheet := range sheets {
		data := make([][]string, 0)
		for _, row := range sheet.Rows {
			r := make([]string, 0)
			for _, cell := range row {
				r = append(r, cell.Value)
			}
			data = append(data, r)
		}

		// Пример специальной обработки для листа "C6"
		if sheet.Properties.Title == "C6" {
			data[1][2] = data[1][2] + " " + data[0][2]
			data[1][3] = data[1][3] + " " + data[0][2]
			data[1][4] = data[1][4] + " " + data[0][2]
			data[1][7] = data[1][7] + " " + data[0][7]
			data[1][8] = data[1][8] + " " + data[0][7]
			data[1][9] = data[1][9] + " " + data[0][7]
			data = data[1:]
		}

		// Сохранение данных в CSV файл
		title := sheet.Properties.Title
		file, err := os.Create(dir + "/" + Addslashes(title) + ".csv")
		checkError(err)
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		err = writer.WriteAll(data)
		checkError(err)
	}
}

// Clear удаляет все файлы в указанной директории
func Clear(dir string) {
	err := os.RemoveAll(dir)
	checkError(err)
}
