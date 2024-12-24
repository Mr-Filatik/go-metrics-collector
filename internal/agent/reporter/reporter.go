package reporter

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/go-resty/resty/v2"
)

func Run(m *metric.AgentMetrics, endpoint string, reportInterval int64) {
	t := time.Tick(time.Duration(reportInterval) * time.Second)

	for range t {
		client := resty.New()
		for _, el := range m.GetAllGauge() {
			if num, err := strconv.ParseFloat(el.Value, 64); err == nil {
				metr := entity.Metrics{
					ID:    el.Name,
					MType: string(el.Type),
					Value: &num,
				}

				dat, rerr := json.Marshal(metr)
				if rerr != nil {
					log.Printf("Error on json create: %v.", rerr.Error())
					continue
				}

				address := endpoint + "/update/"
				log.Printf("Response to %v.", address)
				resp, err := client.R().
					SetHeader("Content-Type", " application/json").
					SetBody(dat).
					Post(address)

				if err != nil {
					log.Printf("Error on response: %v.", err.Error())
					continue
				}
				log.Printf("Response is done. StatusCode: %v.", resp.Status())
			}
		}

		for _, el := range m.GetAllCounter() {
			met := m.GetCounter(el)
			if num, err := strconv.ParseInt(met.Value, 10, 64); err == nil {
				metr := entity.Metrics{
					ID:    met.Name,
					MType: string(met.Type),
					Delta: &num,
				}

				dat, rerr := json.Marshal(metr)
				if rerr != nil {
					log.Printf("Error on json create: %v.", rerr.Error())
					continue
				}

				address := endpoint + "/update/"
				log.Printf("Response to %v.", address)
				resp, rerr := client.R().
					SetHeader("Content-Type", " application/json").
					SetBody(dat).
					Post(address)

				if rerr != nil {
					log.Printf("Error on response: %v.", rerr.Error())
					continue
				}
				m.ClearCounter(el)
				log.Printf("Response is done. StatusCode: %v.", resp.Status())
			}
		}
	}
}
