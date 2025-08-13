package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	crypto "github.com/Mr-Filatik/go-metrics-collector/internal/crypto/rsa"
	"github.com/Mr-Filatik/go-metrics-collector/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey, &privateKey.PublicKey
}

func TestWithDecryption_Success(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	privateKey, publicKey := generateTestKeys(t)

	originalBody := `{"id":"test","value":3.14}`

	encryptedBody, err := crypto.EncryptBig([]byte(originalBody), publicKey)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(encryptedBody))
	req.Header.Set("Content-Type", "application/octet-stream")

	var capturedBody string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)
		w.WriteHeader(http.StatusOK)
	})

	handler := conveyor.WithDecryption(next, privateKey)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, originalBody, capturedBody)
}

func TestWithDecryption_NoKey(t *testing.T) {
	mockLog := &testutil.MockLogger{}
	conveyor := New(mockLog)

	body := `{"id":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(body))

	var capturedBody string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)
		w.WriteHeader(http.StatusOK)
	})

	handler := conveyor.WithDecryption(next, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, body, capturedBody)
}
