package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	crypto "github.com/Mr-Filatik/go-metrics-collector/internal/crypto/rsa"
)

// WithDecryption создает middleware для расшифровки тела запроса с помощью приватного ключа.
func (c *Conveyor) WithDecryption(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey == nil {
				next.ServeHTTP(w, r)
				return
			}

			encryptedBody, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			defer func() {
				if err := r.Body.Close(); err != nil {
					c.log.Error("body close error", err)
				}
			}()

			decryptedBody, err := crypto.DecryptBig(encryptedBody, privateKey)
			if err != nil {
				c.log.Error("Decryption failed", err)
				http.Error(w, "Decryption failed", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decryptedBody))

			next.ServeHTTP(w, r)
		})
	}
}
