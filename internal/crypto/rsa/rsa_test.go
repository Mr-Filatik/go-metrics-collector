package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKeys() (pubPath, privPKCS1, privPKCS8 string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", errors.New(err.Error())
	}
	publicKey := &privateKey.PublicKey

	pubFile, _ := os.CreateTemp("", "public_*.pem")
	priv1File, _ := os.CreateTemp("", "private_pkcs1_*.pem")
	priv8File, _ := os.CreateTemp("", "private_pkcs8_*.pem")

	pubBytes, _ := x509.MarshalPKIXPublicKey(publicKey)
	_ = pem.Encode(pubFile, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	// PKCS#1
	pkcs1Bytes := x509.MarshalPKCS1PrivateKey(privateKey)
	_ = pem.Encode(priv1File, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: pkcs1Bytes})

	// PKCS#8
	pkcs8Bytes, _ := x509.MarshalPKCS8PrivateKey(privateKey)
	_ = pem.Encode(priv8File, &pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes})

	if err := pubFile.Close(); err != nil {
		return "", "", "", fmt.Errorf("failed to close public key file: %w", err)
	}
	if err := priv1File.Close(); err != nil {
		return "", "", "", fmt.Errorf("failed to close PKCS#1 private key file: %w", err)
	}
	if err := priv8File.Close(); err != nil {
		return "", "", "", fmt.Errorf("failed to close PKCS#8 private key file: %w", err)
	}

	return pubFile.Name(), priv1File.Name(), priv8File.Name(), nil
}

func cleanup(paths ...string) error {
	var errs []error

	for _, p := range paths {
		if err := os.Remove(p); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func TestLoadPublicKey_Valid(t *testing.T) {
	pubPath, priv1, priv8, err := generateTestKeys()
	require.NoError(t, err)
	defer func() {
		err := cleanup(pubPath, priv1, priv8)
		assert.NoError(t, err)
	}()

	key, err := LoadPublicKey(pubPath)
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, 2048, key.Size()*8) // 2048 бит
}

func TestLoadPublicKey_FileNotFound(t *testing.T) {
	_, err := LoadPublicKey("/tmp/nonexistent.pem")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read file error")
}

func TestLoadKey_InvalidPEM(t *testing.T) {
	f, _ := os.CreateTemp("", "invalid_*.pem")
	defer func() {
		err := os.Remove(f.Name())
		assert.NoError(t, err)
	}()
	defer func() {
		err := f.Close()
		assert.NoError(t, err)
	}()

	_, err := f.WriteString("-----BEGIN PUBLIC KEY-----\n")
	assert.NoError(t, err)
	_, err = f.WriteString("invalid base64 data\n")
	assert.NoError(t, err)
	_, err = f.WriteString("-----END PUBLIC KEY-----\n")
	assert.NoError(t, err)

	key, err := LoadPublicKey(f.Name())
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPEM, err)
	assert.Nil(t, key)

	fp, _ := os.CreateTemp("", "invalid_priv_*.pem")
	defer func() {
		err := os.Remove(fp.Name())
		assert.NoError(t, err)
	}()
	defer func() {
		err := fp.Close()
		assert.NoError(t, err)
	}()

	_, err = fp.WriteString("-----BEGIN PRIVATE KEY-----\n")
	assert.NoError(t, err)
	_, err = fp.WriteString("invalid base64\n")
	assert.NoError(t, err)
	_, err = fp.WriteString("-----END PRIVATE KEY-----\n")
	assert.NoError(t, err)

	keyp, err := LoadPrivateKey(fp.Name())
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPEM, err)
	assert.Nil(t, keyp)
}

func TestLoadPublicKey_NotRSA(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubBytes, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)

	f, _ := os.CreateTemp("", "ecdsa_*.pem")
	defer func() {
		err := os.Remove(f.Name())
		assert.NoError(t, err)
	}()
	defer func() {
		err := f.Close()
		assert.NoError(t, err)
	}()

	encodeErr := pem.Encode(f, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	assert.NoError(t, encodeErr)

	key, err := LoadPublicKey(f.Name())
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPEM, err)
	assert.Nil(t, key)
}

func TestLoadPrivateKey_PKCS1_Valid(t *testing.T) {
	pubPath, priv1, priv8, err := generateTestKeys()
	require.NoError(t, err)
	defer func() {
		err := cleanup(pubPath, priv1, priv8)
		assert.NoError(t, err)
	}()

	key, err := LoadPrivateKey(priv1)
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, 2048, key.Size()*8)
}

func TestLoadPrivateKey_PKCS8_Valid(t *testing.T) {
	pubPath, priv1, priv8, err := generateTestKeys()
	require.NoError(t, err)
	defer func() {
		err := cleanup(pubPath, priv1, priv8)
		assert.NoError(t, err)
	}()

	key, err := LoadPrivateKey(priv8)
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, 2048, key.Size()*8)
}

func TestLoadPrivateKey_UnsupportedType(t *testing.T) {
	f, _ := os.CreateTemp("", "unsupported_*.pem")
	defer func() {
		err := os.Remove(f.Name())
		assert.NoError(t, err)
	}()
	defer func() {
		err := f.Close()
		assert.NoError(t, err)
	}()

	_, err := f.WriteString("-----BEGIN CERTIFICATE-----\n")
	assert.NoError(t, err)
	_, err = f.WriteString("fake cert\n")
	assert.NoError(t, err)
	_, err = f.WriteString("-----END CERTIFICATE-----\n")
	assert.NoError(t, err)

	key, err := LoadPrivateKey(f.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsuported key type")
	assert.Nil(t, key)
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey

	plaintext := []byte("Hello, secure world!")

	ciphertext, err := Encrypt(plaintext, publicKey)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := Decrypt(ciphertext, privateKey)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptBig_RoundTrip(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey

	plaintext := []byte("Hello, secure world!")

	ciphertext, err := EncryptBig(plaintext, publicKey)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := DecryptBig(ciphertext, privateKey)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDecrypt_InvalidData(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	_, err := Decrypt([]byte("invalid encrypted data"), privateKey)
	assert.Error(t, err)
}

func TestDecryptBig_InvalidData(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	_, err := DecryptBig([]byte("invalid encrypted data"), privateKey)
	assert.Error(t, err)
}

func TestEncrypt_WithNilKey(t *testing.T) {
	_, err := Encrypt([]byte("data"), nil)
	assert.Error(t, err)
}

func TestEncryptBig_WithNilKey(t *testing.T) {
	_, err := EncryptBig([]byte("data"), nil)
	assert.Error(t, err)
}

func TestDecrypt_WithNilKey(t *testing.T) {
	_, err := Decrypt([]byte("data"), nil)
	assert.Error(t, err)
}

func TestDecryptBig_WithNilKey(t *testing.T) {
	_, err := DecryptBig([]byte("data"), nil)
	assert.Error(t, err)
}
