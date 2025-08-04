// Пакет middleware предоставляет реализации всех middleware используемых в серверном приложении.
package middleware

import (
	"crypto/rsa"
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
)

// Conveyor описывает сущность конвеера для регистрации middleware.
type Conveyor struct {
	log        logger.Logger   // логгер
	privateKey *rsa.PrivateKey // приватный ключ
	hashKey    string          // ключ хеширования
}

// New создаёт и инициализирует новый экзепляр *Conveyor.
//
// Параметры:
//   - hashKey: ключ хэширования
//   - l: логгер
func New(hashKey string, privateKey *rsa.PrivateKey, l logger.Logger) *Conveyor {
	return &Conveyor{
		log:        l,
		privateKey: privateKey,
		hashKey:    hashKey,
	}
}

// Middleware описывает сущность для middleware.
type Middleware func(http.Handler) http.Handler

// MainConveyor создаёт основную последовательность middleware.
//
// Параметры:
//   - h: обработчик
func (c *Conveyor) MainConveyor(h http.Handler) http.Handler {
	if c.hashKey != "" {
		return c.registerConveyor(h,
			c.WithHashValidation,
			c.WithCompressedGzip,
			c.WithDecryption(c.privateKey),
			c.WithLogging)
	}
	return c.registerConveyor(h, c.WithCompressedGzip, c.WithLogging)
}

func (c *Conveyor) registerConveyor(h http.Handler, ms ...Middleware) http.Handler {
	for _, m := range ms {
		h = m(h)
	}
	return h
}
