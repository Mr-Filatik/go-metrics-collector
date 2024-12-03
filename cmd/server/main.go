package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/config"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/entity"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/repository"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/storage"
	"github.com/go-chi/chi/v5"
)

var stor storage.MemStorage

func main() {

	config := config.Initialize()
	repo := repository.MemRepository{}
	stor = storage.MemStorage{}
	stor.SetRepository(&repo)

	r := chi.NewRouter()
	r.Get("/", allInfoHandle)
	r.Get("/value/{type}/{name}", getHandle)
	r.Post("/update/{type}/{name}/{value}", updateHandle)

	log.Printf("Start server on endpoint %v.", config.ServerAddress)
	err := http.ListenAndServe(config.ServerAddress, r)
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

	t := (entity.MetricType)(r.PathValue("type"))
	n := r.PathValue("name")

	val, err := stor.Get(t, n)
	if err != nil {
		if err.Error() == "metric not found" {
			http.Error(w, "Error: "+err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "invalid metric type" {
			http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Write([]byte(val))
}

func updateHandle(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	t := (entity.MetricType)(r.PathValue("type"))
	n := r.PathValue("name")
	v := r.PathValue("value")

	err := stor.CreateOrUpdate(t, n, v)
	if err != nil {
		if err.Error() == "invalid metric value" {
			http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error() == "invalid metric type" {
			http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
}
