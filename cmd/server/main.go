package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	crypto "github.com/Mr-Filatik/go-metrics-collector/internal/crypto/rsa"
	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
	repositoryMemory "github.com/Mr-Filatik/go-metrics-collector/internal/repository/memory"
	repositoryPostgres "github.com/Mr-Filatik/go-metrics-collector/internal/repository/postgres"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server"
	config "github.com/Mr-Filatik/go-metrics-collector/internal/server/config"
	"github.com/Mr-Filatik/go-metrics-collector/internal/service"
	storage "github.com/Mr-Filatik/go-metrics-collector/internal/storage/file"

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

	var mainServer server.Server

	// Создание и запуск HTTP сервера
	servConf := &server.HTTPServerConfig{
		Address:       conf.ServerAddress,
		Service:       srvc,
		HashKey:       conf.HashKey,
		TrustedSubnet: conf.TrustedSubnet,
		PrivateRsaKey: key,
	}
	mainServer = server.NewHTTPServer(exitCtx, servConf, log)

	startErr := mainServer.Start(exitCtx)
	if startErr != nil {
		log.Error("Server starting error.", startErr)
		return
	}

	// Создание и запуск gRPC сервера
	var grpcServer *server.GrpcServer

	if conf.GrpcEnabled {
		grpcConf := &server.GrpcServerConfig{
			Address:       conf.ServerAddress,
			Service:       srvc,
			HashKey:       conf.HashKey,
			TrustedSubnet: conf.TrustedSubnet,
			PrivateRsaKey: key,
		}
		grpcServer = server.NewGrpcServer(exitCtx, grpcConf, log)

		gStartErr := grpcServer.Start(exitCtx)
		if gStartErr != nil {
			log.Error("Server starting error.", gStartErr)
			return
		}
	}

	// Ожидание сигнала остановки
	<-exitCtx.Done()
	exitFn()

	// Запускаем полноценную остановку с таймаутом
	log.Info("Application shutdown starting...")
	shutdownCtx, cansel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cansel()

	go func() {
		if conf.GrpcEnabled {
			gshutdownErr := grpcServer.Shutdown(shutdownCtx)
			if gshutdownErr != nil {
				log.Error("Shutdown GRPCserver error", gshutdownErr)
			}
		}
	}()

	go func() {
		shutdownErr := mainServer.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			log.Error("Shutdown server error", shutdownErr)
		}
	}()

	<-shutdownCtx.Done()

	if conf.GrpcEnabled {
		closeErr := grpcServer.Close()
		if closeErr != nil {
			log.Error("Close server error", closeErr)
		}
	}

	closeErr := mainServer.Close()
	if closeErr != nil {
		log.Error("Close server error", closeErr)
	}

	log.Info("Application shutdown is successfull")
}
