package main

import (
	"log"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/agent/config"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/agent/metric"
	"github.com/go-resty/resty/v2"
)

func main() {
	config := config.Initialize()
	metrics := metric.New()

	go RunUpdater(metrics, config.PollInterval)
	RunReporter(metrics, config.ServerAddress, config.ReportInterval)
}

func RunUpdater(m *metric.AgentMetrics, pollInterval int64) {
	t := time.Tick(time.Duration(pollInterval) * time.Second)

	for range t {
		m.Update()
	}
}

func RunReporter(m *metric.AgentMetrics, endpoint string, reportInterval int64) {
	t := time.Tick(time.Duration(reportInterval) * time.Second)

	for range t {
		client := resty.New()
		for _, el := range m.GetAll(true) {
			address := endpoint + "/update/" + string(el.Type) + "/" + el.Name + "/" + el.Value
			log.Printf("Response to %v.", address)
			resp, err := client.R().
				SetHeader("Content-Type", " text/plain").
				Post(address)

			if err != nil {
				log.Printf("Error on response: %v.", err.Error())
			}
			log.Printf("Response is done. StatusCode: %v.", resp.Status())
		}
	}
}
