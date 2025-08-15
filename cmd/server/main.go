package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/signal"
	"strings"
	"syscall"
	"time"

	crypto "github.com/Mr-Filatik/go-metrics-collector/internal/crypto/rsa"
	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
	repositoryMemory "github.com/Mr-Filatik/go-metrics-collector/internal/repository/memory"
	repositoryPostgres "github.com/Mr-Filatik/go-metrics-collector/internal/repository/postgres"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server"
	config "github.com/Mr-Filatik/go-metrics-collector/internal/server/config"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server/interceptor"
	"github.com/Mr-Filatik/go-metrics-collector/internal/service"
	storage "github.com/Mr-Filatik/go-metrics-collector/internal/storage/file"
	"github.com/Mr-Filatik/go-metrics-collector/proto"
	"google.golang.org/grpc"

	// Installing the gzip encoding registers it as an available compressor.
	// The gRPC will automatically negotiate and use gzip if the client supports it.
	_ "google.golang.org/grpc/encoding/gzip"
)

// go run -ldflags "-X main.buildVersion=v2.0.0 -X main.buildDate=2025-07-07 -X main.buildCommit=98d1d98".
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

const (
	shutdownTimeout = 5 * time.Second
)

func main() {
	log := logger.New(logger.LevelDebug)
	defer log.Close()

	log.Info(fmt.Sprintf("Build version: %v", buildVersion))
	log.Info(fmt.Sprintf("Build date: %v", buildDate))
	log.Info(fmt.Sprintf("Build commit: %v", buildCommit))

	conf := config.Initialize()

	var key *rsa.PrivateKey = nil
	if conf.CryptoKeyPath != "" {
		k, err := crypto.LoadPrivateKey(conf.CryptoKeyPath)
		if err != nil {
			log.Error("Load private key error", err)
		}
		key = k
	}

	var srvc *service.Service
	if conf.ConnectionString != "" {
		repo, err := repositoryPostgres.New(conf.ConnectionString, log)
		if err != nil {
			panic(err.Error())
		}
		defer repo.Close()
		srvc = service.New(repo, nil, 0, log)
	} else {
		repo := repositoryMemory.New(conf.ConnectionString, log)
		stor := storage.New(conf.FileStoragePath, log)
		srvc = service.New(repo, stor, conf.StoreInterval, log)
	}
	srvc.Start(conf.Restore)
	defer srvc.Stop()

	// Привязка сигналов ОС к контексту
	exitCtx, exitFn := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer exitFn()

	// Создание HTTP сервера
	servConf := &server.ServerConfig{
		Address:       conf.ServerAddress,
		Service:       srvc,
		HashKey:       conf.HashKey,
		TrustedSubnet: conf.TrustedSubnet,
		PrivateRsaKey: key,
		Logger:        log,
	}
	serv := server.NewServer(exitCtx, servConf)

	// Запуск HTTP сервера
	go func() {
		log.Info(
			"Start HTTP server",
			"address", servConf.Address,
		)
		if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Error in HTTP server", err)
		}
		log.Info("Finish HTTP server")
	}()

	// Создание gRPC сервера
	gservConf := &server.GrpcServerConfig{
		Address:       conf.ServerAddress,
		Service:       srvc,
		HashKey:       conf.HashKey,
		TrustedSubnet: conf.TrustedSubnet,
		PrivateRsaKey: key,
		Logger:        log,
	}
	gserv := server.NewGrpcServer(exitCtx, gservConf)

	// Запуск gRPC сервера
	go func() {
		log.Info(
			"Start gRPC server",
			"address", conf.ServerAddress,
		)
		lis, err := net.Listen("tcp", ":18080")
		if err != nil {
			log.Error("Error listen in gRPC server", err)
		}
		conv := interceptor.New(conf.TrustedSubnet, conf.HashKey, log)

		var opts []grpc.ServerOption
		opts = append(opts, grpc.ChainUnaryInterceptor(
			conv.LoggingInterceptor,
			conv.TrustingInterceptor,
			conv.HashingInterceptor,
		))
		grpcServ := grpc.NewServer(opts...)
		proto.RegisterMetricsServiceServer(grpcServ, gserv)
		if err := grpcServ.Serve(lis); err != nil {
			log.Error("Error in gRPC server", err)
		}
		log.Info("Finish gRPC server")
	}()

	// Ожидание сигнала остановки
	<-exitCtx.Done()
	exitFn()

	// Запускаем полноценную остановку с таймаутом
	log.Info("Start shutdown")
	shutdownCtx, cansel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cansel()

	err := serv.Shutdown(shutdownCtx)
	if err != nil {
		log.Error("Shutdown HTTP server error", err)
		cErr := serv.Close()
		if cErr != nil {
			log.Error("Close HTTP server error", cErr)
		}
	}

	log.Info("Finish shutdown")
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
