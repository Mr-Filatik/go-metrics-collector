package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"google.golang.org/grpc"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/proto"
)

// GrpcClient - клиент для отправки запросов к серверу.
type GrpcClient struct {
	conn                 *grpc.ClientConn
	metricsServiceClient proto.MetricsServiceClient
	log                  logger.Logger
	url                  string
}

var _ Client = (*GrpcClient)(nil)

// GrpcClientConfig - структура, содержащая основные параметры для RestyClient.
type GrpcClientConfig struct {
	URL string
}

// NewGrpcClient создаёт новый экземпляр *GrpcClient.
func NewGrpcClient(config *GrpcClientConfig, l logger.Logger) *GrpcClient {
	client := &GrpcClient{
		log: l,
		url: config.URL,
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, connErr := grpc.NewClient(":18080", opts...) // client.url
	if connErr != nil {
		client.log.Error("grpc.NewClient() error", connErr)
	}
	client.conn = conn

	client.metricsServiceClient = proto.NewMetricsServiceClient(client.conn)

	// client.registerMiddlewares(config.HashKey, config.PublicKey)

	return client
}

func (c *GrpcClient) SendMetric(m entity.Metrics) error {
	c.log.Warn("Not implemented *GrpcClient.SendMetric().", nil)
	return nil
}

func (c *GrpcClient) SendMetrics(ms []entity.Metrics) error {
	metrics := make([]*proto.Metric, 0, len(ms))

	for i := range ms {
		pm := &proto.Metric{
			Id:    ms[i].ID,
			Mtype: ms[i].MType,
			Value: ms[i].Value,
			Delta: ms[i].Delta,
		}
		metrics = append(metrics, pm)
	}

	req := &proto.UpdateMetricsRequest{
		Metrics: metrics,
	}

	_, err := c.metricsServiceClient.UpdateMetrics(context.Background(), req)
	if err != nil {
		c.log.Error("UpdateMetric error", err)
	}

	return nil
}

func (c *GrpcClient) Close() error {
	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("close *GrpcClient.Close() error: %w", err)
	}
	return nil
}

func removePortFromURL(input string) (string, error) {
	parsed, err := url.Parse(input)
	if err != nil {
		return "", err
	}

	// Удаляем порт из Host
	host := parsed.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Пересобираем URL
	parsed.Host = host
	return parsed.String(), nil
}
