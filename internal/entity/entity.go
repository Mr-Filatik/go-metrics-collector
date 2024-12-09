package entity

type Metric struct {
	Type  MetricType `json:"type"`
	Name  string     `json:"name"`
	Value string     `json:"value"`
}

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)
