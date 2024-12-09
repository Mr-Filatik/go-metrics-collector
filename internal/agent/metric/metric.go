package metric

import (
	"log"
	"math/rand"
	"runtime"
	"strconv"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
)

type AgentMetrics struct {
	Alloc         float64
	BuckHashSys   float64
	Frees         float64
	GCCPUFraction float64
	GCSys         float64
	HeapAlloc     float64
	HeapIdle      float64
	HeapInuse     float64
	HeapObjects   float64
	HeapReleased  float64
	HeapSys       float64
	LastGC        float64
	Lookups       float64
	MCacheInuse   float64
	MCacheSys     float64
	MSpanInuse    float64
	MSpanSys      float64
	Mallocs       float64
	NextGC        float64
	NumForcedGC   float64
	NumGC         float64
	OtherSys      float64
	PauseTotalNs  float64
	StackInuse    float64
	StackSys      float64
	Sys           float64
	TotalAlloc    float64
	PollCount     int64
	RandomValue   float64
}

func New() *AgentMetrics {
	metrics := AgentMetrics{PollCount: 0}
	return &metrics
}

func (metric *AgentMetrics) Update() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	metric.Alloc = float64(mem.Alloc)
	metric.BuckHashSys = float64(mem.BuckHashSys)
	metric.Frees = float64(mem.Frees)
	metric.GCCPUFraction = mem.GCCPUFraction
	metric.GCSys = float64(mem.GCSys)
	metric.HeapAlloc = float64(mem.HeapAlloc)
	metric.HeapIdle = float64(mem.HeapIdle)
	metric.HeapInuse = float64(mem.HeapInuse)
	metric.HeapObjects = float64(mem.HeapObjects)
	metric.HeapReleased = float64(mem.HeapReleased)
	metric.HeapSys = float64(mem.HeapSys)
	metric.LastGC = float64(mem.LastGC)
	metric.Lookups = float64(mem.Lookups)
	metric.MCacheInuse = float64(mem.MCacheInuse)
	metric.MCacheSys = float64(mem.MCacheSys)
	metric.MSpanInuse = float64(mem.MSpanInuse)
	metric.MSpanSys = float64(mem.MSpanSys)
	metric.Mallocs = float64(mem.Mallocs)
	metric.NextGC = float64(mem.NextGC)
	metric.NumForcedGC = float64(mem.NumForcedGC)
	metric.NumGC = float64(mem.NumGC)
	metric.OtherSys = float64(mem.OtherSys)
	metric.PauseTotalNs = float64(mem.PauseTotalNs)
	metric.StackInuse = float64(mem.StackInuse)
	metric.StackSys = float64(mem.MSpanSys)
	metric.Sys = float64(mem.Sys)
	metric.TotalAlloc = float64(mem.TotalAlloc)
	metric.PollCount++
	metric.RandomValue = rand.Float64()

	log.Printf("Update metrics.")
}

func (metric *AgentMetrics) GetAll(isClearCounters bool) []entity.Metric {
	list := make([]entity.Metric, 0)
	list = append(list,
		addMetric(entity.Gauge, "Alloc", strconv.FormatFloat(metric.Alloc, 'f', -1, 64)),
		addMetric(entity.Gauge, "BuckHashSys", strconv.FormatFloat(metric.BuckHashSys, 'f', -1, 64)),
		addMetric(entity.Gauge, "Frees", strconv.FormatFloat(metric.Frees, 'f', -1, 64)),
		addMetric(entity.Gauge, "Frees", strconv.FormatFloat(metric.Frees, 'f', -1, 64)),
		addMetric(entity.Gauge, "GCCPUFraction", strconv.FormatFloat(metric.GCCPUFraction, 'f', -1, 64)),
		addMetric(entity.Gauge, "GCSys", strconv.FormatFloat(metric.GCSys, 'f', -1, 64)),
		addMetric(entity.Gauge, "HeapAlloc", strconv.FormatFloat(metric.HeapAlloc, 'f', -1, 64)),
		addMetric(entity.Gauge, "HeapIdle", strconv.FormatFloat(metric.HeapIdle, 'f', -1, 64)),
		addMetric(entity.Gauge, "HeapInuse", strconv.FormatFloat(metric.HeapInuse, 'f', -1, 64)),
		addMetric(entity.Gauge, "HeapObjects", strconv.FormatFloat(metric.HeapObjects, 'f', -1, 64)),
		addMetric(entity.Gauge, "HeapReleased", strconv.FormatFloat(metric.HeapReleased, 'f', -1, 64)),
		addMetric(entity.Gauge, "HeapSys", strconv.FormatFloat(metric.HeapSys, 'f', -1, 64)),
		addMetric(entity.Gauge, "LastGC", strconv.FormatFloat(metric.LastGC, 'f', -1, 64)),
		addMetric(entity.Gauge, "Lookups", strconv.FormatFloat(metric.Lookups, 'f', -1, 64)),
		addMetric(entity.Gauge, "MCacheInuse", strconv.FormatFloat(metric.MCacheInuse, 'f', -1, 64)),
		addMetric(entity.Gauge, "MCacheSys", strconv.FormatFloat(metric.MCacheSys, 'f', -1, 64)),
		addMetric(entity.Gauge, "MSpanInuse", strconv.FormatFloat(metric.MSpanInuse, 'f', -1, 64)),
		addMetric(entity.Gauge, "MSpanSys", strconv.FormatFloat(metric.MSpanSys, 'f', -1, 64)),
		addMetric(entity.Gauge, "Mallocs", strconv.FormatFloat(metric.Mallocs, 'f', -1, 64)),
		addMetric(entity.Gauge, "NextGC", strconv.FormatFloat(metric.NextGC, 'f', -1, 64)),
		addMetric(entity.Gauge, "NumForcedGC", strconv.FormatFloat(metric.NumForcedGC, 'f', -1, 64)),
		addMetric(entity.Gauge, "NumGC", strconv.FormatFloat(metric.NumGC, 'f', -1, 64)),
		addMetric(entity.Gauge, "OtherSys", strconv.FormatFloat(metric.OtherSys, 'f', -1, 64)),
		addMetric(entity.Gauge, "PauseTotalNs", strconv.FormatFloat(metric.PauseTotalNs, 'f', -1, 64)),
		addMetric(entity.Gauge, "StackInuse", strconv.FormatFloat(metric.StackInuse, 'f', -1, 64)),
		addMetric(entity.Gauge, "StackSys", strconv.FormatFloat(metric.StackSys, 'f', -1, 64)),
		addMetric(entity.Gauge, "Sys", strconv.FormatFloat(metric.Sys, 'f', -1, 64)),
		addMetric(entity.Gauge, "TotalAlloc", strconv.FormatFloat(metric.TotalAlloc, 'f', -1, 64)),
		addMetric(entity.Gauge, "RandomValue", strconv.FormatFloat(metric.RandomValue, 'f', -1, 64)),
		addMetric(entity.Counter, "PollCount", strconv.FormatInt(metric.PollCount, 10)))
	if isClearCounters {
		metric.PollCount = 0
	}
	return list
}

func addMetric(t entity.MetricType, n string, v string) entity.Metric {
	return entity.Metric{Type: t, Name: n, Value: v}
}

func (metric *AgentMetrics) Log() {
	log.Printf("Log all metrics:")
	log.Printf("- Alloc: %v.", metric.Alloc)
	log.Printf("- BuckHashSys: %v.", metric.BuckHashSys)
	log.Printf("- Frees: %v.", metric.Frees)
	log.Printf("- GCCPUFraction: %v.", metric.GCCPUFraction)
	log.Printf("- GCSys: %v.", metric.GCSys)
	log.Printf("- HeapAlloc: %v.", metric.HeapAlloc)
	log.Printf("- HeapIdle: %v.", metric.HeapIdle)
	log.Printf("- HeapInuse: %v.", metric.HeapInuse)
	log.Printf("- HeapObjects: %v.", metric.HeapObjects)
	log.Printf("- HeapReleased: %v.", metric.HeapReleased)
	log.Printf("- HeapSys: %v.", metric.HeapSys)
	log.Printf("- LastGC: %v.", metric.LastGC)
	log.Printf("- Lookups: %v.", metric.Lookups)
	log.Printf("- MCacheInuse: %v.", metric.MCacheInuse)
	log.Printf("- MCacheSys: %v.", metric.MCacheSys)
	log.Printf("- MSpanInuse: %v.", metric.MSpanInuse)
	log.Printf("- MSpanSys: %v.", metric.MSpanSys)
	log.Printf("- Mallocs: %v.", metric.Mallocs)
	log.Printf("- NextGC: %v.", metric.NextGC)
	log.Printf("- NumForcedGC: %v.", metric.NumForcedGC)
	log.Printf("- NumGC: %v.", metric.NumGC)
	log.Printf("- OtherSys: %v.", metric.OtherSys)
	log.Printf("- PauseTotalNs: %v.", metric.PauseTotalNs)
	log.Printf("- StackInuse: %v.", metric.StackInuse)
	log.Printf("- StackSys: %v.", metric.StackSys)
	log.Printf("- Sys: %v.", metric.Sys)
	log.Printf("- TotalAlloc: %v.", metric.TotalAlloc)
	log.Printf("- PollCount: %v.", metric.PollCount)
	log.Printf("- RandomValue: %v.", metric.RandomValue)
}
