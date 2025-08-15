package main

import (
	"context"
	"crypto/rsa"
	"fmt"
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

	// Создание HTTP сервера
	servConf := &server.HTTPServerConfig{
		Address:       conf.ServerAddress,
		Service:       srvc,
		HashKey:       conf.HashKey,
		TrustedSubnet: conf.TrustedSubnet,
		PrivateRsaKey: key,
	}
	mainServer = server.NewHTTPServer(exitCtx, servConf, log)

	// Создание gRPC сервера
	// servConf := &server.GrpcServerConfig{
	// 	Address:       conf.ServerAddress,
	// 	Service:       srvc,
	// 	HashKey:       conf.HashKey,
	// 	TrustedSubnet: conf.TrustedSubnet,
	// 	PrivateRsaKey: key,
	// }
	// mainServer = server.NewGrpcServer(exitCtx, servConf, log)

	// Запуск сервера
	startErr := mainServer.Start(exitCtx)
	if startErr != nil {
		log.Error("Server starting error.", startErr)
		return
	}

	// Ожидание сигнала остановки
	<-exitCtx.Done()
	exitFn()

	// Запускаем полноценную остановку с таймаутом
	log.Info("Application shutdown starting...")
	shutdownCtx, cansel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cansel()

	shutdownErr := mainServer.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		log.Error("Shutdown server error", shutdownErr)
		closeErr := mainServer.Close()
		if closeErr != nil {
			log.Error("Close server error", closeErr)
		}
	}

	log.Info("Application shutdown is successfull")
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
