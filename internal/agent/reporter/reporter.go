package reporter

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repeater"
	"github.com/go-resty/resty/v2"
)

const (
	EncodingType          = "gzip"
	ContentEncodingHeader = "Content-Encoding"
	AcceptEncodingHeader  = "Accept-Encoding"
	HashHeader            = "HashSHA256"
)

func Run(m *metric.AgentMetrics, endpoint string, reportInterval int64, hashKey string, log logger.Logger) {
	t := time.Tick(time.Duration(reportInterval) * time.Second)

	for range t {
		client := resty.New()

		client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
			if hashKey != "" {
				body := r.Body
				if body != nil {
					byteBody, ok := body.([]byte)
					if !ok {
						log.Error("Hashing body error", errors.New("body is not of type []byte"))
						return nil
					}

					hash := sha256.Sum256(byteBody)
					hashStr := hex.EncodeToString(hash[:])

					r.Header.Set(HashHeader, hashStr)
					log.Debug("HashSHA256 added to request headers", "hash", hashStr)
				} else {
					log.Error("Hashing body error", errors.New("body is empty"))
				}
			}

			if strings.Contains(r.Header.Get(ContentEncodingHeader), EncodingType) {
				body := r.Body
				if body != nil {
					byteBody, ok := body.([]byte)
					if !ok {
						log.Error("Compress body error", errors.New("body does not exist"))
						return nil
					}
					compressedBody, err := compressBody(byteBody)
					if err != nil {
						log.Error("Compress body error", err)
						return nil
					}
					if compressedBody != nil {
						log.Debug("Compress body",
							"fromSize", len(byteBody),
							"toSize", len(compressedBody))
						r.SetBody(compressedBody)
					}
				} else {
					log.Error("Compress body error", errors.New("body is empty"))
				}
			}

			for name, values := range r.Header {
				for _, value := range values {
					log.Debug("Request header",
						"name", name,
						"value", value)
				}
			}
			return nil
		})

		client.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
			for name, values := range r.Header() {
				for _, value := range values {
					log.Debug("Response header",
						"name", name,
						"value", value)
				}
			}
			if strings.Contains(r.Header().Get(AcceptEncodingHeader), EncodingType) {
				val, err := decompressBody(r.Body())
				if err != nil {
					if err.Error() == "gzip: invalid header" {
						log.Error("Decompress body error", errors.New("body not compress"))
					} else {
						log.Error("Decompress body error", err)
					}
					return nil
				}
				log.Debug("Decompress body",
					"fromSize", len(r.Body()),
					"toSize", len(val))
				r.SetBody(val)
			}
			return nil
		})

		var metrics []entity.Metrics
		for _, el := range m.GetAllGauge() {
			if num, err := strconv.ParseFloat(el.Value, 64); err == nil {
				mc := entity.Metrics{
					ID:    el.Name,
					MType: el.Type,
					Value: &num,
				}
				metrics = append(metrics, mc)
			}
		}
		for _, el := range m.GetAllCounter() {
			met := m.GetCounter(el)
			if num, err := strconv.ParseInt(met.Value, 10, 64); err == nil {
				mc := entity.Metrics{
					ID:    met.Name,
					MType: met.Type,
					Delta: &num,
				}
				metrics = append(metrics, mc)
			}
		}

		dat, rerr := json.Marshal(metrics)
		if rerr != nil {
			log.Error("Error on json create", rerr)
			continue
		}

		address := endpoint + "/updates/"

		resp, rerr := repeater.New[[]byte, *resty.Response](log).
			SetFunc(func(b []byte) (*resty.Response, error) {
				log.Info("Response to server",
					"address", address,
					"count", len(metrics))
				resp, rerr := client.R().
					SetHeader("Content-Type", "application/json").
					SetHeader(ContentEncodingHeader, EncodingType).
					SetHeader(AcceptEncodingHeader, EncodingType).
					SetBody(dat).
					Post(address)

				if rerr != nil {
					return resp, errors.New(rerr.Error())
				}
				return resp, nil
			}).
			// SetCondition(...).
			Run(dat)

		if rerr != nil {
			log.Error("Response error", rerr)
			continue
		}

		for _, el := range m.GetAllCounter() {
			m.ClearCounter(el)
		}

		log.Info("Response success",
			"statusCode", resp.Status(),
			"data", string(resp.Body()))
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
		return nil, errors.New(err.Error())
	}
	var buf bytes.Buffer
	_, err = io.CopyN(&buf, gr, int64(buf.Len()))
	if err != nil {
		return nil, errors.New(err.Error())
	}
	if err := gr.Close(); err != nil {
		return nil, errors.New(err.Error())
	}
	return buf.Bytes(), nil
}
