package repository

import (
	"errors"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
)

type MemRepository struct {
	datas []entity.Metrics
}

const (
	ErrorMetricNotFound = "metric not found"
)

func New() *MemRepository {
	return &MemRepository{datas: make([]entity.Metrics, 0)}
}

func (r *MemRepository) GetAll() ([]entity.Metrics, error) {
	if r.datas != nil {
		return r.datas, nil
	}
	return make([]entity.Metrics, 0), nil
}

func (r *MemRepository) Get(id string) (entity.Metrics, error) {
	for _, v := range r.datas {
		if v.ID == id {
			return entity.Metrics{
				ID:    v.ID,
				MType: v.MType,
				Value: v.Value,
				Delta: v.Delta,
			}, nil
		}
	}
	return entity.Metrics{}, errors.New(ErrorMetricNotFound)
}

func (r *MemRepository) Create(e entity.Metrics) (entity.Metrics, error) {
	r.datas = append(r.datas, e)
	return entity.Metrics{
		ID:    e.ID,
		MType: e.MType,
		Value: e.Value,
		Delta: e.Delta,
	}, nil
}

func (r *MemRepository) Update(e entity.Metrics) (entity.Metrics, error) {
	for i, v := range r.datas {
		if v.ID == e.ID {
			item := &r.datas[i]
			item.Value = e.Value
			item.MType = e.MType
			item.Delta = e.Delta

			return entity.Metrics{
				ID:    item.ID,
				MType: item.MType,
				Value: item.Value,
				Delta: item.Delta,
			}, nil
		}
	}
	return entity.Metrics{}, errors.New(ErrorMetricNotFound)
}

func (r *MemRepository) Remove(e entity.Metrics) (entity.Metrics, error) {
	newDatas := make([]entity.Metrics, (len(r.datas) - 1))
	index := 0
	for i, v := range r.datas {
		if v.ID != e.ID {
			newDatas[index] = r.datas[i]
			index++
		} else {
			index++
		}
	}
	r.datas = newDatas
	return e, nil
}
