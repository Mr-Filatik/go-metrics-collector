// Пакет middleware предоставляет реализации всех middleware используемых в серверном приложении.
package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Mr-Filatik/go-metrics-collector/internal/common"
)

// WithHashValidation добавляет хэширование в middleware.
//
// Параметры:
//   - next: следующий обработчик
//   - hashKey: ключ хэширования
func (c *Conveyor) WithHashValidation(next http.Handler, hashKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hashFromHeader := r.Header.Get(common.HeaderHashSHA256)
		c.log.Debug("Hash from header", "hash", hashFromHeader)
		if hashFromHeader != "" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}
			berr := r.Body.Close()
			if berr != nil {
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}

			h := hmac.New(sha256.New, []byte(hashKey))
			h.Write(body)
			calculatedHash := h.Sum(nil)
			calculatedHashStr := hex.EncodeToString(calculatedHash)
			c.log.Debug("Calculated hash", "hash", calculatedHashStr)

			if !strings.EqualFold(calculatedHashStr, hashFromHeader) {
				http.Error(w, "Hash mismatch", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		var responseWriterWrapper struct {
			http.ResponseWriter
			body *bytes.Buffer
		}
		responseWriterWrapper.body = &bytes.Buffer{}
		responseWriterWrapper.ResponseWriter = w

		wrappedWriter := &hashResponseWriter{
			ResponseWriter: responseWriterWrapper.ResponseWriter,
			body:           responseWriterWrapper.body,
		}

		next.ServeHTTP(wrappedWriter, r)

		if hashFromHeader != "" {
			responseBody := wrappedWriter.body.Bytes()
			responseHash := sha256.Sum256(responseBody)
			responseHashStr := hex.EncodeToString(responseHash[:])

			w.Header().Set(common.HeaderHashSHA256, responseHashStr)
		}

		_, cerr := io.Copy(w, wrappedWriter.body)
		if cerr != nil {
			http.Error(w, "Failed to copy request body", http.StatusInternalServerError)
			return
		}
	})
}

type hashResponseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *hashResponseWriter) Write(b []byte) (int, error) {
	// return w.body.Write(b)
	num, err := w.body.Write(b)
	if err != nil {
		return num, errors.New(err.Error())
	}
	return num, nil
}
