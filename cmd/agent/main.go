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

func RunUpdater(m *metric.Metric, pollInterval int64) {

	t := time.Tick(time.Duration(pollInterval) * time.Second)

	for range t {
		m.Update()
	}
}

func RunReporter(m *metric.Metric, endpoint string, reportInterval int64) {

	t := time.Tick(time.Duration(reportInterval) * time.Second)

	for range t {
		client := resty.New()
		err := m.Foreach(func(metricType, metricName, metricValue string) error {
			address := endpoint + "/update/" + metricType + "/" + metricName + "/" + metricValue

			log.Printf("Response to %v.", address)

			resp, err := client.R().
				SetHeader("Content-Type", " text/plain").
				Post(address)

			if err != nil {
				log.Printf("Error on response: %v.", err.Error())
				return err
			}
			log.Printf("Response is done. StatusCode: %v.", resp.Status())
			return nil
		})
		if err != nil {
			log.Printf("Send error count: %v.", err)
		} else {
			log.Printf("Send is done.")
		}
	}
}
