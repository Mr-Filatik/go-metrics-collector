package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
)

// AllClient - клиент для отправки запросов к серверу.
// Совмещает в себе несколько клиентов, используя при каждом запросе разные.
type AllClient struct {
	clients []Client
	mu      sync.Mutex
	current int
}

var _ Client = (*AllClient)(nil)

// NewAllClient создаёт новый экземпляр *AllClient.
func NewAllClient(clnts ...Client) *AllClient {
	client := &AllClient{
		current: 0,
		clients: clnts,
		mu:      sync.Mutex{},
	}

	return client
}

func (c *AllClient) Start(ctx context.Context) error {
	var errs []error
	for cl := range c.clients {
		err := c.clients[cl].Start(ctx)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf("start AllClient error: %w", errors.Join(errs...))
	}
	return nil
}

func (c *AllClient) SendMetric(ctx context.Context, m entity.Metrics) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.clients[c.current].SendMetric(ctx, m)
	c.nextClient()
	if err != nil {
		return fmt.Errorf("send metric in AllClient error: %w", err)
	}
	return nil
}

func (c *AllClient) SendMetrics(ctx context.Context, ms []entity.Metrics) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.clients[c.current].SendMetrics(ctx, ms)
	c.nextClient()
	if err != nil {
		return fmt.Errorf("send metric in AllClient error: %w", err)
	}
	return nil
}

func (c *AllClient) Close() error {
	var errs []error
	for cl := range c.clients {
		err := c.clients[cl].Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf("start AllClient error: %w", errors.Join(errs...))
	}
	return nil
}

func (c *AllClient) nextClient() {
	c.current++
	if c.current >= len(c.clients) {
		c.current = 0
	}
}
