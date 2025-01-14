package middleware

import (
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
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
