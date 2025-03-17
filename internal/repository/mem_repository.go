package repository

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
)

type MemRepository struct {
	datas []entity.Metrics
}

const (
	ErrorMetricNotFound             = "metric not found"
	filePermission      os.FileMode = 0o600
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

func (r *MemRepository) LoadData(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return errors.New("failed to read metrics from file")
	}

	var metrics []entity.Metrics
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return errors.New("failed to deserialize metrics")
	}

	r.datas = metrics
	return nil
}

func (r *MemRepository) SaveData(filePath string) error {
	data, err := json.MarshalIndent(r.datas, "", "  ")
	if err != nil {
		return errors.New("failed to serialize metrics")
	}

	err = os.WriteFile(filePath, data, filePermission)
	if err != nil {
		return errors.New("failed to write metrics to file")
	}

	return nil
}
