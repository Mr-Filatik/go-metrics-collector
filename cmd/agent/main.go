package main

import (
	"net/http"
	_ "net/http/pprof"

	config "github.com/Mr-Filatik/go-metrics-collector/internal/agent/config"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/reporter"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/updater"
	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
)

func main() {
	log := logger.New(logger.LevelInfo)
	defer log.Close()

	conf := config.Initialize()
	metrics := metric.New()

	go updater.Run(metrics, conf.PollInterval)
	go updater.RunMemory(metrics, conf.PollInterval)
	go reporter.Run(metrics, conf.ServerAddress, conf.ReportInterval, conf.HashKey, conf.RateLimit, log)

	err := http.ListenAndServe("localhost:8081", nil)
	if err != nil {
		log.Error("Server error", err)
	}
}
