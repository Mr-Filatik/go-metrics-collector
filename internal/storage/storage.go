// Пакет storage предоставляет абстрактное описание интерфейса,
// которому должно соотвестовать любое хранилище проекта.
package storage

import "github.com/Mr-Filatik/go-metrics-collector/internal/entity"

// Storage описание интерфейса для реализации хранилища.
type Storage interface {
	LoadData() ([]entity.Metrics, error)  // загрузка данных
	SaveData(data []entity.Metrics) error // сохранение данных
}
