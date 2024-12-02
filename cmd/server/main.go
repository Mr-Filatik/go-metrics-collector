package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/analiticmetrics"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/storages"

	"github.com/go-chi/chi/v5"
)

var stor storages.Storage = &storages.MemStorage{}

func main() {

	endpoint := flag.String("a", "localhost:8080", "HTTP server endpoint")
	flag.Parse()

	r := chi.NewRouter()
	r.Get("/", allInfoHandle)
	r.Get("/value/{type}/{name}", getHandle)
	r.Post("/update/{type}/{name}/{value}", updateHandle)

	log.Printf("Start server on endpoint %v.", *endpoint)
	err := http.ListenAndServe(*endpoint, r)
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
			stor.Create(analiticmetrics.Gauge, n, v)
			return
		}
		if err := stor.Update(t, n, v); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case analiticmetrics.Counter:
		if !stor.Contains(t, n) {
			stor.Create(analiticmetrics.Counter, n, v)
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
