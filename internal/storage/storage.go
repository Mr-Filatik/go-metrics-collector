package storage

import (
	"errors"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
)

type Repository interface {
	GetAll() ([]entity.Metrics, error)
	Get(id string) (entity.Metrics, error)
	Create(e entity.Metrics) (entity.Metrics, error)
	Update(e entity.Metrics) (entity.Metrics, error)
	Remove(e entity.Metrics) (entity.Metrics, error)
}

type Storage struct {
	repository Repository
	log        logger.Logger
}

const (
	ErrorMetricType        = "invalid metric type"
	ErrorMetricName        = "invalid metric name"
	ErrorMetricValue       = "invalid metric value"
	UnexpectedMetricCreate = "create error"
	UnexpectedMetricUpdate = "update error"
)

func New(r Repository, l logger.Logger) *Storage {
	return &Storage{
		repository: r,
		log:        l,
	}
}

func IsExpectedError(e error) bool {
	err := e.Error()
	return err == ErrorMetricName ||
		err == ErrorMetricType ||
		err == ErrorMetricValue ||
		err == repository.ErrorMetricNotFound
}

func (s *Storage) GetAll() ([]entity.Metrics, error) {
	return s.repository.GetAll()
}

func (s *Storage) Get(id string, t string) (entity.Metrics, error) {
	if t != entity.Gauge && t != entity.Counter {
		s.reportStorageError(ErrorMetricType, string(t))
		return entity.Metrics{}, errors.New(ErrorMetricType)
	}
	m, err := s.repository.Get(id)
	if err != nil {
		s.reportStorageError(err.Error(), "")
		return entity.Metrics{}, errors.New(err.Error())
	}
	if t != m.MType {
		s.reportStorageError(ErrorMetricType, string(t))
		return entity.Metrics{}, errors.New(ErrorMetricType)
	}
	s.log.Info(
		"Storage get value",
		"name", m.ID,
		"type", m.MType,
		"value", m.Value,
	)
	return m, nil
}

func (s *Storage) CreateOrUpdate(e entity.Metrics) (entity.Metrics, error) {
	if e.MType != entity.Gauge && e.MType != entity.Counter {
		s.reportStorageError(ErrorMetricType, string(e.MType))
		return entity.Metrics{}, errors.New(ErrorMetricType)
	}
	m, err := s.repository.Get(e.ID)
	if err != nil {
		im, iErr := s.repository.Create(e)
		if iErr != nil {
			s.reportStorageError(iErr.Error(), "")
			return entity.Metrics{}, errors.New(UnexpectedMetricCreate)
		}
		s.log.Info(
			"Storage create value",
			"name", im.ID,
			"type", im.MType,
			"value", im.Value,
			"delta", im.Delta,
		)
		return im, nil
	} else {
		if e.MType != m.MType {
			s.reportStorageError(ErrorMetricType, string(e.MType))
			return entity.Metrics{}, errors.New(ErrorMetricType)
		}
		if e.MType == entity.Gauge {
			im, iErr := s.repository.Update(e)
			if iErr != nil {
				s.reportStorageError(iErr.Error(), "")
				return entity.Metrics{}, errors.New(UnexpectedMetricUpdate)
			}
			s.log.Info(
				"Storage update value",
				"name", im.ID,
				"type", im.MType,
				"value", im.Value,
				"delta", im.Delta,
			)
			return im, nil
		}
		if e.MType == entity.Counter {
			val := *m.Delta + *e.Delta
			e.Delta = &val
			im, iErr := s.repository.Update(e)
			if iErr != nil {
				s.reportStorageError(iErr.Error(), "")
				return entity.Metrics{}, errors.New(UnexpectedMetricUpdate)
			}
			s.log.Info(
				"Storage update value",
				"name", im.ID,
				"type", im.MType,
				"value", im.Value,
				"delta", im.Delta,
			)
			return im, nil
		}
		s.reportStorageError(ErrorMetricType, string(e.MType))
		return entity.Metrics{}, errors.New(ErrorMetricType)
	}
}

func (s *Storage) reportStorageError(text string, value string) {
	if value == "" {
		s.log.Info(
			"Storage error",
			"error", text,
		)
	} else {
		s.log.Info(
			"Storage error",
			"error", text,
			"value", value,
		)
	}
}
