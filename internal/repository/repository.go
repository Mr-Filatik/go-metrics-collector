package repository

import "github.com/Mr-Filatik/go-metrics-collector/internal/entity"

const (
	ErrorMetricNotFound = "metric not found"
)

type Repository interface {
	Ping() error
	GetAll() ([]entity.Metrics, error)
	GetByID(id string) (entity.Metrics, error)
	Create(e entity.Metrics) (entity.Metrics, error)
	Update(e entity.Metrics) (entity.Metrics, error)
	Remove(e entity.Metrics) (entity.Metrics, error)
}
