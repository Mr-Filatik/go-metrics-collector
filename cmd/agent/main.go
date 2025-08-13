package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	_ "net/http/pprof"
	"os/signal"
	"syscall"

	config "github.com/Mr-Filatik/go-metrics-collector/internal/agent/config"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/reporter"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/updater"
	crypto "github.com/Mr-Filatik/go-metrics-collector/internal/crypto/rsa"
	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
	"github.com/go-resty/resty/v2"
)

// go run -ldflags "-X main.buildVersion=v2.0.0 -X main.buildDate=2025-07-07 -X main.buildCommit=98d1d98".
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	log := logger.New(logger.LevelInfo)
	defer log.Close()

	log.Info(fmt.Sprintf("Build version: %v", buildVersion))
	log.Info(fmt.Sprintf("Build date: %v", buildDate))
	log.Info(fmt.Sprintf("Build commit: %v", buildCommit))

	conf := config.Initialize()
	metrics := metric.New()

	var key *rsa.PublicKey = nil
	if conf.CryptoKeyPath != "" {
		k, err := crypto.LoadPublicKey(conf.CryptoKeyPath)
		if err != nil {
			log.Error("Load private key error", err)
			return
		}
		key = k
	}

	realIP, err := getExternalRealIP()
	if err != nil {
		panic(err)
	}

	// Привязка сигналов ОС к контексту
	exitCtx, exitFn := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer exitFn()

	go updater.Run(exitCtx, metrics, conf.PollInterval)
	go updater.RunMemory(exitCtx, metrics, conf.PollInterval)
	go reporter.Run(
		exitCtx,
		metrics,
		conf.ServerAddress,
		conf.ReportInterval,
		conf.HashKey,
		conf.RateLimit,
		key,
		realIP,
		log)

	// Ожидание сигнала остановки
	<-exitCtx.Done()
	exitFn()

	log.Info("Finish agent shutdown")
}

func getExternalRealIP() (string, error) {
	client := resty.New()

	resp, err := client.R().Get("https://api.ipify.org")
	if err != nil {
		return "", fmt.Errorf("connect error: %w", err)
	}

	return string(resp.Body()), nil
}
