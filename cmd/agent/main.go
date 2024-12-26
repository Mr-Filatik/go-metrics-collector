package main

import (
	config "github.com/Mr-Filatik/go-metrics-collector/internal/agent/config"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/reporter"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/updater"
)

func main() {
	conf := config.Initialize()
	metrics := metric.New()

	go updater.Run(metrics, conf.PollInterval)
	reporter.Run(metrics, conf.ServerAddress, conf.ReportInterval)
}
