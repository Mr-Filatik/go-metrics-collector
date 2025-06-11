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

var (
	counterNames []string
	gaugeNames   []string
)

type AgentMetrics struct {
	Metrics   map[string]*Metric
	PollCount int64
}

func New() *AgentMetrics {
	metrics := AgentMetrics{
		PollCount: 0,
		Metrics:   map[string]*Metric{},
	}

	gaugeNames = []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"RandomValue",
		"TotalMemory",
		"FreeMemory",
		"CPUutilization1",
	}

	for i := range gaugeNames {
		metrics.Metrics[gaugeNames[i]] = &Metric{Type: entity.Gauge, Name: gaugeNames[i], Value: "0"}
	}

	counterNames = []string{
		"PollCount",
	}

	for i := range counterNames {
		metrics.Metrics[counterNames[i]] = &Metric{Type: entity.Counter, Name: counterNames[i], Value: "0"}
	}

	return &metrics
}

func (metric *AgentMetrics) Update() {
	var mems runtime.MemStats
	runtime.ReadMemStats(&mems)

	m := metric.Metrics["Alloc"]
	m.Value = strconv.FormatUint(mems.Alloc, 10)

	m = metric.Metrics["BuckHashSys"]
	m.Value = strconv.FormatUint(mems.BuckHashSys, 10)

	m = metric.Metrics["Frees"]
	m.Value = strconv.FormatUint(mems.Frees, 10)

	m = metric.Metrics["GCCPUFraction"]
	m.Value = strconv.FormatFloat(mems.GCCPUFraction, 'f', -1, 64)

	m = metric.Metrics["GCSys"]
	m.Value = strconv.FormatUint(mems.GCSys, 10)

	m = metric.Metrics["HeapAlloc"]
	m.Value = strconv.FormatUint(mems.HeapAlloc, 10)

	m = metric.Metrics["HeapIdle"]
	m.Value = strconv.FormatUint(mems.HeapIdle, 10)

	m = metric.Metrics["HeapInuse"]
	m.Value = strconv.FormatUint(mems.HeapInuse, 10)

	m = metric.Metrics["HeapObjects"]
	m.Value = strconv.FormatUint(mems.HeapObjects, 10)

	m = metric.Metrics["HeapReleased"]
	m.Value = strconv.FormatUint(mems.HeapReleased, 10)

	m = metric.Metrics["HeapSys"]
	m.Value = strconv.FormatUint(mems.HeapSys, 10)

	m = metric.Metrics["LastGC"]
	m.Value = strconv.FormatUint(mems.LastGC, 10)

	m = metric.Metrics["Lookups"]
	m.Value = strconv.FormatUint(mems.Lookups, 10)

	m = metric.Metrics["MCacheInuse"]
	m.Value = strconv.FormatUint(mems.MCacheInuse, 10)

	m = metric.Metrics["MCacheSys"]
	m.Value = strconv.FormatUint(mems.MCacheSys, 10)

	m = metric.Metrics["MSpanInuse"]
	m.Value = strconv.FormatUint(mems.MSpanInuse, 10)

	m = metric.Metrics["MSpanSys"]
	m.Value = strconv.FormatUint(mems.MSpanSys, 10)

	m = metric.Metrics["Mallocs"]
	m.Value = strconv.FormatUint(mems.Mallocs, 10)

	m = metric.Metrics["NextGC"]
	m.Value = strconv.FormatUint(mems.NextGC, 10)

	m = metric.Metrics["NumForcedGC"]
	m.Value = strconv.FormatUint(uint64(mems.NumForcedGC), 10)

	m = metric.Metrics["NumGC"]
	m.Value = strconv.FormatUint(uint64(mems.NumGC), 10)

	m = metric.Metrics["OtherSys"]
	m.Value = strconv.FormatUint(mems.OtherSys, 10)

	m = metric.Metrics["PauseTotalNs"]
	m.Value = strconv.FormatUint(mems.PauseTotalNs, 10)

	m = metric.Metrics["StackInuse"]
	m.Value = strconv.FormatUint(mems.StackInuse, 10)

	m = metric.Metrics["MSpanSys"]
	m.Value = strconv.FormatUint(mems.MSpanSys, 10)

	m = metric.Metrics["Sys"]
	m.Value = strconv.FormatUint(mems.Sys, 10)

	m = metric.Metrics["TotalAlloc"]
	m.Value = strconv.FormatUint(mems.TotalAlloc, 10)

	metric.PollCount++
	m = metric.Metrics["PollCount"]
	m.Value = strconv.FormatInt(metric.PollCount, 10)

	m = metric.Metrics["RandomValue"]
	m.Value = strconv.FormatFloat(rand.Float64(), 'f', -1, 64)

	log.Printf("Update metrics.")
}

func (metric *AgentMetrics) UpdateMemory() {
	vals, err := mem.VirtualMemory()
	if err != nil {
		tm := metric.Metrics["TotalMemory"]
		tm.Value = strconv.FormatUint(vals.Total, 10)
		fr := metric.Metrics["FreeMemory"]
		fr.Value = strconv.FormatUint(vals.Free, 10)
	}
	val, cerr := cpu.Percent(time.Second, true)
	if cerr != nil && len(val) > 1 {
		cu := metric.Metrics["TotalMemory"]
		cu.Value = strconv.FormatFloat(val[0], 'f', -1, 64)
	}

	log.Printf("Update memory metrics.")
}

func (metric *AgentMetrics) GetAllGaugeNames() []string {
	log.Printf("Get all gauge metrics. Count: %v.", len(gaugeNames))
	return gaugeNames
}

func (metric *AgentMetrics) GetAllCounterNames() []string {
	log.Printf("Get all counter metrics. Count: %v.", len(counterNames))
	return counterNames
}

func (metric *AgentMetrics) GetByName(name string) Metric {
	met, ok := metric.Metrics[name]
	if ok {
		log.Printf("Get metric. Name: %v. Type: %v.", name, met.Type)
		return *met
	}
	return Metric{}
}

func (metric *AgentMetrics) ClearCounter(name string) {
	if name == "PollCount" {
		log.Printf("Clear counter metric. Name: %v.", name)
		metric.PollCount = 0
		met := metric.Metrics[name]
		met.Value = "0"
	}
}

type Metric struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
