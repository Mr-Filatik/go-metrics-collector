// Пакет entity содержит общие сущности для проектов.
package entity

// Константы - типы метрик.
const (
	Gauge   string = "gauge"   // метрика gauge с заменяемым значением
	Counter string = "counter" // метрика counter с накопительным значением
)

// Metrics описывает метрики для их хранения и обработки.
type Metrics struct {
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
}
