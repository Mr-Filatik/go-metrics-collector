// Пакет storage предоставляет абстрактное описание интерфейса,
// которому должно соотвестовать любое хранилище проекта.
package storage

import "github.com/Mr-Filatik/go-metrics-collector/internal/entity"

type Storage interface {
	LoadData() ([]entity.Metrics, error)
	SaveData(data []entity.Metrics) error
}
