package server

import (
	"context"
	"crypto/rsa"
	"errors"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/service"
	"github.com/Mr-Filatik/go-metrics-collector/proto"
	"google.golang.org/grpc/metadata"
)

// GrpcServer представляет gRPC-сервер приложения.
// Использует service для бизнес-логики и logger для логирования.
type GrpcServer struct {
	proto.UnimplementedMetricsServiceServer
	service *service.Service // сервис с основной логикой
	// conveyor    *middleware.Conveyor // конвейер для middleware
	log logger.Logger // логгер
}

type GrpcServerConfig struct {
	Logger        logger.Logger
	PrivateRsaKey *rsa.PrivateKey
	Service       *service.Service
	Address       string
	HashKey       string
	TrustedSubnet string
}

// NewGrpcServer создаёт и инициализирует новый экзепляр *GrpcServer.
//
// Параметры:
//   - ctx: контекст для остановки;
//   - conf: конфиг сервера.
func NewGrpcServer(ctx context.Context, conf *GrpcServerConfig) *GrpcServer {
	srv := &GrpcServer{
		service: conf.Service,
		// conveyor: middleware.New(conf.Logger),
		log: conf.Logger,
	}
	// srv.registerMiddlewares(conf.HashKey, conf.PrivateRsaKey, conf.TrustedSubnet)
	// srv.registerRoutes()
	return srv
}

// UpdateMetrics обновление списка метрик.
//
// Параметры:
//   - ctx: контекст для отмены;
//   - req: запрос.
func (s *GrpcServer) UpdateMetrics(
	ctx context.Context,
	req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.log.Info("Not metadata.")
	}
	for i := range md {
		s.log.Info(i)
		md.Get(i)
	}

	metr := getMetricsFromProto(req)

	for _, m := range metr {
		_, err := s.service.CreateOrUpdate(ctx, m)
		if err != nil {
			if err.Error() == service.MetricNotFound || err.Error() == service.MetricUncorrect {
				//s.serverResponceBadRequest(w, err)
				return nil, errors.New("uncorrect request data")
			}
			//s.serverResponceInternalServerError(w, err)
			return nil, errors.New("unespected error")
		}
	}

	return &proto.UpdateMetricsResponse{}, nil
}

func getMetricsFromProto(req *proto.UpdateMetricsRequest) []entity.Metrics {
	protoMetrics := req.GetMetrics()
	metrics := make([]entity.Metrics, 0, len(protoMetrics))

	for i := range protoMetrics {
		val := protoMetrics[i].GetValue()
		del := protoMetrics[i].GetDelta()
		pm := entity.Metrics{
			ID:    protoMetrics[i].GetId(),
			MType: protoMetrics[i].GetMtype(),
			Value: &val,
			Delta: &del,
		}
		metrics = append(metrics, pm)
	}

	return metrics
}
