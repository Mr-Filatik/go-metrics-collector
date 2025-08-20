package common

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// CompressBytes выполняет сжатие []byte в формате GZIP.
func CompressBytes(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, writeErr := writer.Write(data)
	if writeErr != nil {
		return nil, fmt.Errorf("gzip write data error: %w", writeErr)
	}
	if closeErr := writer.Close(); closeErr != nil {
		return nil, fmt.Errorf("gzip writer close error: %w", closeErr)
	}
	return buf.Bytes(), nil
}

// DecompressBytes выполняет декомпрессию []byte форматом GZIP.
func DecompressBytes(data []byte) ([]byte, error) {
	reader, readerErr := gzip.NewReader(bytes.NewReader(data))
	if readerErr != nil {
		return nil, fmt.Errorf("gzip read data error: %w", readerErr)
	}
	var buf bytes.Buffer
	_, copyErr := io.CopyN(&buf, reader, int64(buf.Len()))
	if copyErr != nil {
		return nil, fmt.Errorf("gzip buffer copy error: %w", copyErr)
	}
	if closeErr := reader.Close(); closeErr != nil {
		return nil, fmt.Errorf("gzip reader close error: %w", closeErr)
	}
	return buf.Bytes(), nil
}

// HashBytesToString хэширует []byte ключём.
func HashBytesToString(data []byte, key string) (string, error) {
	hasher := hmac.New(sha256.New, []byte(key))
	_, writeErr := hasher.Write(data)
	if writeErr != nil {
		return "", fmt.Errorf("hmac hash write error: %w", writeErr)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// HashValidateStrings сравнивает два хэша.
func HashValidateStrings(hash1 string, hash2 string) bool {
	return hmac.Equal([]byte(hash1), []byte(hash2))
}

const portPrefix = 10000

// ChangePortForGRPC увеличивает порт на 10000
func ChangePortForGRPC(address string) (string, error) {
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "http://" + address
	}

	parsedURL, err := url.Parse(address)
	if err != nil {
		return "", fmt.Errorf("parse string error: %w", err)
	}

	host, portStr, err := net.SplitHostPort(parsedURL.Host)
	if err != nil {
		return "", fmt.Errorf("error format host:port: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", fmt.Errorf("port in not number: %w", err)
	}

	newPort := portPrefix + port

	return fmt.Sprintf("%s:%d", host, newPort), nil
}
