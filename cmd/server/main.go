package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
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

	// Запуск сервера
	serv := server.NewServer(exitCtx, conf.ServerAddress, srvc, conf.HashKey, key, log)
	go func() {
		log.Info(
			"Start server",
			"endpoint", conf.ServerAddress,
		)
		if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Error server", err)
		}
		log.Info(
			"Finish server",
			"endpoint", conf.ServerAddress,
		)
	}()

	// Ожидание сигнала остановки
	<-exitCtx.Done()
	exitFn()

	// Запускаем полноценную остановку с таймаутом
	log.Info("Start server shutdown")
	shutdownCtx, cansel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cansel()

	err := serv.Shutdown(shutdownCtx)
	if err != nil {
		log.Error("Shutdown server error", err)
		cErr := serv.Close()
		if cErr != nil {
			log.Error("Close server error", cErr)
		}
	}

	log.Info("Finish server shutdown")
}
