package abstract

import "github.com/Mr-Filatik/go-metrics-collector/cmd/server/entity"

type Repository interface {
	GetAll() []entity.Metric
	Get(name string) (entity.Metric, error)
	Create(e entity.Metric) error
	Update(e entity.Metric) error
	Remove(e entity.Metric) error
}
