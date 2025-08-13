package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Mr-Filatik/go-metrics-collector/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func compressString(data string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("write error %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("close error %w", err)
	}
	return &buf, nil
}

func TestWithCompressedGzip_RequestDecompress(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	originalBody := `{"id":"test","mtype":"gauge","value":3.14}`
	compressedBody, err := compressString(originalBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/update", compressedBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	var capturedBody string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)
		w.WriteHeader(http.StatusOK)
	})

	handler := conveyor.WithCompressedGzip(next)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, originalBody, capturedBody)
}

func TestWithCompressedGzip_ResponseCompress(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"metrics": [{"id":"cpu","value":0.85}]}`))
	})

	handler := conveyor.WithCompressedGzip(next)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	gr, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	uncompressed, err := io.ReadAll(gr)
	require.NoError(t, gr.Close())
	require.NoError(t, err)

	assert.JSONEq(t, `{"metrics": [{"id":"cpu","value":0.85}]}`, string(uncompressed))
}

func TestWithCompressedGzip_NoCompression_Request(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(`{"id":"test"}`))
	req.Header.Set("Content-Type", "application/json")

	var capturedBody string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)
		w.WriteHeader(http.StatusOK)
	})

	handler := conveyor.WithCompressedGzip(next)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"id":"test"}`, capturedBody)
}

func TestWithCompressedGzip_NoCompression_Response(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	req := httptest.NewRequest(http.MethodPost, "/metrics", http.NoBody)
	req.Header.Set("Accept", "application/json")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value": 42}`))
	})

	handler := conveyor.WithCompressedGzip(next)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Equal(t, `{"value": 42}`, rec.Body.String())
}

func TestWithCompressedGzip_InvalidGzip(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader("invalid gzip data"))
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	handler := conveyor.WithCompressedGzip(next)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Failed to decompress request body")
}
