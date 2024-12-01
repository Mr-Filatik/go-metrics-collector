package storages

import (
	"errors"
	"log"
	"strconv"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/analiticmetrics"
)

type Storage interface {
	//Get() StorageItem
	Update(t analiticmetrics.MetricType, n string, v string) error
	Create(t analiticmetrics.MetricType, n string, v string)
	Contains(t analiticmetrics.MetricType, n string) bool
}

type StorageItem struct {
	Type  *analiticmetrics.MetricType
	Name  *string
	Value *string
}

type MemStorage struct {
	Values []*StorageItem
}

func (s *MemStorage) Create(t analiticmetrics.MetricType, n string, v string) {

	switch t {
	case analiticmetrics.Gauge:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			v = strconv.FormatFloat(num, 'f', -1, 64)
		} else {
			v = "0"
		}
	case analiticmetrics.Counter:
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			v = strconv.FormatInt(num, 10)
		} else {
			v = "0"
		}
	}
	item := &StorageItem{&t, &n, &v}
	s.Values = append(s.Values, item)
	log.Printf("Add storage item: type: %v, name: %v, value: %v.", *item.Type, *item.Name, *item.Value)
}

func (s *MemStorage) Contains(t analiticmetrics.MetricType, n string) bool {
	for i := range s.Values {
		if *s.Values[i].Type == t && *s.Values[i].Name == n {
			return true
		}
	}
	return false
}

func (s *MemStorage) Update(t analiticmetrics.MetricType, n string, v string) error {
	for i := range s.Values {
		item := s.Values[i]
		if *item.Type == t && *item.Name == n {
			oldValue := *item.Value
			switch t {
			case analiticmetrics.Gauge:
				if num, err := strconv.ParseFloat(v, 64); err == nil {
					*item.Value = strconv.FormatFloat(num, 'f', -1, 64)
					log.Printf("Update storage item: type: %v, name: %v, value: %v (old value: %v).", *item.Type, *item.Name, *item.Value, oldValue)
					return nil
				} else {
					log.Printf("Update storage item error: Incorrect metric value.")
					return errors.New("incorrect metric value")
				}
			case analiticmetrics.Counter:
				if num, err := strconv.ParseInt(v, 10, 64); err == nil {
					if dat, err2 := strconv.ParseInt(*item.Value, 10, 64); err2 == nil {
						dat += num
						*item.Value = strconv.FormatInt(dat, 10)
						log.Printf("Update storage item: type: %v, name: %v, value: %v (old value: %v).", *item.Type, *item.Name, *item.Value, oldValue)
						return nil
					}
				} else {
					log.Printf("Update storage item error: Incorrect metric value.")
					return errors.New("incorrect metric value")
				}
			}
		}
	}
	return nil
}
