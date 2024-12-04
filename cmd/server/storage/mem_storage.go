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
		return "0", errors.New("invalid metric type")
	}

	m, err := s.repository.Get(n)
	if err != nil {
		return "0", err
	}
	if t == m.Type {
		log.Printf("Get value: %v - %v.", n, m.Value)
		return m.Value, nil
	} else {
		return "0", errors.New("invalid metric type")
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
			if t == entity.Gauge {
				if num, err := strconv.ParseFloat(v, 64); err == nil {
					v = strconv.FormatFloat(num, 'f', -1, 64)
					s.repository.Update(entity.Metric{Name: n, Type: t, Value: v})
					log.Printf("Update value: %v - %v to %v.", n, m.Value, v)
				} else {
					return errors.New("invalid metric value")
				}
			}
			if t == entity.Counter {
				if nnum, err := strconv.ParseInt(v, 10, 64); err == nil {
					if newnum, err2 := strconv.ParseInt(m.Value, 10, 64); err2 == nil {
						newnum += nnum
						v = strconv.FormatInt(newnum, 10)
						s.repository.Update(entity.Metric{Name: n, Type: t, Value: v})
						log.Printf("Update value: %v - %v to %v.", n, m.Value, v)
					}
				} else {
					return errors.New("invalid metric value")
				}
			}
		} else {
			return errors.New("invalid metric type")
		}
	}
	return nil
}
