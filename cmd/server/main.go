package main

import (
	"log"
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/analitic_metrics"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/storages"
)

var stor storages.Storage = &storages.MemStorage{}

func main() {
	endpoint := "127.0.0.1:8080"
	stor.Create(analitic_metrics.Gauge, "test_gauge", "0")
	stor.Create(analitic_metrics.Counter, "test_counter", "0")

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

	t := (analitic_metrics.MetricType)(r.PathValue("type"))
	n := r.PathValue("name")
	v := r.PathValue("value")

	switch t {
	case analitic_metrics.Gauge:
		if !stor.Contains(t, n) {
			http.Error(w, "Incorrect metric name", http.StatusNotFound)
			return
		}
		if err := stor.Update(t, n, v); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	case analitic_metrics.Counter:
		if !stor.Contains(t, n) {
			http.Error(w, "Incorrect metric name", http.StatusNotFound)
			return
		}
		if err := stor.Update(t, n, v); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	default:
		http.Error(w, "Incorrect metric type", http.StatusBadRequest)
		return
	}
}
