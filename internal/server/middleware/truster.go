package middleware

import (
	"net/http"
	"strings"

	"github.com/Mr-Filatik/go-metrics-collector/internal/common"
)

// WithTrustSubnet создает middleware для ограничения доступа для неразрешённых подсетей.
//
// Параметры:
//   - next: следующий обработчик
//   - ts: разрешённые подсети
func (c *Conveyor) WithTrustSubnet(next http.Handler, ts string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропуск проверки, если разрешённые подсети не указаны.
		if ts == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Получение заголовка "X-Real-IP" из заголовков.
		subnetFromHeader := r.Header.Get(common.HeaderXRealIP)

		// Проверяем значение заголовка "X-Real-IP" с разрешёнными адресами.
		if subnetFromHeader != ts {
			msg := strings.Join([]string{"subnet", subnetFromHeader, "no trusted"}, " ")
			c.log.Info(msg)
			http.Error(w, msg, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
