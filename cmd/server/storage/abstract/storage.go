package storageabstract

import (
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/entity"
)

type Storage interface {
	GetValue(t entity.MetricType, n string) *string
	GetAll() []*StorageItem
	Update(t entity.MetricType, n string, v string) error
	Create(t entity.MetricType, n string, v string)
	Contains(t entity.MetricType, n string) bool
}

type StorageItem struct {
	Type  *entity.MetricType
	Name  *string
	Value *string
}
