package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWithLogging_RequestID_Generated(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		assert.NotEmpty(t, id)
		_, err := uuid.Parse(id)
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	})

	handler := conveyor.WithLogging(next)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
}

func TestWithLogging_RequestID_Exists(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	existingID := "test-123"
	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	req.Header.Set("X-Request-Id", existingID)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, existingID, r.Header.Get("X-Request-Id"))
		w.WriteHeader(http.StatusOK)
	})

	handler := conveyor.WithLogging(next)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
}

func TestWithLogging_LogsCorrectFields(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(`{"value": 42}`))
	req.Header.Set("X-Request-Id", "req-123")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	handler := conveyor.WithLogging(next)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logs := mockLog.GetAllLogs()
	log := mockLog.GetLastLog()
	assert.Equal(t, 1, len(logs))
	assert.Contains(t, log.Keyvals, "request_id")
	assert.Contains(t, log.Keyvals, "req-123")
	assert.Contains(t, log.Keyvals, "request_method")
	assert.Contains(t, log.Keyvals, "POST")
	assert.Contains(t, log.Keyvals, "request_uri")
	assert.Contains(t, log.Keyvals, "/update")
	assert.Contains(t, log.Keyvals, "status")
	assert.Contains(t, log.Keyvals, 200)
	assert.Contains(t, log.Keyvals, "content_lenght")
	assert.Contains(t, log.Keyvals, 16)

	assert.Contains(t, log.Keyvals, "request_duration")
	assert.Contains(t, log.Keyvals, "request_time")
}

func TestWithLogging_DefaultStatus(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	handler := conveyor.WithLogging(next)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestWithLogging_ContentLength(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	req := httptest.NewRequest(http.MethodGet, "/data", http.NoBody)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	handler := conveyor.WithLogging(next)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, 5, len(rec.Body.Bytes()))
}
