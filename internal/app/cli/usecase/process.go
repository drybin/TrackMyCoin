package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/drybin/TrackMyCoin/internal/adapter/webapi"
	"github.com/drybin/TrackMyCoin/internal/app/cli/config"
	"github.com/drybin/TrackMyCoin/internal/domain/model"
	"google.golang.org/api/sheets/v4"
)

type IProcess interface {
	Process(ctx context.Context) error
}

type Process struct {
	googleSheets *webapi.GoogleSheets
	coinGecko    *webapi.CoinGecko
	config       *config.Config
}

func NewProcessUsecase(googleSheets *webapi.GoogleSheets, coinGecko *webapi.CoinGecko, config *config.Config) *Process {
	return &Process{
		googleSheets: googleSheets,
		coinGecko:    coinGecko,
		config:       config,
	}
}

func (u *Process) Process(ctx context.Context) error {
	log.Println("Hello")
	log.Println("Reading Google Sheets document...")

	if u.googleSheets == nil {
		log.Println("Google Sheets client is not initialized. Please set GOOGLE_API_KEY or GOOGLE_SERVICE_ACCOUNT_FILE in .env file")
		return fmt.Errorf("google Sheets client is not initialized")
	}

	spreadsheetID := u.config.GoogleSheetID

	log.Printf("Spreadsheet ID: %s\n", spreadsheetID)

	// Получаем информацию о таблице, включая названия листов
	log.Println("Getting spreadsheet info...")
	spreadsheet, err := u.googleSheets.GetSpreadsheetInfo(ctx, spreadsheetID)
	if err != nil {
		log.Printf("Error getting spreadsheet info: %v\n", err)
		return fmt.Errorf("failed to get spreadsheet info: %w", err)
	}

	log.Printf("Spreadsheet title: %s\n", spreadsheet.Properties.Title)
	log.Printf("Available sheets: %d\n", len(spreadsheet.Sheets))

	for i, sheet := range spreadsheet.Sheets {
		log.Printf("  Sheet %d: %s (ID: %d)\n", i+1, sheet.Properties.Title, sheet.Properties.SheetId)
	}

	// Определяем какой лист читать
	var sheetName string
	var readRange string

	if u.config.GoogleSheetRange != "" {
		// Если указан диапазон в конфиге, используем его
		readRange = u.config.GoogleSheetRange
	} else if len(spreadsheet.Sheets) > 0 {
		// Иначе читаем первый лист полностью
		sheetName = spreadsheet.Sheets[0].Properties.Title
		readRange = sheetName // Просто имя листа без диапазона читает весь лист
	} else {
		return fmt.Errorf("no sheets found in spreadsheet")
	}

	log.Printf("Reading range: %s\n", readRange)

	data, err := u.googleSheets.ReadSpreadsheet(ctx, spreadsheetID, readRange)
	if err != nil {
		log.Printf("Error reading spreadsheet: %v\n", err)
		return fmt.Errorf("failed to read spreadsheet: %w", err)
	}

	if len(data.Values) == 0 {
		log.Println("No data found in spreadsheet")
		return nil
	}

	log.Printf("Found %d rows\n", len(data.Values))

	// Первая строка - заголовки
	if len(data.Values) < 2 {
		log.Println("No data rows found (only headers)")
		return nil
	}

	log.Println("\nHeaders:")
	log.Println(data.Values[0])

	log.Println("\n======================")
	log.Println("Parsing coin price records:")
	log.Println("======================")

	var records []*model.CoinPriceRecord
	var parseErrors []string

	// Парсим строки начиная со второй (первая - заголовки)
	for i, row := range data.Values[1:] {
		rowNum := i + 2 // +2 потому что i начинается с 0 и пропускаем заголовок

		record, err := model.ParseFromRow(row)
		if err != nil {
			errMsg := fmt.Sprintf("Row %d: parse error: %v", rowNum, err)
			parseErrors = append(parseErrors, errMsg)
			log.Println(errMsg)
			continue
		}

		records = append(records, record)
		log.Printf("Row %d: %s\n", rowNum, record.String())
	}

	log.Println("\n======================")
	log.Printf("Successfully parsed: %d records\n", len(records))
	if len(parseErrors) > 0 {
		log.Printf("Parse errors: %d\n", len(parseErrors))
	}
	log.Println("======================")

	// Заполняем пустые цены через CoinGecko API
	if err := u.fillMissingPrices(ctx, records); err != nil {
		return fmt.Errorf("failed to fill missing prices: %w", err)
	}

	// Записываем обновленные данные обратно в Google Sheets
	if err := u.updateGoogleSheets(ctx, spreadsheet, records, readRange); err != nil {
		return fmt.Errorf("failed to update Google Sheets: %w", err)
	}

	log.Println("\n✅ Process completed successfully!")
	return nil
}

// fillMissingPrices заполняет пустые цены через CoinGecko API
func (u *Process) fillMissingPrices(ctx context.Context, records []*model.CoinPriceRecord) error {
	log.Println("\n======================")
	log.Println("Checking and filling missing prices...")
	log.Println("======================")

	now := time.Now()
	totalMissingPrices := 0
	totalUpdated := 0
	var priceErrors []string

	for i, record := range records {
		if record.Coin == "" {
			continue
		}

		recordNum := i + 1
		recordMissingCount := 0
		recordUpdatedCount := 0

		// Проверяем и заполняем Bybit цену
		if record.BybitPrice == 0 {
			recordMissingCount++
			log.Printf("Record %d (%s): Missing Bybit price, fetching from CoinGecko...\n", recordNum, record.Coin)

			price, err := u.coinGecko.GetCurrentPrice(ctx, record.Coin)
			if err != nil {
				errMsg := fmt.Sprintf("Record %d (%s): failed to get Bybit price: %v", recordNum, record.Coin, err)
				priceErrors = append(priceErrors, errMsg)
				log.Printf("  ❌ %s\n", errMsg)
			} else {
				record.BybitPrice = price
				recordUpdatedCount++
				log.Printf("  ✅ Updated Bybit price: $%.2f\n", price)
			}
		}

		// Проверяем и заполняем временные поля
		priceFields := record.GetPriceFields()
		for _, field := range priceFields {
			shouldFetch, err := record.ShouldFetchPrice(field, now)
			if err != nil {
				// Не можем распарсить дату/время, пропускаем эту запись
				continue
			}

			if shouldFetch {
				recordMissingCount++
				log.Printf("Record %d (%s): Missing %s, fetching from CoinGecko...\n", recordNum, record.Coin, field.Name)

				price, err := u.coinGecko.GetCurrentPrice(ctx, record.Coin)
				if err != nil {
					errMsg := fmt.Sprintf("Record %d (%s): failed to get %s: %v", recordNum, record.Coin, field.Name, err)
					priceErrors = append(priceErrors, errMsg)
					log.Printf("  ❌ %s\n", errMsg)
				} else {
					*field.Value = price
					recordUpdatedCount++
					log.Printf("  ✅ Updated %s: $%.2f\n", field.Name, price)
				}
			}
		}

		if recordMissingCount > 0 {
			log.Printf("Record %d (%s): filled %d/%d missing prices\n", recordNum, record.Coin, recordUpdatedCount, recordMissingCount)
		}

		totalMissingPrices += recordMissingCount
		totalUpdated += recordUpdatedCount
	}

	log.Println("\n======================")
	log.Println("Price filling summary:")
	log.Printf("Total missing prices found: %d\n", totalMissingPrices)
	log.Printf("Successfully updated: %d\n", totalUpdated)
	if len(priceErrors) > 0 {
		log.Printf("Failed to update: %d\n", len(priceErrors))
		log.Println("\nErrors:")
		for _, errMsg := range priceErrors {
			log.Printf("  - %s\n", errMsg)
		}
	}
	log.Println("======================")

	return nil
}

// updateGoogleSheets записывает обновленные данные обратно в Google Sheets
func (u *Process) updateGoogleSheets(ctx context.Context, spreadsheet *sheets.Spreadsheet, records []*model.CoinPriceRecord, readRange string) error {
	log.Println("\n======================")
	log.Println("Updating Google Sheets with new data...")
	log.Println("======================")

	if u.googleSheets == nil {
		log.Println("⚠️  Google Sheets client is not available. Skipping update.")
		return nil
	}

	if len(records) == 0 {
		log.Println("No records to update")
		return nil
	}

	// Конвертируем записи обратно в формат для Google Sheets
	var values [][]interface{}
	for _, record := range records {
		values = append(values, record.ToRow())
	}

	// Определяем имя листа из readRange или используем первый лист
	sheetName := readRange
	// Если readRange содержит "!", то берем часть до "!"
	if idx := len(readRange); idx > 0 {
		for i, ch := range readRange {
			if ch == '!' {
				sheetName = readRange[:i]
				break
			}
		}
	}

	// Если имя листа пустое или это весь диапазон, используем первый лист
	if sheetName == "" || sheetName == readRange {
		if len(spreadsheet.Sheets) > 0 {
			sheetName = spreadsheet.Sheets[0].Properties.Title
		}
	}

	// Формируем диапазон для записи: начинаем со строки 2 (после заголовков)
	// Строка 2 это первая строка данных, если у нас 10 записей, то последняя строка = 2 + 10 - 1 = 11
	firstDataRow := 2
	lastRow := firstDataRow + len(values) - 1
	writeRange := fmt.Sprintf("%s!A%d:R%d", sheetName, firstDataRow, lastRow)

	log.Printf("Sheet name: %s\n", sheetName)
	log.Printf("Writing %d records to range: %s (rows %d-%d)\n", len(records), writeRange, firstDataRow, lastRow)

	// Сначала очистим весь диапазон данных (все строки после заголовка), чтобы удалить старые данные
	// Это важно, если в таблице было больше строк, чем мы пишем сейчас
	clearRange := fmt.Sprintf("%s!A%d:R", sheetName, firstDataRow)
	log.Printf("Clearing old data in range: %s\n", clearRange)
	err := u.googleSheets.ClearSpreadsheet(ctx, u.config.GoogleSheetID, clearRange)
	if err != nil {
		log.Printf("⚠️  Warning: failed to clear old data: %v\n", err)
		// Продолжаем даже если очистка не удалась
	}

	// Записываем данные
	err = u.googleSheets.UpdateSpreadsheet(ctx, u.config.GoogleSheetID, writeRange, values)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	log.Println("✅ Successfully updated Google Sheets!")
	log.Printf("Updated %d rows in spreadsheet\n", len(values))
	log.Println("======================")

	return nil
}
