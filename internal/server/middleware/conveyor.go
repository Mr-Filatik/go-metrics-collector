// Пакет middleware предоставляет реализации всех middleware используемых в серверном приложении.
package middleware

import (
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
)

// Conveyor описывает сущность конвеера для регистрации middleware.
type Conveyor struct {
	log     logger.Logger // логгер
	hashKey string        // ключ хеширования
}

func New(hashKey string, l logger.Logger) *Conveyor {
	return &Conveyor{
		log:     l,
		hashKey: hashKey,
	}
}

// Middleware описывает сущность для middleware.
type Middleware func(http.Handler) http.Handler

func (c *Conveyor) MainConveyor(h http.Handler) http.Handler {
	if c.hashKey != "" {
		return c.registerConveyor(h, c.WithHashValidation, c.WithCompressedGzip, c.WithLogging)
	}
	return c.registerConveyor(h, c.WithCompressedGzip, c.WithLogging)
}

func (c *Conveyor) registerConveyor(h http.Handler, ms ...Middleware) http.Handler {
	for _, m := range ms {
		h = m(h)
	}
	return h
}
