package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseFromRow(t *testing.T) {
	t.Run("Полная строка с данными", func(t *testing.T) {
		row := []interface{}{
			"29.12.2025",        // Дата
			"10:30:00",          // Время
			"Binance",           // Источник
			"BTC",               // Монета
			"UP",                // Направление
			"45000.50",          // Цена в источнике
			"45010.00",          // Цена на Bybit
			"45100.00",          // 10 мин
			"45200.00",          // 30 мин
			"45300.00",          // 1 час
			"45400.00",          // 2 часа
			"45500.00",          // 6 часов
			"45600.00",          // 12 часов
			"45700.00",          // 24 часа
			"46000.00",          // 3 дня
			"46500.00",          // 5 дней
			"47000.00",          // 7 дней
			"48000.00",          // 1 месяц
		}

		record, err := ParseFromRow(row)
		assert.NoError(t, err)
		assert.NotNil(t, record)

		assert.Equal(t, "29.12.2025", record.Date)
		assert.Equal(t, "10:30:00", record.Time)
		assert.Equal(t, "Binance", record.Source)
		assert.Equal(t, "BTC", record.Coin)
		assert.Equal(t, "UP", record.Direction)
		assert.Equal(t, 45000.50, record.SourcePrice)
		assert.Equal(t, 45010.00, record.BybitPrice)
		assert.Equal(t, 45100.00, record.Price10Min)
		assert.Equal(t, 48000.00, record.Price1Month)
	})

	t.Run("Строка с пустыми ценами", func(t *testing.T) {
		row := []interface{}{
			"29.12.2025",
			"10:30:00",
			"Binance",
			"BTC",
			"UP",
			"45000.50",
			"",
			"",
			"",
		}

		record, err := ParseFromRow(row)
		assert.NoError(t, err)
		assert.NotNil(t, record)

		assert.Equal(t, "BTC", record.Coin)
		assert.Equal(t, 45000.50, record.SourcePrice)
		assert.Equal(t, 0.0, record.BybitPrice)
		assert.Equal(t, 0.0, record.Price10Min)
	})

	t.Run("Минимальная строка", func(t *testing.T) {
		row := []interface{}{
			"29.12.2025",
			"10:30:00",
			"Binance",
			"BTC",
			"UP",
		}

		record, err := ParseFromRow(row)
		assert.NoError(t, err)
		assert.NotNil(t, record)

		assert.Equal(t, "BTC", record.Coin)
		assert.Equal(t, 0.0, record.SourcePrice)
	})

	t.Run("Недостаточно колонок", func(t *testing.T) {
		row := []interface{}{
			"29.12.2025",
			"10:30:00",
		}

		record, err := ParseFromRow(row)
		assert.Error(t, err)
		assert.Nil(t, record)
	})
}

func TestCoinPriceRecord_GetDateTime(t *testing.T) {
	record := &CoinPriceRecord{
		Date: "29.12.2025",
		Time: "10:30:00",
	}

	assert.Equal(t, "29.12.2025 10:30:00", record.GetDateTime())
}

func TestCoinPriceRecord_String(t *testing.T) {
	record := &CoinPriceRecord{
		Date:        "29.12.2025",
		Time:        "10:30:00",
		Source:      "Binance",
		Coin:        "BTC",
		Direction:   "UP",
		SourcePrice: 45000.50,
		BybitPrice:  45010.00,
	}

	str := record.String()
	assert.Contains(t, str, "BTC")
	assert.Contains(t, str, "Binance")
	assert.Contains(t, str, "45000.50")
}

func TestCoinPriceRecord_TryParseDateTime(t *testing.T) {
	t.Run("Успешный парсинг с GMT+7", func(t *testing.T) {
		record := &CoinPriceRecord{
			Date: "29.12.2025",
			Time: "10:30:00",
		}

		parsedTime, err := record.TryParseDateTime()
		assert.NoError(t, err)
		assert.Equal(t, 29, parsedTime.Day())
		assert.Equal(t, 12, int(parsedTime.Month()))
		assert.Equal(t, 2025, parsedTime.Year())
		assert.Equal(t, 10, parsedTime.Hour())
		assert.Equal(t, 30, parsedTime.Minute())
	})

	t.Run("Неправильный формат даты", func(t *testing.T) {
		record := &CoinPriceRecord{
			Date: "invalid",
			Time: "date",
		}

		_, err := record.TryParseDateTime()
		assert.Error(t, err)
	})
}

func TestCoinPriceRecord_GetPriceFields(t *testing.T) {
	record := &CoinPriceRecord{
		Price10Min:  100.0,
		Price30Min:  200.0,
		Price1Hour:  300.0,
		Price1Month: 1000.0,
	}

	fields := record.GetPriceFields()
	assert.Equal(t, 11, len(fields))

	// Проверяем первое поле
	assert.Equal(t, "Price10Min", fields[0].Name)
	assert.Equal(t, 10*time.Minute, fields[0].Duration)
	assert.Equal(t, 100.0, *fields[0].Value)

	// Проверяем последнее поле
	assert.Equal(t, "Price1Month", fields[10].Name)
	assert.Equal(t, 30*24*time.Hour, fields[10].Duration)
	assert.Equal(t, 1000.0, *fields[10].Value)
}

func TestCoinPriceRecord_ShouldFetchPrice(t *testing.T) {
	t.Run("Цена уже заполнена - не нужно получать", func(t *testing.T) {
		record := &CoinPriceRecord{
			Date:       "29.12.2025",
			Time:       "10:00:00",
			Price10Min: 45000.0,
		}

		fields := record.GetPriceFields()
		field := fields[0] // Price10Min

		// Время прошло, но цена уже есть
		now := time.Date(2025, 12, 29, 11, 0, 0, 0, time.FixedZone("GMT+7", 7*60*60))
		shouldFetch, err := record.ShouldFetchPrice(field, now)

		assert.NoError(t, err)
		assert.False(t, shouldFetch)
	})

	t.Run("Время наступило, цена пустая - нужно получить", func(t *testing.T) {
		record := &CoinPriceRecord{
			Date:       "29.12.2025",
			Time:       "10:00:00",
			Price10Min: 0, // Пустая
		}

		fields := record.GetPriceFields()
		field := fields[0] // Price10Min

		// Прошло 15 минут (больше чем 10)
		now := time.Date(2025, 12, 29, 10, 15, 0, 0, time.FixedZone("GMT+7", 7*60*60))
		shouldFetch, err := record.ShouldFetchPrice(field, now)

		assert.NoError(t, err)
		assert.True(t, shouldFetch)
	})

	t.Run("Время еще не наступило - не нужно получать", func(t *testing.T) {
		record := &CoinPriceRecord{
			Date:       "29.12.2025",
			Time:       "10:00:00",
			Price10Min: 0, // Пустая
		}

		fields := record.GetPriceFields()
		field := fields[0] // Price10Min

		// Прошло только 5 минут (меньше чем 10)
		now := time.Date(2025, 12, 29, 10, 5, 0, 0, time.FixedZone("GMT+7", 7*60*60))
		shouldFetch, err := record.ShouldFetchPrice(field, now)

		assert.NoError(t, err)
		assert.False(t, shouldFetch)
	})
}

func TestCoinPriceRecord_ToRow(t *testing.T) {
	t.Run("Конвертация полной записи", func(t *testing.T) {
		record := &CoinPriceRecord{
			Date:         "29.12.2025",
			Time:         "10:30:00",
			Source:       "Binance",
			Coin:         "BTC",
			Direction:    "UP",
			SourcePrice:  45000.50,
			BybitPrice:   45010.00,
			Price10Min:   45100.00,
			Price30Min:   45200.00,
			Price1Hour:   45300.00,
			Price2Hours:  45400.00,
			Price6Hours:  45500.00,
			Price12Hours: 45600.00,
			Price24Hours: 45700.00,
			Price3Days:   46000.00,
			Price5Days:   46500.00,
			Price7Days:   47000.00,
			Price1Month:  48000.00,
		}

		row := record.ToRow()

		assert.Equal(t, 18, len(row))
		assert.Equal(t, "29.12.2025", row[0])
		assert.Equal(t, "10:30:00", row[1])
		assert.Equal(t, "Binance", row[2])
		assert.Equal(t, "BTC", row[3])
		assert.Equal(t, "UP", row[4])
		assert.Equal(t, 45000.50, row[5])
		assert.Equal(t, 45010.00, row[6])
		assert.Equal(t, 45100.00, row[7])
		assert.Equal(t, 48000.00, row[17])
	})

	t.Run("Конвертация записи с пустыми ценами", func(t *testing.T) {
		record := &CoinPriceRecord{
			Date:        "29.12.2025",
			Time:        "10:30:00",
			Source:      "Binance",
			Coin:        "BTC",
			Direction:   "UP",
			SourcePrice: 45000.50,
			BybitPrice:  45010.00,
			// Остальные цены = 0 (пустые)
		}

		row := record.ToRow()

		assert.Equal(t, 18, len(row))
		assert.Equal(t, "BTC", row[3])
		assert.Equal(t, 45010.00, row[6])
		// Пустые цены должны быть пустыми строками
		assert.Equal(t, "", row[7])  // Price10Min
		assert.Equal(t, "", row[8])  // Price30Min
		assert.Equal(t, "", row[17]) // Price1Month
	})
}

func TestFloatToInterface(t *testing.T) {
	t.Run("Ненулевое значение", func(t *testing.T) {
		result := floatToInterface(45000.50)
		assert.Equal(t, 45000.50, result)
	})

	t.Run("Нулевое значение", func(t *testing.T) {
		result := floatToInterface(0)
		assert.Equal(t, "", result)
	})
}

