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

const (
	EncodingType = "gzip"
)

func Run(m *metric.AgentMetrics, endpoint string, reportInterval int64) {
	t := time.Tick(time.Duration(reportInterval) * time.Second)

	for range t {
		client := resty.New()

		client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
			if strings.Contains(r.Header.Get("Content-Encoding"), EncodingType) {
				body := r.Body
				if body != nil {
					compressedBody, err := compressBody(body.([]byte))
					if err != nil {
						log.Printf("Compress body error: %v", err.Error())
					} else {
						log.Printf("Compress body: %v -> %v", len(body.([]byte)), len(compressedBody))
						r.SetBody(compressedBody)
					}
				} else {
					log.Printf("Compress body error: body is empty")
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
			if strings.Contains(r.Header().Get("Content-Encoding"), EncodingType) {
				val, err := decompressBody(r.Body())
				if err != nil {
					if err.Error() == "gzip: invalid header" {
						log.Printf("Decompress body error: body not compress")
					} else {
						log.Printf("Decompress body error: %v", err.Error())
					}
				} else {
					log.Printf("Decompress body: %v -> %v", len(r.Body()), len(val))
					r.SetBody(val)
				}
			}
			return nil
		})

		for _, el := range m.GetAllGauge() {
			if num, err := strconv.ParseFloat(el.Value, 64); err == nil {
				metr := entity.Metrics{
					ID:    el.Name,
					MType: el.Type,
					Value: &num,
				}

				dat, rerr := json.Marshal(metr)
				if rerr != nil {
					log.Printf("Error on json create: %v.", rerr.Error())
					continue
				}

				address := endpoint + "/update/"
				log.Printf("Response to %v. (%v, %v, %v, %v)", address, metr.ID, metr.MType, *metr.Value, metr.Delta)
				resp, err := client.R().
					SetHeader("Content-Type", "application/json").
					SetHeader("Content-Encoding", EncodingType).
					SetHeader("Accept-Encoding", EncodingType).
					SetBody(dat).
					Post(address)

				if err != nil {
					log.Printf("Error on response: %v.", err.Error())
					continue
				}
				log.Printf("Response is done. StatusCode: %v. Data: %v.", resp.Status(), string(resp.Body()))
			}
		}

		for _, el := range m.GetAllCounter() {
			met := m.GetCounter(el)
			if num, err := strconv.ParseInt(met.Value, 10, 64); err == nil {
				metr := entity.Metrics{
					ID:    met.Name,
					MType: met.Type,
					Delta: &num,
				}

				dat, rerr := json.Marshal(metr)
				if rerr != nil {
					log.Printf("Error on json create: %v.", rerr.Error())
					continue
				}

				address := endpoint + "/update/"
				log.Printf("Response to %v. (%v, %v, %v, %v)", address, metr.ID, metr.MType, metr.Value, *metr.Delta)
				resp, rerr := client.R().
					SetHeader("Content-Type", "application/json").
					SetHeader("Content-Encoding", EncodingType).
					SetHeader("Accept-Encoding", EncodingType).
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
