package middleware

import (
	"compress/gzip"
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
	return c.registerConveyor(h, c.WithLogging, c.WithGzipSupport)
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
			"content_lenght", r.ContentLength,
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

func (c *Conveyor) WithGzipSupport(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" &&
			r.Header.Get("Content-Type") != "text/html" {
			next.ServeHTTP(w, r)
			return
		}

		// Check if the request body is Gzip-encoded
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decode gzip request body", http.StatusInternalServerError)
				return
			}
			r.Body = gr
			if err := gr.Close(); err != nil {
				http.Error(w, "Failed to decode gzip request body", http.StatusInternalServerError)
				return
			}
		}

		// Check if the client supports Gzip encoding
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			c.log.Info(
				"================== Encode ===================",
				"header", r.Header.Get("Accept-Encoding"),
			)

			//Wrap the response writer to enable Gzip encoding
			//w.Header().Set("Content-Encoding", "gzip")
			gw := &gzipWriter{
				ResponseWriter: w,
				Writer:         gzip.NewWriter(w),
			}
			next.ServeHTTP(gw, r)
			if err := gw.Writer.(*gzip.Writer).Close(); err != nil {
				http.Error(w, "Failed to decode gzip request body", http.StatusInternalServerError)
				return
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *gzipWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}
