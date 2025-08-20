package repository

import (
	"context"
	"testing"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	"github.com/Mr-Filatik/go-metrics-collector/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	mockLog := &testutil.MockLogger{}

	repo := New("test-conn", mockLog)

	require.NotNil(t, repo)
	assert.Equal(t, "test-conn", repo.dbConn)
	assert.Empty(t, repo.datas)
	log := mockLog.GetLastLog()
	assert.Contains(t, log.Message, "Create MemoryRepository")
}

func TestPing(t *testing.T) {
	repo := &MemoryRepository{}
	ctx := context.Background()
	err := repo.Ping(ctx)
	assert.NoError(t, err)
}

func TestGetAll_Empty(t *testing.T) {
	repo := &MemoryRepository{
		datas: make([]entity.Metrics, 0),
		log:   &testutil.MockLogger{},
	}
	ctx := context.Background()

	result, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetAll_WithData(t *testing.T) {
	testData := []entity.Metrics{
		{ID: "metric1", MType: "gauge", Value: floatPtr(1.5)},
		{ID: "metric2", MType: "counter", Delta: intPtr(10)},
	}
	ctx := context.Background()

	mockLog := &testutil.MockLogger{}
	repo := &MemoryRepository{
		datas: testData,
		log:   mockLog,
	}

	result, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, testData, result)
	log := mockLog.GetLastLog()
	assert.Contains(t, log.Message, "Query all metrics from MemRepository")
}

func TestGetByID_Found(t *testing.T) {
	testData := []entity.Metrics{
		{ID: "metric1", MType: "gauge", Value: floatPtr(1.5)},
	}
	ctx := context.Background()

	mockLog := &testutil.MockLogger{}
	repo := &MemoryRepository{
		datas: testData,
		log:   mockLog,
	}

	result, err := repo.GetByID(ctx, "metric1")
	assert.NoError(t, err)
	assert.Equal(t, "metric1", result.ID)
	assert.Equal(t, "gauge", result.MType)
	assert.Equal(t, floatPtr(1.5), result.Value)
	log := mockLog.GetLastLog()
	assert.Contains(t, log.Message, "Getting metric from MemRepository")
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &MemoryRepository{
		datas: []entity.Metrics{},
		log:   &testutil.MockLogger{},
	}
	ctx := context.Background()

	result, err := repo.GetByID(ctx, "unknown")
	assert.Error(t, err)
	assert.Equal(t, entity.Metrics{}, result)
	assert.Equal(t, err.Error(), repository.ErrorMetricNotFound)
}

func TestCreate(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	repo := &MemoryRepository{
		datas: make([]entity.Metrics, 0),
		log:   mockLog,
	}
	ctx := context.Background()

	newMetric := entity.Metrics{
		ID:    "new_metric",
		MType: "counter",
		Delta: intPtr(5),
	}

	id, err := repo.Create(ctx, newMetric)
	assert.NoError(t, err)
	assert.Equal(t, "new_metric", id)
	assert.Len(t, repo.datas, 1)
	assert.Equal(t, newMetric, repo.datas[0])
	log := mockLog.GetLastLog()
	assert.Contains(t, log.Message, "Creating a new metric in MemRepository")
}

func TestUpdate_Existing(t *testing.T) {
	existing := entity.Metrics{
		ID:    "metric1",
		MType: "gauge",
		Value: floatPtr(1.0),
	}
	repo := &MemoryRepository{
		datas: []entity.Metrics{existing},
		log:   &testutil.MockLogger{},
	}

	updated := entity.Metrics{
		ID:    "metric1",
		MType: "gauge",
		Value: floatPtr(2.5),
	}

	ctx := context.Background()
	value, delta, err := repo.Update(ctx, updated)
	assert.NoError(t, err)
	assert.Equal(t, 2.5, value)
	assert.Equal(t, int64(0), delta)

	current := repo.datas[0]
	assert.Equal(t, floatPtr(2.5), current.Value)
}

func TestUpdate_NotFound(t *testing.T) {
	repo := &MemoryRepository{
		datas: []entity.Metrics{},
		log:   &testutil.MockLogger{},
	}
	ctx := context.Background()

	metric := entity.Metrics{ID: "unknown", MType: "gauge", Value: floatPtr(1.0)}
	value, delta, err := repo.Update(ctx, metric)

	assert.Error(t, err)
	assert.Equal(t, 0.0, value)
	assert.Equal(t, int64(0), delta)
	assert.Equal(t, err.Error(), repository.ErrorMetricNotFound)
}

func TestRemove(t *testing.T) {
	repo := &MemoryRepository{
		datas: []entity.Metrics{
			{ID: "metric1", MType: "gauge", Value: floatPtr(1.0)},
			{ID: "metric2", MType: "counter", Delta: intPtr(10)},
		},
		log: &testutil.MockLogger{},
	}

	ctx := context.Background()
	deletedID, err := repo.Remove(ctx, entity.Metrics{ID: "metric1"})
	assert.NoError(t, err)
	assert.Equal(t, "metric1", deletedID)
	assert.Len(t, repo.datas, 1)
	assert.Equal(t, "metric2", repo.datas[0].ID)
}

func TestRemove_NonExistent(t *testing.T) {
	repo := &MemoryRepository{
		datas: []entity.Metrics{
			{ID: "metric1", MType: "gauge", Value: floatPtr(1.0)},
		},
		log: &testutil.MockLogger{},
	}

	ctx := context.Background()
	deletedID, err := repo.Remove(ctx, entity.Metrics{ID: "unknown"})
	assert.Error(t, err)
	assert.Equal(t, "", deletedID)
	assert.Len(t, repo.datas, 1)
}

func floatPtr(f float64) *float64 { return &f }
func intPtr(i int64) *int64       { return &i }
