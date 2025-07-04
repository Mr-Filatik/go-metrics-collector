// Пакет service предоставляет реализацию основной логики для серверного приложения.
package service

import (
	"errors"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	"github.com/Mr-Filatik/go-metrics-collector/internal/storage"
)

// Service представляет основную логику приложения.
// Использует или репозиторий или хранилище для хранения данных.
type Service struct {
	repository       repository.Repository // репозиторий
	storage          storage.Storage       // хранилище
	log              logger.Logger         // логгер
	storSaveInterval int64                 // интервал сохранения данных (в секундах)
}

// Константы - основные ошибки сервиса.
const (
	MetricUncorrect        = "invalid metric"               // ошибка, некорректная метрика
	MetricNotFound         = repository.ErrorMetricNotFound // ошибка, метрика не найдена
	UnexpectedMetricCreate = "create error"                 // ошибка создания метрики
	UnexpectedMetricUpdate = "update error"                 // ошибка обновления значения метрики
)

// New создаёт и инициализирует новый экзепляр *Service.
//
// Параметры:
//   - r: репозиторий с данными
//   - s: хранилище
//   - strInterval: интервал сохранения данных (в секундах)
//   - l: логгер
func New(r repository.Repository, s storage.Storage, strInterval int64, l logger.Logger) *Service {
	srvc := Service{
		repository:       r,
		storage:          s,
		log:              l,
		storSaveInterval: strInterval,
	}

	return &srvc
}

// Start запускает основную логику приложения.
//
// Параметры:
//   - restoreData: флаг, указывающий загружать ли данные при старте
func (s *Service) Start(restoreData bool) {
	if s.storage != nil && restoreData {
		data, serr := s.storage.LoadData()
		if serr != nil {
			s.log.Error("Load data from storage error", serr)
		}
		for _, val := range data {
			_, err := s.CreateOrUpdate(val)
			if err != nil {
				s.log.Error("Set data to repository error", err)
			}
		}
		s.log.Info(
			"Restore data from storage to repository is success",
			"time", time.Now(),
		)
	}

	if s.storage != nil && s.storSaveInterval != 0 {
		go s.autoSaveDataWithInterval(s.storSaveInterval)
	}
}

// Stop останавливает основную логику приложения.
func (s *Service) Stop() {
	err := s.saveDataWithoutInterval()
	if err != nil {
		s.log.Error("Stop service error", err)
	}
}

// Ping проверяет доступность логики.
func (s *Service) Ping() error {
	err := s.repository.Ping()
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
}

// GetAll возвращает все хранящиеся метрики.
func (s *Service) GetAll() ([]entity.Metrics, error) {
	vals, err := s.repository.GetAll()
	if err != nil {
		return make([]entity.Metrics, 0), errors.New(err.Error())
	}
	return vals, nil
}

// Get получает одну метрику.
//
// Параметры:
//   - id: идентификатор метрики
//   - t: тип метрики
func (s *Service) Get(id string, t string) (entity.Metrics, error) {
	m, err := s.repository.GetByID(id)
	if err != nil {
		s.reportStorageError(err.Error(), "")
		return entity.Metrics{}, errors.New(err.Error())
	}
	if t != m.MType {
		s.reportStorageError(MetricUncorrect, t)
		return entity.Metrics{}, errors.New(MetricUncorrect)
	}
	s.reportMetricInfo("Storage get value", m)
	return m, nil
}

// CreateOrUpdate обновляет значение метрики.
// Если метрика не была создана - создаёт её.
//
// Параметры:
//   - e: метрика
func (s *Service) CreateOrUpdate(e entity.Metrics) (entity.Metrics, error) {
	m, err := s.repository.GetByID(e.ID)
	if err != nil {
		_, iErr := s.repository.Create(e)
		if iErr != nil {
			s.reportStorageError(iErr.Error(), "")
			return entity.Metrics{}, errors.New(UnexpectedMetricCreate)
		}
		s.reportMetricInfo("Storage create value", e)
		return e, nil
	} else {
		if e.MType != m.MType {
			s.reportStorageError(MetricUncorrect, e.MType)
			return entity.Metrics{}, errors.New(MetricUncorrect)
		}
		if e.MType == entity.Gauge {
			ival, idel, iErr := s.repository.Update(e)
			if iErr != nil {
				s.reportStorageError(iErr.Error(), "")
				return entity.Metrics{}, errors.New(UnexpectedMetricUpdate)
			}

			if s.storage != nil && s.storSaveInterval == 0 {
				err := s.saveDataWithoutInterval()
				if err != nil {
					return entity.Metrics{}, errors.New(UnexpectedMetricUpdate)
				}
			}

			e.Value = &ival
			e.Delta = &idel
			s.reportMetricInfo("Storage update value", e)
			return e, nil
		}
		if e.MType == entity.Counter {
			val := *m.Delta + *e.Delta
			e.Delta = &val
			ival, idel, iErr := s.repository.Update(e)
			if iErr != nil {
				s.reportStorageError(iErr.Error(), "")
				return entity.Metrics{}, errors.New(UnexpectedMetricUpdate)
			}

			if s.storage != nil && s.storSaveInterval == 0 {
				err := s.saveDataWithoutInterval()
				if err != nil {
					return entity.Metrics{}, errors.New(UnexpectedMetricUpdate)
				}
			}

			e.Value = &ival
			e.Delta = &idel
			s.reportMetricInfo("Storage update value", e)
			return e, nil
		}
		s.reportStorageError(MetricUncorrect, e.MType)
		return entity.Metrics{}, errors.New(MetricUncorrect)
	}
}

func (s *Service) autoSaveDataWithInterval(interval int64) {
	t := time.Tick(time.Duration(interval) * time.Second)

	for range t {
		data, rerr := s.repository.GetAll()
		if rerr != nil {
			s.log.Error("Get data from repository error", rerr)
		}
		if s.storage != nil {
			serr := s.storage.SaveData(data)
			if serr != nil {
				s.log.Error("Auto save data to storage error", serr)
			}
			s.log.Info(
				"Auto save data to storage is success",
				"time", time.Now(),
			)
		}
	}
}

func (s *Service) saveDataWithoutInterval() error {
	data, rerr := s.repository.GetAll()
	if rerr != nil {
		s.log.Error("Save data to storage error", rerr)
		return errors.New(UnexpectedMetricUpdate)
	}
	if s.storage != nil {
		serr := s.storage.SaveData(data)
		if serr != nil {
			s.log.Error("Save data to storage error", serr)
			return errors.New(UnexpectedMetricUpdate)
		}
		s.log.Info(
			"Save data to storage is success",
			"time", time.Now(),
		)
	}
	return nil
}

func (s *Service) reportStorageError(text string, value string) {
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

func (s *Service) reportMetricInfo(t string, m entity.Metrics) {
	s.log.Debug(
		t,
		"name", m.ID,
		"type", m.MType,
		"value", m.Value,
		"delta", m.Delta,
	)
}
