package abstract

import (
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/entity"
)

type Storage interface {
	GetAll() []entity.Metric
	Get(t entity.MetricType, n string) (string, error)
	CreateOrUpdate(t entity.MetricType, n string, v string) error
}
