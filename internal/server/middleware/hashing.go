package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
)

func (c *Conveyor) WithHashValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hashFromHeader := r.Header.Get("HashSHA256")
		c.log.Debug("Hash from header", "hash", hashFromHeader)
		if hashFromHeader != "" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}
			r.Body.Close()

			h := hmac.New(sha256.New, []byte(c.hashKey))
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

			w.Header().Set("HashSHA256", responseHashStr)
		}

		io.Copy(w, wrappedWriter.body)
	})
}

type hashResponseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *hashResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}
