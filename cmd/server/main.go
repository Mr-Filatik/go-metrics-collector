package main

import (
	"log"
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/analiticmetrics"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/storages"
)

var stor storages.Storage = &storages.MemStorage{}

func main() {
	endpoint := "127.0.0.1:8080"
	stor.Create(analiticmetrics.Gauge, "testGauge", "0")
	stor.Create(analiticmetrics.Counter, "testCounter", "0")

	stor.Create(analiticmetrics.Gauge, "Alloc", "0")
	stor.Create(analiticmetrics.Gauge, "BuckHashSys", "0")
	stor.Create(analiticmetrics.Gauge, "Frees", "0")
	stor.Create(analiticmetrics.Gauge, "GCCPUFraction", "0")
	stor.Create(analiticmetrics.Gauge, "GCSys", "0")
	stor.Create(analiticmetrics.Gauge, "HeapAlloc", "0")
	stor.Create(analiticmetrics.Gauge, "HeapIdle", "0")
	stor.Create(analiticmetrics.Gauge, "HeapInuse", "0")
	stor.Create(analiticmetrics.Gauge, "HeapObjects", "0")
	stor.Create(analiticmetrics.Gauge, "HeapReleased", "0")
	stor.Create(analiticmetrics.Gauge, "HeapSys", "0")
	stor.Create(analiticmetrics.Gauge, "LastGC", "0")
	stor.Create(analiticmetrics.Gauge, "Lookups", "0")
	stor.Create(analiticmetrics.Gauge, "MCacheInuse", "0")
	stor.Create(analiticmetrics.Gauge, "MCacheSys", "0")
	stor.Create(analiticmetrics.Gauge, "MSpanInuse", "0")
	stor.Create(analiticmetrics.Gauge, "MSpanSys", "0")
	stor.Create(analiticmetrics.Gauge, "Mallocs", "0")
	stor.Create(analiticmetrics.Gauge, "NextGC", "0")
	stor.Create(analiticmetrics.Gauge, "NumForcedGC", "0")
	stor.Create(analiticmetrics.Gauge, "NumGC", "0")
	stor.Create(analiticmetrics.Gauge, "OtherSys", "0")
	stor.Create(analiticmetrics.Gauge, "PauseTotalNs", "0")
	stor.Create(analiticmetrics.Gauge, "StackInuse", "0")
	stor.Create(analiticmetrics.Gauge, "StackSys", "0")
	stor.Create(analiticmetrics.Gauge, "Sys", "0")
	stor.Create(analiticmetrics.Gauge, "TotalAlloc", "0")
	stor.Create(analiticmetrics.Counter, "PollCount", "0")
	stor.Create(analiticmetrics.Gauge, "RandomValue", "0")

	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{type}/{name}/{value}", updateHandle)

	log.Printf("Start server on endpoint %v.", endpoint)
	err := http.ListenAndServe(endpoint, mux)
	if err != nil {
		panic(err)
	}
}

func updateHandle(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	t := (analiticmetrics.MetricType)(r.PathValue("type"))
	n := r.PathValue("name")
	v := r.PathValue("value")

	switch t {
	case analiticmetrics.Gauge:
		if !stor.Contains(t, n) {
			http.Error(w, "Incorrect metric name", http.StatusNotFound)
			return
		}
		if err := stor.Update(t, n, v); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case analiticmetrics.Counter:
		if !stor.Contains(t, n) {
			http.Error(w, "Incorrect metric name", http.StatusNotFound)
			return
		}
		if err := stor.Update(t, n, v); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Incorrect metric type", http.StatusBadRequest)
		return
	}
}
