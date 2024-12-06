package storage

import (
	"errors"
	"log"
	"strconv"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
)

type Repository interface {
	GetAll() []entity.Metric
	Get(name string) (entity.Metric, error)
	Create(e entity.Metric) error
	Update(e entity.Metric) error
	Remove(e entity.Metric) error
}

type Storage struct {
	repository Repository
}

func New(r Repository) *Storage {
	return &Storage{repository: r}
}

func (s *Storage) GetAll() []entity.Metric {
	return s.repository.GetAll()
}

func (s *Storage) Get(t entity.MetricType, n string) (string, error) {
	if t != entity.Gauge && t != entity.Counter {
		log.Printf("Mem storage error: %v - %v.", "invalid metric type", t)
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
		log.Printf("Mem storage error: %v - %v.", "invalid metric type", t)
		return "", errors.New("invalid metric type")
	}
}

func (s *Storage) CreateOrUpdate(t entity.MetricType, n string, v string) error {
	if t != entity.Gauge && t != entity.Counter {
		log.Printf("Mem storage error: %v - %v.", "invalid metric type", t)
		return errors.New("invalid metric type")
	}

	m, err := s.repository.Get(n)
	if err != nil {
		cErr := s.repository.Create(entity.Metric{Name: n, Type: t, Value: v})
		if cErr != nil {
			log.Printf("Mem storage error: %v - %v.", "create for name error", n)
			return errors.New("create for name error")
		}
		log.Printf("Create value: %v - %v.", n, v)
	} else {
		if t == m.Type {
			return s.updateMetric(m, v)
		}
		log.Printf("Mem storage error: %v - %v.", "invalid metric type", t)
		return errors.New("invalid metric type")
	}
	return nil
}

func (s *Storage) updateMetric(currentMetric entity.Metric, newValue string) error {
	if currentMetric.Type == entity.Gauge {
		return s.updateGaugeMetric(currentMetric, newValue)
	}
	if currentMetric.Type == entity.Counter {
		return s.updateCounterMetric(currentMetric, newValue)
	}
	return nil
}

func (s *Storage) updateGaugeMetric(currentMetric entity.Metric, newValue string) error {
	if num, err := strconv.ParseFloat(newValue, 64); err == nil {
		newValue = strconv.FormatFloat(num, 'f', -1, 64)
		uErr := s.repository.Update(entity.Metric{Name: currentMetric.Name, Type: currentMetric.Type, Value: newValue})
		if uErr != nil {
			log.Printf("Mem storage error: %v - %v.", "update for name error", currentMetric.Name)
			return errors.New("update for name error")
		}
		log.Printf("Update value: %v - %v to %v.", currentMetric.Name, currentMetric.Value, newValue)
		return nil
	}
	log.Printf("Mem storage error: %v %v - %v.", "invalid metric value", currentMetric.Name, newValue)
	return errors.New("invalid metric value")
}

func (s *Storage) updateCounterMetric(currentMetric entity.Metric, newValue string) error {
	if nnum, err := strconv.ParseInt(newValue, 10, 64); err == nil {
		if newnum, err2 := strconv.ParseInt(currentMetric.Value, 10, 64); err2 == nil {
			newnum += nnum
			newValue = strconv.FormatInt(newnum, 10)
			uErr := s.repository.Update(entity.Metric{Name: currentMetric.Name, Type: currentMetric.Type, Value: newValue})
			if uErr != nil {
				log.Printf("Mem storage error: %v - %v.", "update for name error", currentMetric.Name)
				return errors.New("update for name error")
			}
			log.Printf("Update value: %v - %v to %v.", currentMetric.Name, currentMetric.Value, newValue)
			return nil
		}
	}
	log.Printf("Mem storage error: %v %v - %v.", "invalid metric value", currentMetric.Name, newValue)
	return errors.New("invalid metric value")
}
