// Пакет middleware предоставляет реализации всех middleware используемых в серверном приложении.
package middleware

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
)

// WithCompressedGzip добавляет сжатие в middleware.
//
// Параметры:
//   - next: следующий обработчик
func (c *Conveyor) WithCompressedGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		valueContent := r.Header.Get("Content-Type")
		isType := strings.Contains(valueContent, "application/json") || strings.Contains(valueContent, "text/html")
		if isType && strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				c.httpError(w, "Failed to decompress request body", http.StatusBadRequest, err)
				return
			}
			defer func() {
				if err := gz.Close(); err != nil {
					c.httpError(w, "Failed to decompress request body", http.StatusBadRequest, err)
				}
			}()
			r.Body = gz
		}

		valueAccept := r.Header.Get("Accept")
		isType = strings.Contains(valueAccept, "application/json") || strings.Contains(valueAccept, "text/html")
		if isType && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz := gzip.NewWriter(w)
			defer func() {
				if err := gz.Close(); err != nil {
					c.httpError(w, "Failed to compress response body", http.StatusBadRequest, err)
				}
			}()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Type", valueAccept)
			w = &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		}

		next.ServeHTTP(w, r)
	})
}

func (c *Conveyor) httpError(w http.ResponseWriter, message string, code int, err error) {
	c.log.Error(message, err)
	http.Error(w, message, code)
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	num, err := w.Writer.Write(b)
	if err != nil {
		return num, errors.New(err.Error())
	}
	return num, nil
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}
