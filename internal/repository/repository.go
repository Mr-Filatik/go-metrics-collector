// Пакет repository предоставляет абстрактное описание интерфейса,
// которому должнен соотвестовать любой репозиторий проекта.
package repository

import (
	"context"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
)

// Константы - общие ошибки для репозиториев.
const (
	ErrorMetricNotFound = "metric not found" // ошибка, метрики не существует
)

type Repository interface {
	Ping(ctx context.Context) error
	GetAll(ctx context.Context) ([]entity.Metrics, error)
	GetByID(ctx context.Context, id string) (entity.Metrics, error)
	Create(ctx context.Context, e entity.Metrics) (string, error)
	Update(ctx context.Context, e entity.Metrics) (float64, int64, error)
	Remove(ctx context.Context, e entity.Metrics) (string, error)
}
