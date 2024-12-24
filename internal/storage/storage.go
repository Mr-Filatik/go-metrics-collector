package storage

import (
	"errors"
	"log"
	"strconv"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
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

const (
	ErrorMetricType        = "invalid metric type"
	ErrorMetricName        = "invalid metric name"
	ErrorMetricValue       = "invalid metric value"
	UnexpectedMetricCreate = "create error"
	UnexpectedMetricUpdate = "update error"
)

func New(r Repository) *Storage {
	return &Storage{repository: r}
}

func IsExpectedError(e error) bool {
	err := e.Error()
	return err == ErrorMetricName ||
		err == ErrorMetricType ||
		err == ErrorMetricValue ||
		err == repository.ErrorMetricNotFound
}

func (s *Storage) GetAll() []entity.Metric {
	return s.repository.GetAll()
}

func (s *Storage) Get(t entity.MetricType, n string) (string, error) {
	if t != entity.Gauge && t != entity.Counter {
		reportStorageError(ErrorMetricType, string(t))
		return "", errors.New(ErrorMetricType)
	}

	m, err := s.repository.Get(n)
	if err != nil {
		return "", errors.New(err.Error())
	}
	if t == m.Type {
		log.Printf("Get value: %v %v - %v.", t, n, m.Value)
		return m.Value, nil
	} else {
		reportStorageError(ErrorMetricType, string(t))
		return "", errors.New(ErrorMetricType)
	}
}

func (s *Storage) CreateOrUpdate(t entity.MetricType, n string, v string) error {
	if t != entity.Gauge && t != entity.Counter {
		reportStorageError(ErrorMetricType, string(t))
		return errors.New(ErrorMetricType)
	}

	m, err := s.repository.Get(n)
	if err != nil {
		cErr := s.repository.Create(entity.Metric{Name: n, Type: t, Value: v})
		if cErr != nil {
			reportStorageError(UnexpectedMetricCreate, n)
			return errors.New(UnexpectedMetricCreate)
		}
		log.Printf("Create value: %v %v - %v.", t, n, v)
	} else {
		if t == m.Type {
			return s.updateMetric(m, v)
		}
		reportStorageError(ErrorMetricType, string(t))
		return errors.New(ErrorMetricType)
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
			reportStorageError(UnexpectedMetricUpdate, currentMetric.Name)
			return errors.New(UnexpectedMetricUpdate)
		}
		log.Printf("Update value: %v - %v to %v.", currentMetric.Name, currentMetric.Value, newValue)
		return nil
	}
	log.Printf("Mem storage error: %v (value - %v).", ErrorMetricValue, newValue)
	return errors.New(ErrorMetricValue)
}

func (s *Storage) updateCounterMetric(currentMetric entity.Metric, newValue string) error {
	if nnum, err := strconv.ParseInt(newValue, 10, 64); err == nil {
		if newnum, err2 := strconv.ParseInt(currentMetric.Value, 10, 64); err2 == nil {
			newnum += nnum
			newValue = strconv.FormatInt(newnum, 10)
			uErr := s.repository.Update(entity.Metric{Name: currentMetric.Name, Type: currentMetric.Type, Value: newValue})
			if uErr != nil {
				reportStorageError(UnexpectedMetricUpdate, currentMetric.Name)
				return errors.New(UnexpectedMetricUpdate)
			}
			log.Printf("Update value: %v - %v to %v.", currentMetric.Name, currentMetric.Value, newValue)
			return nil
		}
	}
	log.Printf("Mem storage error: %v (value - %v).", ErrorMetricValue, newValue)
	return errors.New(ErrorMetricValue)
}

func reportStorageError(text string, value string) {
	if text == ErrorMetricName {
		log.Printf("Mem storage error: %v (name - %v).", text, value)
		return
	}
	if text == ErrorMetricType {
		log.Printf("Mem storage error: %v (type - %v).", text, value)
		return
	}
	if text == ErrorMetricValue {
		log.Printf("Mem storage error: %v (value - %v).", text, value)
		return
	}
	log.Printf("Mem storage error: %v (name - %v).", text, value)
}
