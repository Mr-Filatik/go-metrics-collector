package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/urfave/negroni"
)

func (c *Conveyor) WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-Id") == "" {
			r.Header.Set("X-Request-Id", uuid.New().String())
		}
		startTime := time.Now().UTC()
		requestID := r.Header.Get("X-Request-Id")

		lwr := negroni.NewResponseWriter(w)

		next.ServeHTTP(lwr, r)

		statusCode := lwr.Status()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		c.log.Info(
			"Request-Response",
			"request_id", requestID,
			"request_method", r.Method,
			"request_uri", r.RequestURI,
			"request_time", startTime.String(),
			"request_duration", time.Since(startTime),
			"status", statusCode,
			"content_lenght", lwr.Size(),
		)
	})
}
