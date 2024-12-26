package middleware

import (
	"net/http"
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
	return c.registerConveyor(h, c.WithLogging)
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
