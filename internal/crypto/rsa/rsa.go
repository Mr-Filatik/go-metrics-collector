package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

var ErrInvalidPEM = errors.New("invalid PEM")

// LoadPublicKey загружает публичный ключ из PEM-файла.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file error %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrInvalidPEM
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, ErrInvalidPEM
	}
	if k, ok := key.(*rsa.PublicKey); ok {
		return k, nil
	}
	return nil, ErrInvalidPEM
}

// LoadPrivateKey загружает приватный ключ из PEM-файла.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file error %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrInvalidPEM
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
			return k, nil
		}
	case "PRIVATE KEY":
		if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			if key, ok := k.(*rsa.PrivateKey); ok {
				return key, nil
			}
		}
	default:
		return nil, errors.New("unsuported key type")
	}
	return nil, ErrInvalidPEM
}

// Encrypt шифрует данные публичным ключом.
// Имеет ограничение из-за ключа на объём данных, равным 256 - 14 ~ 242 Б.
func Encrypt(data []byte, pub *rsa.PublicKey) ([]byte, error) {
	if pub == nil {
		return nil, errors.New("public key is nil")
	}

	res, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return res, nil
}

// Decrypt расшифровывает данные приватным ключом.
// Имеет ограничение из-за ключа на объём данных, равным 256 - 14 ~ 242 Б.
func Decrypt(data []byte, priv *rsa.PrivateKey) ([]byte, error) {
	if priv == nil {
		return nil, errors.New("private key is nil")
	}

	res, err := rsa.DecryptPKCS1v15(rand.Reader, priv, data)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return res, nil
}

const (
	encryptedKeySizeByte = 32
	encryptedKeySizeBit  = 8 * encryptedKeySizeByte
)

// EncryptBig шифрует данные публичным ключом.
// Не имеет ограничение из-за ключа на объём данных.
// Возвращает один слайс байт: encryptedKey + ciphertext.
func EncryptBig(data []byte, pub *rsa.PublicKey) ([]byte, error) {
	aesKey := make([]byte, encryptedKeySizeByte)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, fmt.Errorf("read full error %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("new cipher error %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new GCM error %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("read full error %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	encryptedKey, err := Encrypt(aesKey, pub)
	if err != nil {
		return nil, err
	}

	result := make([]byte, len(encryptedKey)+len(ciphertext))
	copy(result[:len(encryptedKey)], encryptedKey)
	copy(result[len(encryptedKey):], ciphertext)

	return result, nil
}

// DecryptBig расшифровывает данные приватным ключом.
// Не имеет ограничение из-за ключа на объём данных.
// Возвращает один слайс байт: encryptedKey + ciphertext.
func DecryptBig(combined []byte, priv *rsa.PrivateKey) ([]byte, error) {
	if len(combined) < encryptedKeySizeBit {
		return nil, errors.New("combined data too short")
	}

	var encryptedKey, ciphertext []byte

	encryptedKeySize := priv.Size()

	if len(combined) < encryptedKeySize {
		return nil, errors.New("encrypted key length mismatch")
	}

	encryptedKey = combined[:encryptedKeySize]
	ciphertext = combined[encryptedKeySize:]

	aesKey, err := Decrypt(encryptedKey, priv)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("new cipher error %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new GCM error %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce, cipherblob := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, cipherblob, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
