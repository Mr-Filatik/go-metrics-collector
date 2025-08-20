// Пакет server предоставляет реализацию серверного приложения.
package server

import (
	"errors"

	"github.com/Mr-Filatik/go-metrics-collector/internal/common"
)

var (
	ErrEmptyBody   = errors.New("body is empty")
	ErrNotByteBody = errors.New("body is not of type []byte")

	ErrServerNotStarted = errors.New("server not started")
)

type Server interface {
	common.Starter
	common.Shutdowner
}
