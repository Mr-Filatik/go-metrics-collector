package server

import (
	"context"
	"crypto/rsa"
	"errors"
	"net"

	"github.com/Mr-Filatik/go-metrics-collector/internal/common"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server/interceptor"
	"github.com/Mr-Filatik/go-metrics-collector/internal/service"
	"github.com/Mr-Filatik/go-metrics-collector/proto"
	"google.golang.org/grpc"
)

// GrpcServer представляет gRPC-сервер приложения.
// Использует service для бизнес-логики и logger для логирования.
type GrpcServer struct {
	proto.UnimplementedMetricsServiceServer
	serv    *grpc.Server
	service *service.Service // сервис с основной логикой
	// conveyor    *middleware.Conveyor // конвейер для middleware
	log           logger.Logger // логгер
	address       string
	trustedSubnet string
	hashKey       string
}

var _ Server = (*GrpcServer)(nil)

type GrpcServerConfig struct {
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
func NewGrpcServer(ctx context.Context, conf *GrpcServerConfig, log logger.Logger) *GrpcServer {
	log.Info("GrpcServer creating...")

	srv := &GrpcServer{
		service:       conf.Service,
		log:           log,
		trustedSubnet: conf.TrustedSubnet,
		address:       conf.Address,
		hashKey:       conf.HashKey,
	}

	if adr, err := common.ChangePortForGRPC(conf.Address); err == nil {
		srv.address = adr
	}

	log.Info("GrpcServer create is successfull")
	return srv
}

func (s *GrpcServer) Start(ctx context.Context) error {
	s.log.Info(
		"GrpcServer starting...",
		"address", s.address,
	)

	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		s.log.Error("Error listen in gRPC server", err)
	}
	conv := interceptor.New(s.trustedSubnet, s.hashKey, s.log)

	var opts []grpc.ServerOption
	opts = append(opts, grpc.ChainUnaryInterceptor(
		conv.LoggingInterceptor,
		conv.TrustingInterceptor,
		conv.HashingInterceptor,
	))
	grpcServ := grpc.NewServer(opts...)
	s.serv = grpcServ
	proto.RegisterMetricsServiceServer(grpcServ, s)
	go func() {
		if err := grpcServ.Serve(lis); err != nil {
			s.log.Error("Error in GrpcServer", err)
		}
	}()

	s.log.Info("GrpcServer start is successfull")
	return nil
}

func (s *GrpcServer) Shutdown(ctx context.Context) error {
	s.log.Info("GRPCServer shutdown starting...")
	s.serv.GracefulStop()
	s.log.Info("GRPCServer shutdown is successfull")
	return nil
}

func (s *GrpcServer) Close() error {
	s.log.Info("GRPCServer close starting...")
	s.serv.Stop()
	s.log.Info("GRPCServer close is successfull")
	return nil
}

// UpdateMetrics обновление списка метрик.
//
// Параметры:
//   - ctx: контекст для отмены;
//   - req: запрос.
func (s *GrpcServer) UpdateMetrics(
	ctx context.Context,
	req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	metr := getMetricsFromProto(req)

	for _, m := range metr {
		_, err := s.service.CreateOrUpdate(ctx, m)
		if err != nil {
			if err.Error() == service.MetricNotFound || err.Error() == service.MetricUncorrect {
				return nil, errors.New("uncorrect request data")
			}
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
