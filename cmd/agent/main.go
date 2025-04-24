package main

import (
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
	reporter.Run(metrics, conf.ServerAddress, conf.ReportInterval, conf.HashKey, conf.RateLimit, log)
}
