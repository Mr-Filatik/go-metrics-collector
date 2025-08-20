package client

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/Mr-Filatik/go-metrics-collector/internal/common"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	myProto "github.com/Mr-Filatik/go-metrics-collector/proto"
)

// GrpcClient - клиент для отправки запросов к серверу.
type GrpcClient struct {
	conn                 *grpc.ClientConn
	metricsServiceClient myProto.MetricsServiceClient
	log                  logger.Logger
	url                  string
	xRealIP              string
	hashKey              string
}

var _ Client = (*GrpcClient)(nil)

// GrpcClientConfig - структура, содержащая основные параметры для RestyClient.
type GrpcClientConfig struct {
	URL     string
	XRealIP string
	HashKey string
}

// NewGrpcClient создаёт новый экземпляр *GrpcClient.
func NewGrpcClient(config *GrpcClientConfig, l logger.Logger) *GrpcClient {
	client := &GrpcClient{
		log:     l,
		xRealIP: config.XRealIP,
		url:     config.URL,
		hashKey: config.HashKey,
	}

	if adr, err := common.ChangePortForGRPC(config.URL); err == nil {
		client.url = adr
	}

	return client
}

func (c *GrpcClient) Start(_ context.Context) error {
	c.log.Info(
		"Start GrpcClient...",
		"address", c.url,
	)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, connErr := grpc.NewClient(c.url, opts...)
	if connErr != nil {
		return fmt.Errorf("start GrpcClient error: %w", connErr)
	}

	c.conn = conn
	c.metricsServiceClient = myProto.NewMetricsServiceClient(c.conn)
	c.log.Info("Start GrpcClient is successfull")
	return nil
}

func (c *GrpcClient) SendMetric(_ context.Context, _ entity.Metrics) error {
	if c.conn == nil {
		err := fmt.Errorf("GrpcClient: %w", ErrClientNotStarted)
		c.log.Error("Error in *GrpcClient.SendMetric()", err)
		return err
	}

	c.log.Warn("Not implemented *GrpcClient.SendMetric()", nil)
	return nil
}

func (c *GrpcClient) SendMetrics(ctx context.Context, ms []entity.Metrics) error {
	if c.conn == nil {
		err := fmt.Errorf("GrpcClient: %w", ErrClientNotStarted)
		c.log.Error("Error in *GrpcClient.SendMetrics()", err)
		return err
	}

	metrics := make([]*myProto.Metric, 0, len(ms))

	for i := range ms {
		pm := &myProto.Metric{
			Id:    ms[i].ID,
			Mtype: ms[i].MType,
			Value: ms[i].Value,
			Delta: ms[i].Delta,
		}
		metrics = append(metrics, pm)
	}

	req := &myProto.UpdateMetricsRequest{
		Metrics: metrics,
	}

	data, merr := proto.Marshal(req)
	if merr != nil {
		c.log.Error("Failed to marshal request", merr)
	}
	hashStr, herr := common.HashBytesToString(data, c.hashKey)
	if herr != nil {
		c.log.Error("Calculate hash error", herr)
	}

	md := metadata.Pairs(
		strings.ToLower(common.HeaderXRealIP), c.xRealIP,
		strings.ToLower(common.HeaderHashSHA256), hashStr,
	)
	ctxUpd := metadata.NewOutgoingContext(ctx, md)

	_, err := c.metricsServiceClient.UpdateMetrics(ctxUpd, req, grpc.UseCompressor(gzip.Name))
	if err != nil {
		c.log.Error("UpdateMetric error", err)
	}

	return nil
}

func (c *GrpcClient) Close() error {
	if c.conn == nil {
		err := fmt.Errorf("GrpcClient: %w", ErrClientNotStarted)
		c.log.Error("Error in *GrpcClient.Close()", err)
		return err
	}

	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("close *GrpcClient.Close() error: %w", err)
	}
	return nil
}
