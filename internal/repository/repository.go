// Пакет repository предоставляет абстрактное описание интерфейса,
// которому должнен соотвестовать любой репозиторий проекта.
package repository

import "github.com/Mr-Filatik/go-metrics-collector/internal/entity"

const (
	ErrorMetricNotFound = "metric not found"
)

type Repository interface {
	Ping() error
	GetAll() ([]entity.Metrics, error)
	GetByID(id string) (entity.Metrics, error)
	Create(e entity.Metrics) (string, error)
	Update(e entity.Metrics) (float64, int64, error)
	Remove(e entity.Metrics) (string, error)
}
