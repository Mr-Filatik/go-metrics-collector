package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/entity"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/storage/abstract"
)

func GetAllMetricsHandle(s abstract.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("Invalid request type %v, needed GET.", r.Method)
			http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		res, err := json.Marshal(s.GetAll())
		if err != nil {
			log.Printf("Unexpected server error: %v.", err.Error())
			http.Error(w, "Unexpected error.", http.StatusInternalServerError)
			return
		}

		w.Write(res)
	}
}

func GetMetricHandle(s abstract.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("Invalid request type %v, needed GET.", r.Method)
			http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		t := (entity.MetricType)(r.PathValue("type"))
		n := r.PathValue("name")

		val, err := s.Get(t, n)
		if err != nil {
			if err.Error() == "metric not found" {
				log.Printf("Server error: %v.", err.Error())
				http.Error(w, "Error: "+err.Error(), http.StatusNotFound)
				return
			}
			if err.Error() == "invalid metric type" {
				log.Printf("Server error: %v.", err.Error())
				http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
				return
			}
		}
		w.Write([]byte(val))
	}
}

func UpdateMetricHandle(s abstract.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("Invalid request type %v, needed POST.", r.Method)
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		t := (entity.MetricType)(r.PathValue("type"))
		n := r.PathValue("name")
		v := r.PathValue("value")

		err := s.CreateOrUpdate(t, n, v)
		if err != nil {
			if err.Error() == "invalid metric value" {
				log.Printf("Server error: %v.", err.Error())
				http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err.Error() == "invalid metric type" {
				log.Printf("Server error: %v.", err.Error())
				http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
				return
			}
		}
	}
}
