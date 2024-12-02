package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/analiticmetrics"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/storages"

	"github.com/go-chi/chi/v5"
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

	r := chi.NewRouter()
	r.Get("/", allInfoHandle)
	r.Get("/value/{type}/{name}", getHandle)
	r.Post("/update/{type}/{name}/{value}", updateHandle)

	log.Printf("Start server on endpoint %v.", endpoint)
	err := http.ListenAndServe(endpoint, r)
	if err != nil {
		panic(err)
	}
}

func allInfoHandle(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	res, err := json.Marshal(stor.GetAll())
	if err != nil {
		http.Error(w, "Unexpected error.", http.StatusInternalServerError)
		return
	}

	w.Write(res)
}

func getHandle(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	t := (analiticmetrics.MetricType)(r.PathValue("type"))
	n := r.PathValue("name")

	switch t {
	case analiticmetrics.Gauge:
		if !stor.Contains(t, n) {
			http.Error(w, "Incorrect metric name", http.StatusNotFound)
			return
		}
		w.Write([]byte(*stor.GetValue(t, n)))
	case analiticmetrics.Counter:
		if !stor.Contains(t, n) {
			http.Error(w, "Incorrect metric name", http.StatusNotFound)
			return
		}
		w.Write([]byte(*stor.GetValue(t, n)))
	default:
		http.Error(w, "Incorrect metric type", http.StatusBadRequest)
		return
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
