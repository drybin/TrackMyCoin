package model

import (
	"fmt"
	"strconv"
	"time"
)

// CoinPriceRecord представляет запись о цене монеты из Google Sheets
type CoinPriceRecord struct {
	Date         string  // Дата
	Time         string  // Время
	Source       string  // Источник
	Coin         string  // Монета
	Direction    string  // Направление
	SourcePrice  float64 // Цена в источнике
	BybitPrice   float64 // Цена на Bybit
	Price10Min   float64 // Цена через 10 минут
	Price30Min   float64 // Цена через 30 минут
	Price1Hour   float64 // Цена через 1 час
	Price2Hours  float64 // Цена через 2 часа
	Price6Hours  float64 // Цена через 6 часов
	Price12Hours float64 // Цена через 12 часов
	Price24Hours float64 // Цена через 24 часов
	Price3Days   float64 // Цена через 3 дня
	Price5Days   float64 // Цена через 5 дней
	Price7Days   float64 // Цена через 7 дней
	Price1Month  float64 // Цена через 1 месяц
}

// ParseFromRow парсит строку из Google Sheets в CoinPriceRecord
// Ожидаемый порядок колонок:
// Дата, Время, Источник, Монета, Направление, Цена в источнике, Цена на Bybit,
// Цена через 10 минут, Цена через 30 минут, Цена через 1 час, Цена через 2 часа,
// Цена через 6 часов, Цена через 12 часов, Цена через 24 часов,
// Цена через 3 дня, Цена через 5 дней, Цена через 7 дней, Цена через 1 месяц
func ParseFromRow(row []interface{}) (*CoinPriceRecord, error) {
	if len(row) < 5 {
		return nil, fmt.Errorf("invalid row: expected at least 5 columns, got %d", len(row))
	}

	record := &CoinPriceRecord{
		Date:      getStringValue(row, 0),
		Time:      getStringValue(row, 1),
		Source:    getStringValue(row, 2),
		Coin:      getStringValue(row, 3),
		Direction: getStringValue(row, 4),
	}

	// Парсим цены (могут быть пустыми)
	record.SourcePrice = getFloatValue(row, 5)
	record.BybitPrice = getFloatValue(row, 6)
	record.Price10Min = getFloatValue(row, 7)
	record.Price30Min = getFloatValue(row, 8)
	record.Price1Hour = getFloatValue(row, 9)
	record.Price2Hours = getFloatValue(row, 10)
	record.Price6Hours = getFloatValue(row, 11)
	record.Price12Hours = getFloatValue(row, 12)
	record.Price24Hours = getFloatValue(row, 13)
	record.Price3Days = getFloatValue(row, 14)
	record.Price5Days = getFloatValue(row, 15)
	record.Price7Days = getFloatValue(row, 16)
	record.Price1Month = getFloatValue(row, 17)

	return record, nil
}

// GetDateTime возвращает объединенную дату и время как строку
func (r *CoinPriceRecord) GetDateTime() string {
	return fmt.Sprintf("%s %s", r.Date, r.Time)
}

// TryParseDateTime пытается распарсить дату и время в time.Time
// Возвращает время в часовом поясе GMT+7
func (r *CoinPriceRecord) TryParseDateTime() (time.Time, error) {
	// Пробуем различные форматы
	formats := []string{
		"02.01.2006 15:04:05",
		"02.01.2006 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}

	dateTimeStr := r.GetDateTime()

	// GMT+7 часовой пояс
	location := time.FixedZone("GMT+7", 7*60*60)

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateTimeStr, location); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date/time: %s", dateTimeStr)
}

// PriceField представляет поле с ценой и соответствующий ему временной интервал
type PriceField struct {
	Name     string        // Название поля (например, "Price10Min")
	Duration time.Duration // Интервал времени
	Value    *float64      // Указатель на значение в структуре
}

// GetPriceFields возвращает список всех полей с ценами и их временными интервалами
func (r *CoinPriceRecord) GetPriceFields() []PriceField {
	return []PriceField{
		{Name: "Price10Min", Duration: 10 * time.Minute, Value: &r.Price10Min},
		{Name: "Price30Min", Duration: 30 * time.Minute, Value: &r.Price30Min},
		{Name: "Price1Hour", Duration: 1 * time.Hour, Value: &r.Price1Hour},
		{Name: "Price2Hours", Duration: 2 * time.Hour, Value: &r.Price2Hours},
		{Name: "Price6Hours", Duration: 6 * time.Hour, Value: &r.Price6Hours},
		{Name: "Price12Hours", Duration: 12 * time.Hour, Value: &r.Price12Hours},
		{Name: "Price24Hours", Duration: 24 * time.Hour, Value: &r.Price24Hours},
		{Name: "Price3Days", Duration: 3 * 24 * time.Hour, Value: &r.Price3Days},
		{Name: "Price5Days", Duration: 5 * 24 * time.Hour, Value: &r.Price5Days},
		{Name: "Price7Days", Duration: 7 * 24 * time.Hour, Value: &r.Price7Days},
		{Name: "Price1Month", Duration: 30 * 24 * time.Hour, Value: &r.Price1Month}, // Примерно 1 месяц
	}
}

// ShouldFetchPrice проверяет, нужно ли получать цену для указанного временного интервала
// Возвращает true, если время уже наступило и цена еще не заполнена
func (r *CoinPriceRecord) ShouldFetchPrice(field PriceField, now time.Time) (bool, error) {
	// Если цена уже заполнена, не нужно получать
	if *field.Value != 0 {
		return false, nil
	}

	// Парсим дату и время записи
	recordTime, err := r.TryParseDateTime()
	if err != nil {
		return false, err
	}

	// Рассчитываем время, когда должна быть эта цена
	targetTime := recordTime.Add(field.Duration)

	// Если текущее время больше или равно целевому времени, нужно получить цену
	return now.After(targetTime) || now.Equal(targetTime), nil
}

// Вспомогательные функции

func getStringValue(row []interface{}, index int) string {
	if index >= len(row) {
		return ""
	}
	if row[index] == nil {
		return ""
	}
	return fmt.Sprintf("%v", row[index])
}

func getFloatValue(row []interface{}, index int) float64 {
	if index >= len(row) {
		return 0
	}
	if row[index] == nil {
		return 0
	}

	str := fmt.Sprintf("%v", row[index])
	if str == "" {
		return 0
	}

	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return val
}

// String возвращает строковое представление записи
func (r *CoinPriceRecord) String() string {
	return fmt.Sprintf("Date: %s, Time: %s, Source: %s, Coin: %s, Direction: %s, SourcePrice: %.2f, BybitPrice: %.2f",
		r.Date, r.Time, r.Source, r.Coin, r.Direction, r.SourcePrice, r.BybitPrice)
}

// ToRow конвертирует запись обратно в формат строки для Google Sheets
func (r *CoinPriceRecord) ToRow() []interface{} {
	return []interface{}{
		r.Date,
		r.Time,
		r.Source,
		r.Coin,
		r.Direction,
		floatToInterface(r.SourcePrice),
		floatToInterface(r.BybitPrice),
		floatToInterface(r.Price10Min),
		floatToInterface(r.Price30Min),
		floatToInterface(r.Price1Hour),
		floatToInterface(r.Price2Hours),
		floatToInterface(r.Price6Hours),
		floatToInterface(r.Price12Hours),
		floatToInterface(r.Price24Hours),
		floatToInterface(r.Price3Days),
		floatToInterface(r.Price5Days),
		floatToInterface(r.Price7Days),
		floatToInterface(r.Price1Month),
	}
}

// floatToInterface конвертирует float64 в interface{} для записи в Google Sheets
// Если значение 0, возвращает пустую строку
func floatToInterface(val float64) interface{} {
	if val == 0 {
		return ""
	}
	return val
}
