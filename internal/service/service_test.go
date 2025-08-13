package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	"github.com/Mr-Filatik/go-metrics-collector/internal/storage"
	"github.com/stretchr/testify/assert"
)

var _ repository.Repository = (*MockRepository)(nil)

// MockRepository — реализация Repository для тестов.
type MockRepository struct {
	GetByIDFunc func(ctx context.Context, id string) (entity.Metrics, error)
	GetAllFunc  func(ctx context.Context) ([]entity.Metrics, error)
	CreateFunc  func(ctx context.Context, e entity.Metrics) (string, error)
	UpdateFunc  func(ctx context.Context, e entity.Metrics) (float64, int64, error)
	RemoveFunc  func(ctx context.Context, e entity.Metrics) (string, error)
	PingFunc    func(ctx context.Context) error
}

func (m MockRepository) GetByID(ctx context.Context, id string) (entity.Metrics, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m MockRepository) Create(ctx context.Context, e entity.Metrics) (string, error) {
	return m.CreateFunc(ctx, e)
}

func (m MockRepository) Update(ctx context.Context, e entity.Metrics) (float64, int64, error) {
	return m.UpdateFunc(ctx, e)
}

func (m MockRepository) Remove(ctx context.Context, e entity.Metrics) (string, error) {
	return m.RemoveFunc(ctx, e)
}

func (m MockRepository) GetAll(ctx context.Context) ([]entity.Metrics, error) {
	return m.GetAllFunc(ctx)
}

func (m MockRepository) Ping(ctx context.Context) error {
	return m.PingFunc(ctx)
}

var _ storage.Storage = (*MockStorage)(nil)

// MockStorage — реализация Storage для тестов.
type MockStorage struct {
	SaveDataFunc func([]entity.Metrics) error
	LoadDataFunc func() ([]entity.Metrics, error)
}

func (m MockStorage) SaveData(data []entity.Metrics) error {
	if m.SaveDataFunc != nil {
		return m.SaveDataFunc(data)
	}
	return nil
}

func (m MockStorage) LoadData() ([]entity.Metrics, error) {
	if m.LoadDataFunc != nil {
		return m.LoadDataFunc()
	}
	return nil, nil
}

func TestNewService(t *testing.T) {
	log := logger.New(logger.LevelDebug)
	mockRepo := &MockRepository{}
	mockStore := &MockStorage{}

	service := New(mockRepo, mockStore, 5, log)

	assert.NotNil(t, service)
	assert.Equal(t, int64(5), service.storSaveInterval)
	assert.Equal(t, mockRepo, service.repository)
	assert.Equal(t, mockStore, service.storage)
}

func TestGetByID(t *testing.T) {
	log := logger.New(logger.LevelDebug)
	ctx := context.Background()

	mockRepo := MockRepository{
		GetByIDFunc: func(ctx context.Context, id string) (entity.Metrics, error) {
			if id == "valid" {
				val := 42.0
				delta := int64(10)
				return entity.Metrics{ID: id, MType: entity.Gauge, Value: &val, Delta: &delta}, nil
			}
			return entity.Metrics{}, errors.New(repository.ErrorMetricNotFound)
		},
	}

	t.Run("get valid gauge", func(t *testing.T) {
		s := New(&mockRepo, nil, 0, log)
		m, err := s.Get(ctx, "valid", entity.Gauge)
		assert.NoError(t, err)
		assert.Equal(t, "valid", m.ID)
	})

	t.Run("metric not found", func(t *testing.T) {
		s := New(&mockRepo, nil, 0, log)
		m, err := s.Get(ctx, "invalid", entity.Counter)
		assert.Error(t, err)
		assert.Equal(t, MetricNotFound, err.Error())
		assert.Empty(t, m)
	})
}

func TestGetAll(t *testing.T) {
	log := logger.New(logger.LevelDebug)
	ctx := context.Background()

	metricList := []entity.Metrics{
		{ID: "gauge1", MType: "gauge", Value: new(float64)},
		{ID: "counter1", MType: "counter", Delta: new(int64)},
	}

	mockRepo := MockRepository{
		GetAllFunc: func(ctx context.Context) ([]entity.Metrics, error) {
			return metricList, nil
		},
	}

	s := New(&mockRepo, nil, 0, log)
	result, err := s.GetAll(ctx)

	assert.NoError(t, err)
	assert.Equal(t, len(metricList), len(result))
	assert.Equal(t, result[0].ID, "gauge1")
	assert.Equal(t, result[1].MType, "counter")
}

func TestCreateOrUpdate(t *testing.T) {
	log := logger.New(logger.LevelDebug)
	ctx := context.Background()

	t.Run("create new metric", func(t *testing.T) {
		mockRepo := MockRepository{
			GetByIDFunc: func(ctx context.Context, id string) (entity.Metrics, error) {
				return entity.Metrics{}, errors.New(repository.ErrorMetricNotFound)
			},
			CreateFunc: func(ctx context.Context, e entity.Metrics) (string, error) {
				return e.ID, nil
			},
		}

		s := New(&mockRepo, nil, 0, log)
		e := entity.Metrics{ID: "new_metric", MType: entity.Gauge, Value: new(float64)}
		res, err := s.CreateOrUpdate(ctx, e)

		assert.NoError(t, err)
		assert.Equal(t, e.ID, res.ID)
	})

	t.Run("update counter", func(t *testing.T) {
		mockRepo := MockRepository{
			GetByIDFunc: func(ctx context.Context, id string) (entity.Metrics, error) {
				delta := int64(5)
				return entity.Metrics{ID: id, MType: entity.Counter, Delta: &delta}, nil
			},
			UpdateFunc: func(ctx context.Context, e entity.Metrics) (float64, int64, error) {
				del := *e.Delta
				return 0, del, nil
			},
		}

		s := New(&mockRepo, nil, 0, log)
		delta := int64(0)
		e := entity.Metrics{ID: "counter1", MType: entity.Counter, Delta: &delta}
		*e.Delta = 1

		res, err := s.CreateOrUpdate(ctx, e)

		assert.NoError(t, err)
		assert.Equal(t, int64(6), *res.Delta)
	})
}

func TestStart_AutoSave(t *testing.T) {
	log := logger.New(logger.LevelDebug)

	mockStore := MockStorage{
		SaveDataFunc: func(data []entity.Metrics) error {
			assert.NotEmpty(t, data)
			return nil
		},
	}

	mockRepo := MockRepository{
		GetAllFunc: func(ctx context.Context) ([]entity.Metrics, error) {
			return []entity.Metrics{{ID: "test", MType: entity.Gauge}}, nil
		},
	}

	s := New(&mockRepo, &mockStore, 1, log)
	s.Start(false)

	time.Sleep(2 * time.Second)
}

func TestPing_WithErrors(t *testing.T) {
	ctx := context.Background()

	mockRepo := MockRepository{
		PingFunc: func(ctx context.Context) error {
			return errors.New("can't ping repo")
		},
	}

	s := New(&mockRepo, nil, 0, logger.New(logger.LevelDebug))
	err := s.Ping(ctx)

	assert.Error(t, err)
}
