package reporter

import (
	"log"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	"github.com/go-resty/resty/v2"
)

func Run(m *metric.AgentMetrics, endpoint string, reportInterval int64) {
	t := time.Tick(time.Duration(reportInterval) * time.Second)

	for range t {
		client := resty.New()
		for _, el := range m.GetAllGauge() {
			address := endpoint + "/update/" + string(el.Type) + "/" + el.Name + "/" + el.Value
			log.Printf("Response to %v.", address)
			resp, err := client.R().
				SetHeader("Content-Type", " text/plain").
				Post(address)

			if err != nil {
				log.Printf("Error on response: %v.", err.Error())
				continue
			}
			log.Printf("Response is done. StatusCode: %v.", resp.Status())
		}

		for _, el := range m.GetAllCounter() {
			met := m.GetCounter(el)
			address := endpoint + "/update/" + string(met.Type) + "/" + met.Name + "/" + met.Value
			log.Printf("Response to %v.", address)
			resp, err := client.R().
				SetHeader("Content-Type", " text/plain").
				Post(address)

			if err != nil {
				log.Printf("Error on response: %v.", err.Error())
				continue
			}
			m.ClearCounter(el)
			log.Printf("Response is done. StatusCode: %v.", resp.Status())
		}
	}
}
