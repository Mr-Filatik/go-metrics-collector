package main

import (
	"log"

	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/reporter"
	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/updater"
	config "github.com/Mr-Filatik/go-metrics-collector/internal/config/agent"
	"github.com/go-resty/resty/v2"
)

func main() {
	conf := config.Initialize()
	metrics := metric.New()

	client := resty.New()
	address := conf.ServerAddress + "/"
	_, err := client.R().
		SetHeader("Content-Type", " application/json").
		Get(address)

	if err != nil {
		log.Printf("Error on response: %v.", err.Error())
	}

	go updater.Run(metrics, conf.PollInterval)
	reporter.Run(metrics, conf.ServerAddress, conf.ReportInterval)
}
