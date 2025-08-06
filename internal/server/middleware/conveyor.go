// Пакет middleware предоставляет реализации всех middleware используемых в серверном приложении.
package middleware

import (
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
)

// Middleware описывает сущность для middleware.
type Middleware func(http.Handler) http.Handler

// Conveyor описывает сущность конвеера для регистрации middleware.
type Conveyor struct {
	log         logger.Logger // логгер
	middlewares []Middleware  // зарегистрированные middlewares
}

// New создаёт и инициализирует новый экзепляр *Conveyor.
//
// Параметры:
//   - l: логгер
func New(l logger.Logger) *Conveyor {
	return &Conveyor{
		log:         l,
		middlewares: make([]Middleware, 0),
	}
}

// RegisterMiddlewares регистрирует слайс middlewares для применения их ко всем http.handler.
//
// Порядок применения middleware следующий: m1 -> m2 -> m3 -> ... -> m3 -> m2 -> m1.
//
// Использовать так (для поддержки замыкания):
//
//	func(h http.Handler) http.Handler {
//		return s.conveyor.WithMiddleware(h, data)
//	}
func (c *Conveyor) RegisterMiddlewares(ms ...Middleware) {
	c.middlewares = ms
}

// Middlewares оборачивает http.handler в зарегистрированные middlewares.
//
// Порядок применения middleware следующий: m1 -> m2 -> m3 -> ... -> m3 -> m2 -> m1.
func (c *Conveyor) Middlewares(h http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}
