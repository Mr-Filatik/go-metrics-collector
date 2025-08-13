// Пакет reporter предоставляет реализацию воркера для отправки метрик на сервер.
// Пакет использует клиент resty, поддерживает отправку наборами данных и их сжатие по алгоритму gzip.
package reporter

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	crypto "github.com/Mr-Filatik/go-metrics-collector/internal/crypto/rsa"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repeater"
	"github.com/go-resty/resty/v2"
)

const (
	EncodingType          = "gzip"             // тип кодирования
	ContentEncodingHeader = "Content-Encoding" // заголовок кодирования контента
	AcceptEncodingHeader  = "Accept-Encoding"  // заголовок поддерживаемого кодирования
	HashHeader            = "HashSHA256"       // заголовок хеширования содержимого
)

// Run запускает цикл отправки метрик на удалённый сервер.
// Создаёт пул воркеров и посылает сигналы на отправку каждые reportInterval секунд.
//
// Параметры:
//   - ctx: контекст для отмены
//   - m: объект метрик (AgentMetrics)
//   - endpoint: адрес сервера, куда отправляются метрики
//   - reportInterval: интервал отправки метрик (в секундах)
//   - hashKey: ключ для хэширования метрик
//   - lim: количество параллельных воркеров
//   - log: логгер
func Run(
	ctx context.Context,
	m *metric.AgentMetrics,
	endpoint string,
	reportInterval int64,
	hashKey string,
	lim int64,
	publicKey *rsa.PublicKey,
	realIP string,
	log logger.Logger) {
	jobs := make(chan struct{}, lim)
	defer close(jobs)

	for w := int64(1); w <= lim; w++ {
		go worker(ctx, m, endpoint, hashKey, publicKey, realIP, log, jobs)
	}

	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case jobs <- struct{}{}:
			default:
				// log.Warn("Job queue full, skipping report", "queue_size", lim)
			}
		}
	}
}

func worker(
	ctx context.Context,
	m *metric.AgentMetrics,
	endpoint string,
	hashKey string,
	publicKey *rsa.PublicKey,
	realIP string,
	log logger.Logger,
	jobs <-chan struct{},
) {
	client := resty.New()

	// Middleware: Before Request
	client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		if hashKey != "" {
			body := r.Body
			if body != nil {
				byteBody, ok := body.([]byte)
				if !ok {
					log.Error("Hashing body error", errors.New("body is not of type []byte"))
					return nil
				}

				h := hmac.New(sha256.New, []byte(hashKey))
				h.Write(byteBody)
				hashBytes := h.Sum(nil)
				hashStr := hex.EncodeToString(hashBytes)

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

		if publicKey != nil {
			body := r.Body
			if body != nil {
				byteBody, ok := body.([]byte)
				if !ok {
					log.Error("Encrypt body error", errors.New("body does not exist"))
					return nil
				}
				encrypted, err := crypto.EncryptBig(byteBody, publicKey)
				if err != nil {
					log.Error("Encryption failed", err)
					return nil
				}
				r.SetBody(encrypted)
			} else {
				log.Error("Encryption body error", errors.New("body is empty"))
			}
		}

		headers := r.Header
		for name := range r.Header {
			for i := range headers[name] {
				log.Debug("Request header",
					"name", name,
					"value", headers[name][i])
			}
		}
		return nil
	})

	// Middleware: After Response
	client.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		headers := r.Header()
		for name := range headers {
			for i := range headers[name] {
				log.Debug("Response header",
					"name", name,
					"value", headers[name][i])
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

	for {
		select {
		case <-ctx.Done():
			return
		case <-jobs:
			var metrics []entity.Metrics
			gMetrics := m.GetAllGaugeNames()
			for _, name := range gMetrics {
				met := m.GetByName(name)
				if num, err := strconv.ParseFloat(met.Value, 64); err == nil {
					metrics = append(metrics, entity.Metrics{
						ID:    met.Name,
						MType: met.Type,
						Value: &num,
					})
				}
			}
			cMetrics := m.GetAllCounterNames()
			for _, name := range cMetrics {
				met := m.GetByName(name)
				if num, err := strconv.ParseInt(met.Value, 10, 64); err == nil {
					metrics = append(metrics, entity.Metrics{
						ID:    met.Name,
						MType: met.Type,
						Delta: &num,
					})
				}
			}

			if len(metrics) == 0 {
				continue
			}

			dat, err := json.Marshal(metrics)
			if err != nil {
				log.Error("JSON marshal error", err)
				continue
			}

			address := endpoint + "/updates/"

			resp, err := repeater.New[[]byte, *resty.Response](log).
				SetFunc(func(b []byte) (*resty.Response, error) {
					log.Info("Sending metrics", "count", len(metrics), "url", address)
					resp, err := client.R().
						SetHeader("Content-Type", "application/json").
						SetHeader(ContentEncodingHeader, EncodingType).
						SetHeader(AcceptEncodingHeader, EncodingType).
						SetHeader("X-Real-IP", realIP).
						SetBody(dat).
						Post(address)

					if err != nil {
						return resp, errors.New(err.Error())
					}
					return resp, nil
				}).
				Run(dat)

			if err != nil {
				log.Error("Send failed", err)
				continue
			}

			for _, name := range cMetrics {
				m.ClearCounter(name)
			}

			log.Info("Send success", "status", resp.Status(), "count", len(metrics))
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
