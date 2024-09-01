package sheet

import (
	"context"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"log"
	"os"
	"sailerBot/logger"
	"strconv"
)

func InitSheet() (*sheets.Service, string, context.Context) {
	// Путь к JSON-файлу с учетными данными
	credentialsFile := "config/GoogleSheets/Gkey.json"

	// Чтение учетных данных
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// Получение конфигурации клиента
	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// Создание клиента
	ctx := context.Background()
	client := config.Client(ctx)

	// Подключение к Google Sheets API
	service, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	spreadsheetId := "1z1mLFjgMi93FPyFOL2N3y-6IvtTyD9MAQ-Snb2Re7YQ"
	return service, spreadsheetId, ctx
}
func AddToCart(userID int64, productPrice, productLink, productSize string) error {
	srv, spreadsheetId, _ := InitSheet() // Инициализация сервиса и идентификатора таблицы

	writeRange := "Cart!A:D" // Диапазон, куда будет добавлена новая строка

	// Подготовка данных для записи
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{
		userID,
		productPrice,
		productLink,
		productSize,
	})

	// Метод Append добавляет данные в конец указанного диапазона
	_, err := srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, &vr).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	logger.Debug.Printf("Data successfully added to cart!")
	return nil
}
func ReadCart(userID int64) (map[string][]string, error) {
	srv, spreadsheetId, _ := InitSheet()

	readRange := "Cart!A:D"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		logger.Error.Println("Unable to retrieve data from sheet: %v", err)
		return nil, err
	}
	if len(resp.Values) == 0 {
		logger.Error.Println("No data found in range: %v", readRange)
		return nil, nil
	}
	// Конвертация userID в строку для поиска
	//userIDStr := strconv.FormatInt(userID, 10)

	result := make(map[string][]string)

	// Первую строку предполагаем как заголовки столбцов
	headers := resp.Values[0]

	for _, row := range resp.Values[1:] { // Пропускаем заголовок
		if len(row) > 0 {
			// Попытка приведения первого столбца к int64
			id, err := strconv.ParseInt(row[0].(string), 10, 64)
			if err != nil {
				continue // Пропускаем строки с некорректным userID
			}

			// Сравниваем id с userID
			if id == userID {
				for i, value := range row {
					header := headers[i].(string)
					result[header] = append(result[header], fmt.Sprintf("%v", value))
				}
			}
		}
	}

	return result, nil
}
func AddProductsByToken(token string, link string, size string, price int) error {
	srv, spreadsheetId, _ := InitSheet() // Инициализация сервиса и идентификатора таблицы
	writeRange := "ProductsByToken!A:D"
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{
		token,
		link,
		size,
		price,
	})

	// Метод Append добавляет данные в конец указанного диапазона
	_, err := srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, &vr).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	logger.Debug.Printf("Data successfully added to ProductsByToken!")
	return nil
}
func AddOrder(userID int64, token string, name string, address string, phone string, status string) error {
	srv, spreadsheetId, _ := InitSheet() // Инициализация сервиса и идентификатора таблицы
	writeRange := "Orders!A:F"
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{
		userID,
		token,
		name,
		address,
		phone,
		status,
	})

	// Метод Append добавляет данные в конец указанного диапазона
	_, err := srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, &vr).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	logger.Debug.Printf("Data successfully added to Orders!")
	return nil
}
func DeleteFromCart(userID int64, position int) error {
	// Инициализация таблицы
	srv, spreadsheetID, ctx := InitSheet()

	// Задаем диапазон для поиска заказов пользователя
	readRange := "Cart!A:G"

	// Чтение данных из таблицы
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Context(ctx).Do()
	if err != nil {
		logger.Error.Println("Unable to retrieve data from sheet: %v", err)
		return err
	}

	// Поиск строки, соответствующей userID и позиции
	rowIndex := -1
	userOrderCount := 0
	for i, row := range resp.Values {
		if len(row) > 0 {
			id, err := strconv.ParseInt(row[0].(string), 10, 64)
			if err != nil {
				continue // Пропускаем строки с некорректным userID
			}

			if id == userID {
				if userOrderCount == position {
					rowIndex = i
					break
				}
				userOrderCount++
			}
		}
	}

	// Если строка не найдена, возвращаем ошибку
	if rowIndex == -1 {
		return fmt.Errorf("order not found for user %d at position %d", userID, position)
	}

	// Удаление строки
	deleteRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteDimension: &sheets.DeleteDimensionRequest{
					Range: &sheets.DimensionRange{
						SheetId:    0, // Предположим, что это первый лист (Cart)
						Dimension:  "ROWS",
						StartIndex: int64(rowIndex),
						EndIndex:   int64(rowIndex + 1),
					},
				},
			},
		},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, deleteRequest).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unable to delete row: %v", err)
	}

	logger.Debug.Printf("Product successfully deleted from cart!")
	return nil
}

func ClearCart(userID int64) error {
	srv, spreadsheetID, _ := InitSheet()

	// Определяем диапазон для чтения данных
	readRange := "Cart!A:D" // Замените на имя вашего листа и диапазон, если нужно

	// Читаем данные из таблицы
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		log.Println("No data found.")
		return nil
	}

	var rowsToKeep [][]interface{}
	var rowIndicesToDelete []int

	// Преобразуем userID в строку для сравнения
	userIDStr := strconv.FormatInt(userID, 10)

	// Обрабатываем каждую строку
	for i, row := range resp.Values {
		if len(row) > 0 {
			// Предполагается, что userID находится в первой колонке (индекс 0)
			cellValue, ok := row[0].(string)
			if !ok {
				continue
			}
			if cellValue != userIDStr {
				rowsToKeep = append(rowsToKeep, row)
			} else {
				rowIndicesToDelete = append(rowIndicesToDelete, i+1) // Строки индексы начинаются с 1 в Sheets API
			}
		}
	}

	// Если нет строк для удаления, ничего не делаем
	if len(rowIndicesToDelete) == 0 {
		logger.Debug.Println("No rows with the specified userID found.")
		return nil
	}

	// Пересоздаем таблицу без ненужных строк
	writeRange := "Cart!A1:D" // Убедитесь, что диапазон начинается с A1
	valueRange := &sheets.ValueRange{
		Values: rowsToKeep,
	}
	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, writeRange, valueRange).
		ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to update data in sheet: %v", err)
	}

	// Удаляем строки, начиная с последних, чтобы не нарушить индексы строк
	for i := len(rowIndicesToDelete) - 1; i >= 0; i-- {
		rowIndex := rowIndicesToDelete[i]
		deleteRange := &sheets.DeleteDimensionRequest{
			Range: &sheets.DimensionRange{
				SheetId:    0, // Укажите идентификатор листа, если не 0
				Dimension:  "ROWS",
				StartIndex: int64(rowIndex - 1), // Преобразуем в int64
				EndIndex:   int64(rowIndex),     // Преобразуем в int64
			},
		}
		_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					DeleteDimension: deleteRange,
				},
			},
		}).Do()
		if err != nil {
			return fmt.Errorf("unable to delete row %d: %v", rowIndex, err)
		}
	}

	return nil
}
