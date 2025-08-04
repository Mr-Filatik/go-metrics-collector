package main

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	config "github.com/Mr-Filatik/go-metrics-collector/internal/agent/config"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/reporter"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/updater"
	crypto "github.com/Mr-Filatik/go-metrics-collector/internal/crypto/rsa"
	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
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

	go updater.Run(metrics, conf.PollInterval)
	go updater.RunMemory(metrics, conf.PollInterval)
	go reporter.Run(
		metrics,
		conf.ServerAddress,
		conf.ReportInterval,
		conf.HashKey,
		conf.RateLimit,
		key,
		log)

	err := http.ListenAndServe("localhost:8081", nil)
	if err != nil {
		log.Error("Server error", err)
	}
}
