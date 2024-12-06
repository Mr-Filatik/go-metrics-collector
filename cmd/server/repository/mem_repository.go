package repository

import (
	"errors"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
)

type MemRepository struct {
	datas []entity.Metric
}

func New() *MemRepository {
	return &MemRepository{datas: make([]entity.Metric, 0)}
}

func (r *MemRepository) GetAll() []entity.Metric {
	if r.datas != nil {
		return r.datas
	}
	return make([]entity.Metric, 0)
}

func (r *MemRepository) Get(name string) (entity.Metric, error) {
	for _, v := range r.datas {
		if v.Name == name {
			return v, nil
		}
	}
	return entity.Metric{}, errors.New("metric not found")
}

func (r *MemRepository) Create(e entity.Metric) error {
	r.datas = append(r.datas, e)
	return nil
}

func (r *MemRepository) Update(e entity.Metric) error {
	for i, v := range r.datas {
		if v.Name == e.Name {
			item := &r.datas[i]
			item.Value = e.Value
			item.Type = e.Type
		}
	}
	return nil
}

func (r *MemRepository) Remove(e entity.Metric) error {
	newDatas := make([]entity.Metric, (len(r.datas) - 1))
	index := 0
	for i, v := range r.datas {
		if v.Name != e.Name {
			newDatas[index] = r.datas[i]
			index++
		} else {
			index++
		}
	}
	r.datas = newDatas
	return nil
}
