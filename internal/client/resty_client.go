package client

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Mr-Filatik/go-metrics-collector/internal/common"
	crypto "github.com/Mr-Filatik/go-metrics-collector/internal/crypto/rsa"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repeater"
	"github.com/go-resty/resty/v2"
)

// RestyClient - клиент для отправки запросов к серверу.
type RestyClient struct {
	restyClient *resty.Client
	publicKey   *rsa.PublicKey
	log         logger.Logger
	url         string
	xRealIP     string
	hashKey     string
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
		url:       config.URL + "/updates/",
		xRealIP:   config.XRealIP,
		log:       l,
		publicKey: config.PublicKey,
		hashKey:   config.HashKey,
	}

	return client
}

func (c *RestyClient) Start(_ context.Context) error {
	c.log.Info("Start RestyClient...")
	c.restyClient = resty.New()
	c.registerMiddlewares(c.hashKey, c.publicKey)
	c.log.Info("Start RestyClient is successfull.")
	return nil
}

func (c *RestyClient) SendMetric(_ context.Context, m entity.Metrics) error {
	if c.restyClient == nil {
		err := fmt.Errorf("RestyClient: %w", ErrClientNotStarted)
		c.log.Error("Error in *RestyClient.SendMetric().", err)
		return err
	}

	c.log.Warn("Not implemented *RestyClient.SendMetric().", nil)
	return nil
}

func (c *RestyClient) SendMetrics(_ context.Context, ms []entity.Metrics) error {
	if c.restyClient == nil {
		err := fmt.Errorf("RestyClient: %w", ErrClientNotStarted)
		c.log.Error("Error in *RestyClient.SendMetrics().", err)
		return err
	}

	if len(ms) == 0 {
		c.log.Warn("Sending metrics is empty.", nil)
		return nil
	}

	dat, err := json.Marshal(ms)
	if err != nil {
		return fmt.Errorf("JSON marshal error: %w", err)
	}

	resp, err := repeater.New[[]byte, *resty.Response](c.log).
		SetFunc(func(b []byte) (*resty.Response, error) {
			c.log.Info("Sending metrics", "url", c.url)
			resp, err := c.restyClient.R().
				SetHeader(common.HeaderContentType, common.HeaderContentTypeValueApplicationJSON).
				SetHeader(common.HeaderContentEncoding, common.HeaderEncodingValueGZIP).
				SetHeader(common.HeaderAcceptEncoding, common.HeaderEncodingValueGZIP).
				SetHeader(common.HeaderXRealIP, c.xRealIP).
				SetBody(dat).
				Post(c.url)

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

	c.log.Debug("Send metrics success", "url", c.url, "status", resp.Status())
	return nil
}

func (c *RestyClient) Close() error {
	if c.restyClient == nil {
		err := fmt.Errorf("RestyClient: %w", ErrClientNotStarted)
		c.log.Error("Error in *RestyClient.Close().", err)
		return err
	}

	c.log.Warn("Not implemented *RestyClient.Close().", nil)
	return nil
}

// registerMiddlewares регистрирует все необходимые middleware для клиента.
func (c *RestyClient) registerMiddlewares(hashKey string, publicKey *rsa.PublicKey) {
	c.restyClient.OnBeforeRequest(func(cc *resty.Client, r *resty.Request) error {
		hashErr := c.hashingMiddleware(r, hashKey)
		if hashErr != nil {
			c.log.Error("Hashing body error", hashErr)
			return nil
		}

		comprErr := c.compressingMiddleware(r)
		if comprErr != nil {
			c.log.Error("Compressing body error", comprErr)
			return nil
		}

		encrErr := c.encryptingMiddleware(r, publicKey)
		if encrErr != nil {
			c.log.Error("Encryption body error", encrErr)
			return nil
		}

		// Логирование заголовков запроса.
		hdrs := convertHeadersToSlice(r.Header)
		c.log.Debug("Request headers", hdrs...)

		return nil
	})

	c.restyClient.OnAfterResponse(func(cc *resty.Client, r *resty.Response) error {
		// Логирование заголовков ответа.
		hdrs := convertHeadersToSlice(r.Header())
		c.log.Debug("Response headers", hdrs...)

		decompErr := c.decompressingMiddleware(r)
		if decompErr != nil {
			c.log.Error("Decompress body error", decompErr)
			return nil
		}

		return nil
	})
}

// hashingMiddleware добавляет заголовок с хэшем тела запроса.
func (c *RestyClient) hashingMiddleware(r *resty.Request, hashKey string) error {
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

	hashStr, err := common.HashBytesToString(byteBody, hashKey)
	if err != nil {
		return fmt.Errorf("calculate hash error: %w", err)
	}

	r.Header.Set(common.HeaderHashSHA256, hashStr)

	c.log.Debug("HashSHA256 added to request headers")
	return nil
}

// compressingMiddleware сжимает тело запроса.
func (c *RestyClient) compressingMiddleware(r *resty.Request) error {
	if !strings.Contains(r.Header.Get(common.HeaderContentEncoding), common.HeaderEncodingValueGZIP) {
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

	compressedBody, err := common.CompressBytes(byteBody)
	if err != nil {
		return fmt.Errorf("compress error: %w", err)
	}
	// if compressedBody == nil {
	// 	return fmt.Errorf("compress error: %w", errors.New("compressed body is nil"))
	// }

	r.SetBody(compressedBody)

	c.log.Debug("Compress body", "fromSize", len(byteBody), "toSize", len(compressedBody))
	return nil
}

// decompressingMiddleware расжимает тело ответа.
func (c *RestyClient) decompressingMiddleware(r *resty.Response) error {
	if !strings.Contains(r.Header().Get(common.HeaderAcceptEncoding), common.HeaderEncodingValueGZIP) {
		// Сжатие отключено
		return nil
	}

	val, err := common.DecompressBytes(r.Body())
	if err != nil {
		if err.Error() == "gzip: invalid header" {
			return fmt.Errorf("decompress error: %w", errors.New("body not compress"))
		} else {
			return fmt.Errorf("decompress error: %w", err)
		}
	}

	r.SetBody(val)

	c.log.Debug("Decompress body", "fromSize", len(r.Body()), "toSize", len(val))
	return nil
}

// encryptingMiddleware шифрует тело запроса.
func (c *RestyClient) encryptingMiddleware(r *resty.Request, publicKey *rsa.PublicKey) error {
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

	c.log.Debug("Encrypt body")
	return nil
}

func convertHeadersToSlice(headers http.Header) []interface{} {
	sensitiveHeaders := map[string]struct{}{
		"authorization": {},
		"cookie":        {},
		"x-api-key":     {},
		"x-api-secret":  {},
		"set-cookie":    {},
		"hashsha256":    {},
	}

	hdrs := make([]interface{}, 0)
	for name := range headers {
		lowerName := strings.ToLower(name)
		if _, ok := sensitiveHeaders[lowerName]; ok {
			hdrs = append(hdrs, name, "***")
		} else {
			hdrs = append(hdrs, name, strings.Join(headers[name], " "))
		}
	}

	return hdrs
}
