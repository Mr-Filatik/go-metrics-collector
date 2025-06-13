// Пакет repository предоставляет конкретную реализацию репозитория
// для доступа к memory-хранилищу.
package repository

import (
	"errors"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
)

// MemoryRepository хранилище данных в оперативной памяти.
type MemoryRepository struct {
	log    logger.Logger    // логгер
	dbConn string           // строка подключения
	datas  []entity.Metrics // хранилище данных метрик
}

// New создаёт и инициализирует новый экзепляр *MemoryRepository.
//
// Параметры:
//   - dbConn: строка подключения к базе данных
//   - l: логгер
func New(dbConn string, l logger.Logger) *MemoryRepository {
	l.Info("Create MemoryRepository")

	return &MemoryRepository{
		datas:  make([]entity.Metrics, 0),
		log:    l,
		dbConn: dbConn,
	}
}

// Ping проверяет доступность и готовность репозитория.
func (r *MemoryRepository) Ping() error {
	return nil
}

// GetAll возвращает все хранящиеся метрики или ошибку.
func (r *MemoryRepository) GetAll() ([]entity.Metrics, error) {
	if r.datas != nil {
		r.log.Debug(
			"Query all metrics from MemRepository",
			"count", len(r.datas),
		)
		return r.datas, nil
	}
	r.log.Debug("Querying empty data in MemRepository")
	return make([]entity.Metrics, 0), nil
}

// GetByID возвращает метрику по идентификатору или ошибку.
//
// Параметры:
//   - id: идентификатор метрики
func (r *MemoryRepository) GetByID(id string) (entity.Metrics, error) {
	for _, v := range r.datas {
		if v.ID == id {
			r.log.Debug(
				"Getting metric from MemRepository",
				"id", v.ID,
				"type", v.MType,
				"value", v.Value,
				"delta", v.Delta,
			)

			return entity.Metrics{
				ID:    v.ID,
				MType: v.MType,
				Value: v.Value,
				Delta: v.Delta,
			}, nil
		}
	}
	return entity.Metrics{}, errors.New(repository.ErrorMetricNotFound)
}

// Create создаёт новую метрику или возвращает ошибку.
//
// Параметры:
//   - e: метрика
func (r *MemoryRepository) Create(e entity.Metrics) (string, error) {
	r.datas = append(r.datas, e)

	r.log.Debug(
		"Creating a new metric in MemRepository",
		"id", e.ID,
		"type", e.MType,
		"value", e.Value,
		"delta", e.Delta,
	)

	return e.ID, nil
}

// Update обновляет значение метрики или возвращает ошибку.
//
// Параметры:
//   - e: метрика
func (r *MemoryRepository) Update(e entity.Metrics) (float64, int64, error) {
	for i, v := range r.datas {
		if v.ID == e.ID {
			item := &r.datas[i]
			item.Value = e.Value
			item.MType = e.MType
			item.Delta = e.Delta

			r.log.Debug(
				"Updating metric data in MemRepository",
				"id", item.ID,
				"type", item.MType,
				"value", item.Value,
				"delta", item.Delta,
			)

			value := float64(0)
			if item.Value != nil {
				value = *item.Value
			}
			delta := int64(0)
			if item.Delta != nil {
				delta = *item.Delta
			}
			return value, delta, nil
		}
	}
	return 0, 0, errors.New(repository.ErrorMetricNotFound)
}

// Update удаляет метрику или возвращает ошибку.
//
// Параметры:
//   - e: метрика
func (r *MemoryRepository) Remove(e entity.Metrics) (string, error) {
	newDatas := make([]entity.Metrics, (len(r.datas) - 1))
	index := 0
	for i, v := range r.datas {
		if v.ID != e.ID {
			newDatas[index] = r.datas[i]
			index++
		} else {
			r.log.Debug(
				"Deleting a metric in MemRepository",
				"id", e.ID,
			)

			index++
		}
	}
	r.datas = newDatas
	return e.ID, nil
}
