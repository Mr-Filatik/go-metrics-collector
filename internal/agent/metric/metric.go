package metric

import (
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type AgentMetrics struct {
	Alloc           float64
	BuckHashSys     float64
	Frees           float64
	GCCPUFraction   float64
	GCSys           float64
	HeapAlloc       float64
	HeapIdle        float64
	HeapInuse       float64
	HeapObjects     float64
	HeapReleased    float64
	HeapSys         float64
	LastGC          float64
	Lookups         float64
	MCacheInuse     float64
	MCacheSys       float64
	MSpanInuse      float64
	MSpanSys        float64
	Mallocs         float64
	NextGC          float64
	NumForcedGC     float64
	NumGC           float64
	OtherSys        float64
	PauseTotalNs    float64
	StackInuse      float64
	StackSys        float64
	Sys             float64
	TotalAlloc      float64
	PollCount       int64
	RandomValue     float64
	TotalMemory     float64
	FreeMemory      float64
	CPUutilization1 float64
}

func New() *AgentMetrics {
	metrics := AgentMetrics{PollCount: 0}
	return &metrics
}

func (metric *AgentMetrics) Update() {
	var mems runtime.MemStats
	runtime.ReadMemStats(&mems)
	metric.Alloc = float64(mems.Alloc)
	metric.BuckHashSys = float64(mems.BuckHashSys)
	metric.Frees = float64(mems.Frees)
	metric.GCCPUFraction = mems.GCCPUFraction
	metric.GCSys = float64(mems.GCSys)
	metric.HeapAlloc = float64(mems.HeapAlloc)
	metric.HeapIdle = float64(mems.HeapIdle)
	metric.HeapInuse = float64(mems.HeapInuse)
	metric.HeapObjects = float64(mems.HeapObjects)
	metric.HeapReleased = float64(mems.HeapReleased)
	metric.HeapSys = float64(mems.HeapSys)
	metric.LastGC = float64(mems.LastGC)
	metric.Lookups = float64(mems.Lookups)
	metric.MCacheInuse = float64(mems.MCacheInuse)
	metric.MCacheSys = float64(mems.MCacheSys)
	metric.MSpanInuse = float64(mems.MSpanInuse)
	metric.MSpanSys = float64(mems.MSpanSys)
	metric.Mallocs = float64(mems.Mallocs)
	metric.NextGC = float64(mems.NextGC)
	metric.NumForcedGC = float64(mems.NumForcedGC)
	metric.NumGC = float64(mems.NumGC)
	metric.OtherSys = float64(mems.OtherSys)
	metric.PauseTotalNs = float64(mems.PauseTotalNs)
	metric.StackInuse = float64(mems.StackInuse)
	metric.StackSys = float64(mems.MSpanSys)
	metric.Sys = float64(mems.Sys)
	metric.TotalAlloc = float64(mems.TotalAlloc)
	metric.PollCount++
	metric.RandomValue = rand.Float64()

	log.Printf("Update metrics.")
}

func (metric *AgentMetrics) UpdateMemory() {
	vals, err := mem.VirtualMemory()
	if err != nil {
		metric.TotalMemory = float64(vals.Total)
		metric.FreeMemory = float64(vals.Free)
	}
	val, cerr := cpu.Percent(time.Second, true)
	if cerr != nil && len(val) > 1 {
		metric.CPUutilization1 = val[0]
	}

	log.Printf("Update memory metrics.")
}

func (metric *AgentMetrics) GetAllGauge() []Metric {
	list := make([]Metric, 0)
	list = append(list,
		addMetric(entity.Gauge, "Alloc", strconv.FormatFloat(metric.Alloc, 'f', -1, 64)),
		addMetric(entity.Gauge, "BuckHashSys", strconv.FormatFloat(metric.BuckHashSys, 'f', -1, 64)),
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
		addMetric(entity.Gauge, "TotalMemory", strconv.FormatFloat(metric.TotalMemory, 'f', -1, 64)),
		addMetric(entity.Gauge, "FreeMemory", strconv.FormatFloat(metric.FreeMemory, 'f', -1, 64)),
		addMetric(entity.Gauge, "CPUutilization1", strconv.FormatFloat(metric.CPUutilization1, 'f', -1, 64)))
	log.Printf("Get all gauge metrics. Count: %v.", len(list))
	return list
}

func (metric *AgentMetrics) GetAllCounter() []string {
	list := make([]string, 0)
	list = append(list, "PollCount")
	log.Printf("Get all counter metrics. Count: %v.", len(list))
	return list
}

func (metric *AgentMetrics) GetCounter(name string) Metric {
	if name == "PollCount" {
		log.Printf("Get counter metric. Name: %v.", name)
		return addMetric(entity.Counter, name, strconv.FormatInt(metric.PollCount, 10))
	}
	return Metric{}
}

func (metric *AgentMetrics) ClearCounter(name string) {
	if name == "PollCount" {
		log.Printf("Clear counter metric. Name: %v.", name)
		metric.PollCount = 0
	}
}

func addMetric(t string, n string, v string) Metric {
	return Metric{Type: t, Name: n, Value: v}
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
	log.Printf("- TotalMemory: %v.", metric.TotalMemory)
	log.Printf("- FreeMemory: %v.", metric.FreeMemory)
	log.Printf("- CPUutilization1: %v.", metric.CPUutilization1)
}

type Metric struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
