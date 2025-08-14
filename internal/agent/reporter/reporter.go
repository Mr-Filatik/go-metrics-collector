// Пакет reporter предоставляет реализацию воркера для отправки метрик на сервер.
// Пакет использует клиент resty, поддерживает отправку наборами данных и их сжатие по алгоритму gzip.
package reporter

import (
	"context"
	"strconv"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/agent/metric"
	"github.com/Mr-Filatik/go-metrics-collector/internal/client"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
)

// Run запускает цикл отправки метрик на удалённый сервер.
// Создаёт пул воркеров и посылает сигналы на отправку каждые reportInterval секунд.
//
// Параметры:
//   - ctx: контекст для отмены
//   - m: объект метрик (AgentMetrics)
//   - endpoint: адрес сервера, куда отправляются метрики
//   - reportInterval: интервал отправки метрик (в секундах)
//   - hashKey: ключ для хэширования метрик
//   - lim: количество параллельных воркеров
//   - log: логгер
func Run(
	ctx context.Context,
	m *metric.AgentMetrics,
	reportInterval int64,
	lim int64,
	client client.Client,
	log logger.Logger) {
	jobs := make(chan struct{}, lim)
	defer close(jobs)

	for w := int64(1); w <= lim; w++ {
		go worker(ctx, m, client, log, jobs)
	}

	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case jobs <- struct{}{}:
			default:
				// log.Warn("Job queue full, skipping report", "queue_size", lim)
			}
		}
	}
}

func worker(
	ctx context.Context,
	m *metric.AgentMetrics,
	client client.Client,
	log logger.Logger,
	jobs <-chan struct{},
) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-jobs:
			var metrics []entity.Metrics
			gMetrics := m.GetAllGaugeNames()
			for _, name := range gMetrics {
				met := m.GetByName(name)
				if num, err := strconv.ParseFloat(met.Value, 64); err == nil {
					metrics = append(metrics, entity.Metrics{
						ID:    met.Name,
						MType: met.Type,
						Value: &num,
					})
				}
			}
			cMetrics := m.GetAllCounterNames()
			for _, name := range cMetrics {
				met := m.GetByName(name)
				if num, err := strconv.ParseInt(met.Value, 10, 64); err == nil {
					metrics = append(metrics, entity.Metrics{
						ID:    met.Name,
						MType: met.Type,
						Delta: &num,
					})
				}
			}

			err := client.SendMetrics(metrics)
			if err != nil {
				log.Error("Sending metrics error", err)
				continue
			}

			log.Info("Send metrics success")

			for _, name := range cMetrics {
				m.ClearCounter(name)
			}

			log.Info("Clear counter metrics success")
		}
	}
}
