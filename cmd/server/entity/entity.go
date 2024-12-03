package entity

type Metric struct {
	Type  MetricType
	Name  string
	Value string
}

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)
