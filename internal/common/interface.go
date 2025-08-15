package common

import (
	"context"
	"io"
)

// Starter - интерфейс для всех запускаемых компонентов.
type Starter interface {
	Start(ctx context.Context) error // Метод для запуска компонентов.
}

// Shutdowner - интерфейс для всех останавливаемых компонентов.
type Shutdowner interface {
	Shutdown(ctx context.Context) error // Метод для мягкой остановки компонентов.
	io.Closer                           // Метод для жёсткой остановки компонентов.
}
