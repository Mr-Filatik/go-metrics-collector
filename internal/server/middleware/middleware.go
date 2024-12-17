package middleware

import (
	"log"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func MainConveyor(h http.Handler) http.Handler {
	return registerConveyor(h, WithLogging)
}

func registerConveyor(h http.Handler, ms ...Middleware) http.Handler {
	for _, m := range ms {
		h = m(h)
	}
	return h
}

func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %v", r.Method)
		next.ServeHTTP(w, r)
		log.Printf("Responce: %v", r.Method)
	})
}
