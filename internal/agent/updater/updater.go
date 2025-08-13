// Пакет updater предоставляет реализацию воркера для отправки метрик на сервер.
// Пакет использует клиент resty, поддерживает отправку наборами данных и их сжатие по алгоритму gzip.
package updater

import (
	"context"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
)

// Run запускает обновление основных метрик с заданным интервалом.
//
// Параметры:
//   - ctx: контекст для отмены
//   - m: объект метрик (AgentMetrics)
//   - pollInterval: интервал обновления метрик (в секундах)
func Run(ctx context.Context, m *metric.AgentMetrics, pollInterval int64) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.Update()
		}
	}
}

// RunMemory запускает обновление метрик памяти приложения с заданным интервалом.
//
// Параметры:
//   - ctx: контекст для отмены
//   - m: объект метрик (AgentMetrics)
//   - pollInterval: интервал обновления метрик (в секундах)
func RunMemory(ctx context.Context, m *metric.AgentMetrics, pollInterval int64) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.UpdateMemory()
		}
	}
}
