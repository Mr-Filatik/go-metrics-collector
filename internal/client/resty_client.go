package client

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	crypto "github.com/Mr-Filatik/go-metrics-collector/internal/crypto/rsa"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repeater"
	"github.com/go-resty/resty/v2"
)

// RestyClient - клиент для отправки запросов к серверу.
type RestyClient struct {
	restyClient *resty.Client
	log         logger.Logger
	url         string
	xRealIP     string
}

var _ Client = (*RestyClient)(nil)

// RestyClientConfig - структура, содержащая основные параметры для RestyClient.
type RestyClientConfig struct {
	PublicKey *rsa.PublicKey
	URL       string
	XRealIP   string
	HashKey   string
}

// NewRestyClient создаёт новый экземпляр *RestyClient.
func NewRestyClient(config *RestyClientConfig, l logger.Logger) *RestyClient {
	client := &RestyClient{
		restyClient: resty.New(),
		url:         config.URL + "/updates/",
		xRealIP:     config.XRealIP,
		log:         l,
	}

	client.registerMiddlewares(config.HashKey, config.PublicKey)

	return client
}

// registerMiddlewares регистрирует все необходимые middleware для клиента.
func (client *RestyClient) registerMiddlewares(hashKey string, publicKey *rsa.PublicKey) {
	client.restyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		hashErr := client.hashingMiddleware(r, hashKey)
		if hashErr != nil {
			client.log.Error("Hashing body error", hashErr)
			return nil
		}

		comprErr := client.compressingMiddleware(r)
		if comprErr != nil {
			client.log.Error("Compressing body error", comprErr)
			return nil
		}

		encrErr := client.encryptingMiddleware(r, publicKey)
		if encrErr != nil {
			client.log.Error("Encryption body error", encrErr)
			return nil
		}

		// Логирование заголовков запроса.
		headers := r.Header
		hdrs := make([]interface{}, 0)
		for name := range headers {
			hdrs = append(hdrs, name, strings.Join(headers[name], " "))
		}
		client.log.Debug("Request headers", hdrs...)

		return nil
	})

	client.restyClient.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		// Логирование заголовков ответа.
		headers := r.Header()
		hdrs := make([]interface{}, 0)
		for name := range headers {
			hdrs = append(hdrs, name, strings.Join(headers[name], " "))
		}
		client.log.Debug("Response headers", hdrs...)

		decompErr := client.decompressingMiddleware(r)
		if decompErr != nil {
			client.log.Error("Decompress body error", decompErr)
			return nil
		}

		return nil
	})
}

// hashingMiddleware добавляет заголовок с хэшем тела запроса.
func (client *RestyClient) hashingMiddleware(r *resty.Request, hashKey string) error {
	if hashKey == "" {
		// Хеширование отключено
		return nil
	}

	body := r.Body
	if body == nil {
		return ErrEmptyBody
	}

	byteBody, ok := body.([]byte)
	if !ok {
		return ErrNotByteBody
	}

	h := hmac.New(sha256.New, []byte(hashKey))
	_, err := h.Write(byteBody)
	if err != nil {
		return fmt.Errorf("hash write error: %w", err)
	}
	hashBytes := h.Sum(nil)
	hashStr := hex.EncodeToString(hashBytes)

	r.Header.Set(HashHeader, hashStr)

	client.log.Debug("HashSHA256 added to request headers", "hash", hashStr)
	return nil
}

// compressingMiddleware сжимает тело запроса.
func (client *RestyClient) compressingMiddleware(r *resty.Request) error {
	if !strings.Contains(r.Header.Get(ContentEncodingHeader), EncodingType) {
		// Сжатие отключено
		return nil
	}

	body := r.Body
	if body == nil {
		return ErrEmptyBody
	}

	byteBody, ok := body.([]byte)
	if !ok {
		return ErrNotByteBody
	}

	compressedBody, err := compressBody(byteBody)
	if err != nil {
		return fmt.Errorf("compress error: %w", err)
	}
	if compressedBody == nil {
		return fmt.Errorf("compress error: %w", errors.New("compressed body is nil"))
	}

	r.SetBody(compressedBody)

	client.log.Debug("Compress body", "fromSize", len(byteBody), "toSize", len(compressedBody))
	return nil
}

// decompressingMiddleware расжимает тело ответа.
func (client *RestyClient) decompressingMiddleware(r *resty.Response) error {
	if !strings.Contains(r.Header().Get(AcceptEncodingHeader), EncodingType) {
		// Сжатие отключено
		return nil
	}

	val, err := decompressBody(r.Body())
	if err != nil {
		if err.Error() == "gzip: invalid header" {
			return fmt.Errorf("decompress error: %w", errors.New("body not compress"))
		} else {
			return fmt.Errorf("decompress error: %w", err)
		}
	}

	r.SetBody(val)

	client.log.Debug("Decompress body", "fromSize", len(r.Body()), "toSize", len(val))
	return nil
}

// encryptingMiddleware шифрует тело запроса.
func (client *RestyClient) encryptingMiddleware(r *resty.Request, publicKey *rsa.PublicKey) error {
	if publicKey == nil {
		// Шифрование отключено
		return nil
	}

	body := r.Body
	if body == nil {
		return ErrEmptyBody
	}

	byteBody, ok := body.([]byte)
	if !ok {
		return ErrNotByteBody
	}

	encrypted, err := crypto.EncryptBig(byteBody, publicKey)
	if err != nil {
		return fmt.Errorf("encrypt error: %w", err)
	}

	r.SetBody(encrypted)

	client.log.Debug("Encrypt body")
	return nil
}

func (client *RestyClient) SendMetric(m entity.Metrics) error {
	client.log.Warn("SendMetric(m entity.Metrics) not worked", nil)
	return nil
}

func (client *RestyClient) SendMetrics(ms []entity.Metrics) error {
	if len(ms) == 0 {
		client.log.Warn("Sending metrics is empty", nil)
		return nil
	}

	dat, err := json.Marshal(ms)
	if err != nil {
		return fmt.Errorf("JSON marshal error: %w", err)
	}

	resp, err := repeater.New[[]byte, *resty.Response](client.log).
		SetFunc(func(b []byte) (*resty.Response, error) {
			client.log.Info("Sending metrics", "url", client.url)
			resp, err := client.restyClient.R().
				SetHeader("Content-Type", "application/json").
				SetHeader(ContentEncodingHeader, EncodingType).
				SetHeader(AcceptEncodingHeader, EncodingType).
				SetHeader("X-Real-IP", client.xRealIP).
				SetBody(dat).
				Post(client.url)

			if err != nil {
				return resp, errors.New(err.Error())
			}
			return resp, nil
		}).
		Run(dat)

	if err != nil {
		return fmt.Errorf("sending metrics error: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		err := errors.New("responce status is " + resp.Status())
		return fmt.Errorf("responce status code not OK: %w", err)
	}

	client.log.Debug("Send metrics success", "url", client.url, "status", resp.Status())
	return nil
}
