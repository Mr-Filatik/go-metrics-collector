package storage

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempFile(t *testing.T, content []byte) string {
	t.Helper()
	file, err := os.CreateTemp("", "metrics_*.json")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := os.Remove(file.Name()); err != nil {
			assert.NoError(t, err)
		}
	})
	if content != nil {
		_, err = file.Write(content)
		require.NoError(t, err)
	}
	if err := file.Close(); err != nil {
		assert.NoError(t, err)
	}
	return file.Name()
}

func floatPtr(f float64) *float64 { return &f }
func intPtr(i int64) *int64       { return &i }

func TestNew(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	storage := New("/tmp/metrics.json", mockLog)

	require.NotNil(t, storage)
	assert.Equal(t, "/tmp/metrics.json", storage.filePath)
	assert.Equal(t, mockLog, storage.log)
}

func TestSaveData_Success(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	tmpFile := createTempFile(t, nil) // пустой файл

	storage := New(tmpFile, mockLog)

	data := []entity.Metrics{
		{ID: "gauge1", MType: "gauge", Value: floatPtr(3.14)},
		{ID: "counter1", MType: "counter", Delta: intPtr(42)},
	}

	err := storage.SaveData(data)
	assert.NoError(t, err)

	// info, err := os.Stat(tmpFile)
	// require.NoError(t, err)
	// assert.Equal(t, filePermission, info.Mode().Perm()) // on Windows dont work

	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	var loaded []entity.Metrics
	err = json.Unmarshal(content, &loaded)
	require.NoError(t, err)
	assert.Equal(t, data, loaded)
}

func TestLoadData_Success(t *testing.T) {
	data := []entity.Metrics{
		{ID: "gauge1", MType: "gauge", Value: floatPtr(2.71)},
		{ID: "counter1", MType: "counter", Delta: intPtr(100)},
	}

	content, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err)

	tmpFile := createTempFile(t, content)
	mockLog := &testutil.MockLogger{}

	storage := New(tmpFile, mockLog)
	loaded, err := storage.LoadData()

	assert.NoError(t, err)
	assert.Equal(t, data, loaded)
}

func TestLoadData_FileNotFound(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	storage := New("/tmp/nonexistent.json", mockLog)

	data, err := storage.LoadData()

	assert.Error(t, err)
	assert.Equal(t, "file does not exist", err.Error())
	assert.Empty(t, data)
}

func TestLoadData_InvalidJSON(t *testing.T) {
	tmpFile := createTempFile(t, []byte("invalid json {]"))
	mockLog := &testutil.MockLogger{}

	storage := New(tmpFile, mockLog)
	data, err := storage.LoadData()

	assert.Error(t, err)
	assert.Equal(t, "failed to deserialize metrics", err.Error())
	assert.Empty(t, data)
}
