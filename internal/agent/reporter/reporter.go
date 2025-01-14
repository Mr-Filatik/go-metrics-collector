package reporter

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/go-resty/resty/v2"
)

func Run(m *metric.AgentMetrics, endpoint string, reportInterval int64) {
	t := time.Tick(time.Duration(reportInterval) * time.Second)

	for range t {
		client := resty.New()

		client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
			body := r.Body
			log.Printf("Request body: %v  (%v)", len(body.([]byte)), string(body.([]byte)))
			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				if body != nil {
					compressedBody, err := compressBody(body.([]byte))
					if err != nil {
						log.Printf("Request compress body error: %v", err.Error())
					} else {
						log.Printf("Request compress body: %v (%v)", len(compressedBody), string(compressedBody))
						r.SetBody(compressedBody)
						r.Header.Add("Content-Length", strconv.FormatInt((int64(len(compressedBody))), 10))
					}
				} else {
					log.Printf("Request compress body error: body is empty")
				}
			}
			for name, values := range r.Header {
				for _, value := range values {
					log.Printf("Request header: %v: %v", name, value)
				}
			}
			return nil
		})

		client.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
			for name, values := range r.Header() {
				for _, value := range values {
					log.Printf("Response header: %v: %v", name, value)
				}
			}
			body := r.Body()
			log.Printf("Response body: %v (%v)", len(body), string(body))
			if strings.Contains(r.Header().Get("Content-Encoding"), "gzip") {
				val, err := decompressBody(body)
				if err != nil {
					if err.Error() == "gzip: invalid header" {
						log.Printf("Response decompress body error: body not compress")
					} else {
						log.Printf("Response decompress body error: %v", err.Error())
					}
				} else {
					log.Printf("Response decompress body: %v (%v)", len(val), string(val))
					r.SetBody(val)
				}
			}
			return nil
		})

		// for _, el := range m.GetAllGauge() {
		// 	if num, err := strconv.ParseFloat(el.Value, 64); err == nil {
		// 		metr := entity.Metrics{
		// 			ID:    el.Name,
		// 			MType: el.Type,
		// 			Value: &num,
		// 		}

		// 		dat, rerr := json.Marshal(metr)
		// 		if rerr != nil {
		// 			log.Printf("Error on json create: %v.", rerr.Error())
		// 			continue
		// 		}

		// 		address := endpoint + "/update/"
		// 		log.Printf("Response to %v. (%v, %v, %v, %v)", address, metr.ID, metr.MType, *metr.Value, metr.Delta)
		// 		resp, err := client.R().
		// 			SetHeader("Content-Type", " application/json").
		// 			SetBody(dat).
		// 			Post(address)

		// 		if err != nil {
		// 			log.Printf("Error on response: %v.", err.Error())
		// 			continue
		// 		}
		// 		log.Printf("Response is done. StatusCode: %v. Data: %v.", resp.Status(), string(resp.Body()))
		// 	}
		// }

		for _, el := range m.GetAllCounter() {
			met := m.GetCounter(el)
			if num, err := strconv.ParseInt(met.Value, 10, 64); err == nil {
				metr := entity.Metrics{
					ID:    met.Name,
					MType: met.Type,
					Delta: &num,
				}

				dat, err := json.Marshal(metr)
				if err != nil {
					log.Printf("Error on json create: %v.", err.Error())
					continue
				}

				address := endpoint + "/update/"
				log.Printf("Response to %v. (%v, %v, %v, %v)", address, metr.ID, metr.MType, metr.Value, *metr.Delta)
				resp, rerr := client.R().
					SetHeader("Content-Type", "application/json").
					SetHeader("Content-Encoding", "gzip").
					SetHeader("Accept-Encoding", "gzip").
					SetBody(dat).
					Post(address)

				if rerr != nil {
					log.Printf("Error on response: %v.", rerr.Error())
					continue
				}

				m.ClearCounter(el)
				log.Printf("Response is done. StatusCode: %v. Data: %v.", resp.Status(), string(resp.Body()))
			}
		}
	}
}

func compressBody(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, gerr := gw.Write(data)
	if gerr != nil {
		return nil, errors.New(gerr.Error())
	}
	if err := gw.Close(); err != nil {
		return nil, errors.New(err.Error())
	}
	return buf.Bytes(), nil
}

func decompressBody(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, gr)
	if err != nil {
		return nil, err
	}
	if err := gr.Close(); err != nil {
		return nil, errors.New(err.Error())
	}
	return buf.Bytes(), nil
}
