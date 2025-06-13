// Пакет updater предоставляет реализацию воркера для отправки метрик на сервер.
// Пакет использует клиент resty, поддерживает отправку наборами данных и их сжатие по алгоритму gzip.
package updater

import (
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
)

// Run запускает обновление основных метрик с заданным интервалом.
//
// Параметры:
//   - m: объект метрик (AgentMetrics)
//   - pollInterval: интервал обновления метрик (в секундах)
func Run(m *metric.AgentMetrics, pollInterval int64) {
	t := time.Tick(time.Duration(pollInterval) * time.Second)

	for range t {
		m.Update()
	}
}

// RunMemory запускает обновление метрик памяти приложения с заданным интервалом.
//
// Параметры:
//   - m: объект метрик (AgentMetrics)
//   - pollInterval: интервал обновления метрик (в секундах)
func RunMemory(m *metric.AgentMetrics, pollInterval int64) {
	t := time.Tick(time.Duration(pollInterval) * time.Second)

	for range t {
		m.UpdateMemory()
	}
}
