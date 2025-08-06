package storage

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLogger struct {
	debugCalled []string
	infoCalled  []string
	errorCalled []string
	warnCalled  []string
}

func (m *mockLogger) Log(log logger.LogLevel, msg string, keyvals ...interface{}) {
	m.debugCalled = append(m.debugCalled, msg)
}

func (m *mockLogger) Debug(msg string, keyvals ...interface{}) {
	m.debugCalled = append(m.debugCalled, msg)
}

func (m *mockLogger) Info(msg string, keyvals ...interface{}) {
	m.infoCalled = append(m.infoCalled, msg)
}

func (m *mockLogger) Error(msg string, err error, keyvals ...interface{}) {
	m.errorCalled = append(m.errorCalled, msg)
}

func (m *mockLogger) Warn(msg string, keyvals ...interface{}) {
	m.warnCalled = append(m.warnCalled, msg)
}

func (m *mockLogger) Close() {}

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
	mockLog := &mockLogger{}
	storage := New("/tmp/metrics.json", mockLog)

	require.NotNil(t, storage)
	assert.Equal(t, "/tmp/metrics.json", storage.filePath)
	assert.Equal(t, mockLog, storage.log)
}

func TestSaveData_Success(t *testing.T) {
	mockLog := &mockLogger{}
	tmpFile := createTempFile(t, nil) // пустой файл

	storage := New(tmpFile, mockLog)

	data := []entity.Metrics{
		{ID: "gauge1", MType: "gauge", Value: floatPtr(3.14)},
		{ID: "counter1", MType: "counter", Delta: intPtr(42)},
	}

	err := storage.SaveData(data)
	assert.NoError(t, err)

	info, err := os.Stat(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, filePermission, info.Mode().Perm()) // on Windows dont work

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
	mockLog := &mockLogger{}

	storage := New(tmpFile, mockLog)
	loaded, err := storage.LoadData()

	assert.NoError(t, err)
	assert.Equal(t, data, loaded)
}

func TestLoadData_FileNotFound(t *testing.T) {
	mockLog := &mockLogger{}
	storage := New("/tmp/nonexistent.json", mockLog)

	data, err := storage.LoadData()

	assert.Error(t, err)
	assert.Equal(t, "file does not exist", err.Error())
	assert.Empty(t, data)
}

func TestLoadData_InvalidJSON(t *testing.T) {
	tmpFile := createTempFile(t, []byte("invalid json {]"))
	mockLog := &mockLogger{}

	storage := New(tmpFile, mockLog)
	data, err := storage.LoadData()

	assert.Error(t, err)
	assert.Equal(t, "failed to deserialize metrics", err.Error())
	assert.Empty(t, data)
}

func TestSaveData_WriteError(t *testing.T) {
	mockLog := &mockLogger{}
	storage := New("/root/forbidden.json", mockLog)

	data := []entity.Metrics{{ID: "test", MType: "gauge", Value: floatPtr(1.0)}}

	err := storage.SaveData(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write metrics to file")
}

func TestSaveData_Overwrite(t *testing.T) {
	mockLog := &mockLogger{}
	tmpFile := createTempFile(t, nil)

	storage := New(tmpFile, mockLog)

	err := storage.SaveData([]entity.Metrics{{ID: "first", MType: "gauge", Value: floatPtr(1.0)}})
	require.NoError(t, err)

	err = storage.SaveData([]entity.Metrics{{ID: "second", MType: "gauge", Value: floatPtr(2.0)}})
	require.NoError(t, err)

	loaded, _ := storage.LoadData()
	assert.Len(t, loaded, 1)
	assert.Equal(t, "second", loaded[0].ID)
}
