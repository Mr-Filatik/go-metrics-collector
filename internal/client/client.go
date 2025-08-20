package client

import (
	"context"
	"errors"
	"io"

	"github.com/Mr-Filatik/go-metrics-collector/internal/common"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
)

var (
	ErrEmptyBody   = errors.New("body is empty")
	ErrNotByteBody = errors.New("body is not of type []byte")

	ErrClientNotStarted = errors.New("client not started")
)

// Client - интерфейс для всех клиентов приложения.
type Client interface {
	common.Starter
	io.Closer
	SendMetric(ctx context.Context, m entity.Metrics) error
	SendMetrics(ctx context.Context, ms []entity.Metrics) error
}
