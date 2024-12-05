package storage

import (
	"errors"
	"log"
	"strconv"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/entity"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/repository/abstract"
)

type MemStorage struct {
	repository abstract.Repository
}

func New(r abstract.Repository) *MemStorage {

	return &MemStorage{repository: r}
}

func (s *MemStorage) GetAll() []entity.Metric {

	return s.repository.GetAll()
}

func (s *MemStorage) Get(t entity.MetricType, n string) (string, error) {

	if t != entity.Gauge && t != entity.Counter {
		return "", errors.New("invalid metric type")
	}

	m, err := s.repository.Get(n)
	if err != nil {
		return "", err
	}
	if t == m.Type {
		log.Printf("Get value: %v - %v.", n, m.Value)
		return m.Value, nil
	} else {
		return "", errors.New("invalid metric type")
	}
}

func (s *MemStorage) CreateOrUpdate(t entity.MetricType, n string, v string) error {

	if t != entity.Gauge && t != entity.Counter {
		return errors.New("invalid metric type")
	}

	m, err := s.repository.Get(n)
	if err != nil {
		s.repository.Create(entity.Metric{Name: n, Type: t, Value: v})
		log.Printf("Create value: %v - %v.", n, v)
	} else {
		if t == m.Type {
			return s.updateMetric(m, v)
		}
		return errors.New("invalid metric type")
	}
	return nil
}

func (s *MemStorage) updateMetric(currentMetric entity.Metric, newValue string) error {

	if currentMetric.Type == entity.Gauge {
		return s.updateGaugeMetric(currentMetric, newValue)
	}
	if currentMetric.Type == entity.Counter {
		return s.updateCounterMetric(currentMetric, newValue)
	}
	return nil
}

func (s *MemStorage) updateGaugeMetric(currentMetric entity.Metric, newValue string) error {

	if num, err := strconv.ParseFloat(newValue, 64); err == nil {
		newValue = strconv.FormatFloat(num, 'f', -1, 64)
		s.repository.Update(entity.Metric{Name: currentMetric.Name, Type: currentMetric.Type, Value: newValue})
		log.Printf("Update value: %v - %v to %v.", currentMetric.Name, currentMetric.Value, newValue)
		return nil
	}
	log.Printf("Invalid metric value: %v - %v.", currentMetric.Name, newValue)
	return errors.New("invalid metric value")
}

func (s *MemStorage) updateCounterMetric(currentMetric entity.Metric, newValue string) error {

	if nnum, err := strconv.ParseInt(newValue, 10, 64); err == nil {
		if newnum, err2 := strconv.ParseInt(currentMetric.Value, 10, 64); err2 == nil {
			newnum += nnum
			newValue = strconv.FormatInt(newnum, 10)
			s.repository.Update(entity.Metric{Name: currentMetric.Name, Type: currentMetric.Type, Value: newValue})
			log.Printf("Update value: %v - %v to %v.", currentMetric.Name, currentMetric.Value, newValue)
			return nil
		}
	}
	log.Printf("Invalid metric value: %v - %v.", currentMetric.Name, newValue)
	return errors.New("invalid metric value")
}