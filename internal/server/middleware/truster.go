package middleware

import (
	"net/http"
	"strings"
)

// WithTrustSubnet создает middleware для ограничения доступа для неразрешённых подсетей.
//
// Параметры:
//   - next: следующий обработчик
//   - ts: разрешённые подсети
func (c *Conveyor) WithTrustSubnet(next http.Handler, ts string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ts == "" {
			next.ServeHTTP(w, r)
			return
		}

		subnetFromHeader := r.Header.Get("X-Real-IP")
		if subnetFromHeader != ts {
			msg := strings.Join([]string{"subnet", subnetFromHeader, "no trusted"}, " ")
			http.Error(w, msg, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
