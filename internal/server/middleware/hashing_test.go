package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func calculateHash(body []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

func TestWithHashValidation_ValidRequestHash(t *testing.T) {
	mockLog := &mockLogger{}
	conveyor := New(mockLog)

	hashKey := "mysecret"
	body := `{"id":"test","mtype":"gauge","value":3.14}`
	expectedHash := calculateHash([]byte(body), hashKey)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(body))
	req.Header.Set("HashSHA256", expectedHash)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	handler := conveyor.WithHashValidation(next, hashKey)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Body.String())
}

func TestWithHashValidation_InvalidRequestHash(t *testing.T) {
	mockLog := &mockLogger{}
	conveyor := New(mockLog)

	hashKey := "mysecret"
	body := `{"id":"test"}`
	invalidHash := "invalidhash123"

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(body))
	req.Header.Set("HashSHA256", invalidHash)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	handler := conveyor.WithHashValidation(next, hashKey)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Hash mismatch")
}

func TestWithHashValidation_NoHash(t *testing.T) {
	mockLog := &mockLogger{}
	conveyor := New(mockLog)

	hashKey := "mysecret"
	body := `{"id":"test"}`

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(body))

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	handler := conveyor.WithHashValidation(next, hashKey)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Body.String())
	assert.Empty(t, rec.Header().Get("HashSHA256"))
}

func TestWithHashValidation_ResponseSigned(t *testing.T) {
	mockLog := &mockLogger{}
	conveyor := New(mockLog)

	hashKey := "mysecret"
	body := `{"id":"test"}`
	requestHash := calculateHash([]byte(body), hashKey)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(body))
	req.Header.Set("HashSHA256", requestHash)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result": "success"}`))
	})

	handler := conveyor.WithHashValidation(next, hashKey)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	responseHash := rec.Header().Get("HashSHA256")
	assert.NotEmpty(t, responseHash)

	responseHashFromBody := sha256.Sum256(rec.Body.Bytes())
	responseHashStrFromBody := hex.EncodeToString(responseHashFromBody[:])
	assert.Equal(t, responseHashStrFromBody, responseHash)
}

func TestWithHashValidation_EmptyBodyWithHash(t *testing.T) {
	mockLog := &mockLogger{}
	conveyor := New(mockLog)

	hashKey := "mysecret"
	emptyBody := ""
	expectedHash := calculateHash([]byte(emptyBody), hashKey)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(emptyBody))
	req.Header.Set("HashSHA256", expectedHash)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := conveyor.WithHashValidation(next, hashKey)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
