package storage

import "github.com/Mr-Filatik/go-metrics-collector/internal/entity"

type Storage interface {
	LoadData() ([]entity.Metrics, error)
	SaveData(data []entity.Metrics) error
}
