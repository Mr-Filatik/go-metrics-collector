package client

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
)

const (
	EncodingType          = "gzip"             // тип кодирования
	ContentEncodingHeader = "Content-Encoding" // заголовок кодирования контента
	AcceptEncodingHeader  = "Accept-Encoding"  // заголовок поддерживаемого кодирования
	HashHeader            = "HashSHA256"       // заголовок хеширования содержимого
)

var (
	ErrEmptyBody   = errors.New("body is empty")
	ErrNotByteBody = errors.New("body is not of type []byte")
)

// Client - интерфейс для всех клиентов приложения.
type Client interface {
	io.Closer
	SendMetric(m entity.Metrics) error
	SendMetrics(ms []entity.Metrics) error
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
