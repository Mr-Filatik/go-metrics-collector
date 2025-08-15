package common

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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
