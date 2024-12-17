package main

import (
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/reporter"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/updater"
	config "github.com/Mr-Filatik/go-metrics-collector/internal/config/agent"
)

func main() {
	conf := config.Initialize()
	metrics := metric.New()

	go updater.Run(metrics, conf.PollInterval)
	reporter.Run(metrics, conf.ServerAddress, conf.ReportInterval)
}
