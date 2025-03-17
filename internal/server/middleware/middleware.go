package middleware

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/google/uuid"
	"github.com/urfave/negroni"
)

type Conveyor struct {
	log logger.Logger
}

func New(l logger.Logger) *Conveyor {
	return &Conveyor{
		log: l,
	}
}

type Middleware func(http.Handler) http.Handler

func (c *Conveyor) MainConveyor(h http.Handler) http.Handler {
	return c.registerConveyor(h, c.WithCompressedGzip, c.WithLogging)
}

func (c *Conveyor) registerConveyor(h http.Handler, ms ...Middleware) http.Handler {
	for _, m := range ms {
		h = m(h)
	}
	return h
}

func (c *Conveyor) WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-Id") == "" {
			r.Header.Set("X-Request-Id", uuid.New().String())
		}
		startTime := time.Now().UTC()
		requestID := r.Header.Get("X-Request-Id")

		lwr := negroni.NewResponseWriter(w)

		next.ServeHTTP(lwr, r)

		c.log.Info(
			"Request",
			"request_id", requestID,
			"request_method", r.Method,
			"request_uri", r.RequestURI,
			"request_time", startTime.String(),
			"request_duration", time.Since(startTime),
		)

		statusCode := lwr.Status()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		c.log.Info(
			"Responce",
			"request_id", requestID,
			"status", statusCode,
			"content_lenght", lwr.Size(),
		)
	})
}

func (c *Conveyor) WithCompressedGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		valueContent := r.Header.Get("Content-Type")
		isType := strings.Contains(valueContent, "application/json") || strings.Contains(valueContent, "text/html")
		if isType && strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request body", http.StatusBadRequest)
				return
			}
			defer func() {
				if err := gz.Close(); err != nil {
					http.Error(w, "Failed to decompress request body", http.StatusBadRequest)
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
					http.Error(w, "Failed to decompress request body", http.StatusBadRequest)
				}
			}()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Type", valueAccept)
			w = &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		}

		next.ServeHTTP(w, r)
	})
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
