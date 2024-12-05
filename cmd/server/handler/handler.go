package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/entity"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/storage/abstract"
)

func GetAllMetricsHandle(s abstract.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		res, err := json.Marshal(s.GetAll())
		if err != nil {
			http.Error(w, "Unexpected error.", http.StatusInternalServerError)
			return
		}

		w.Write(res)
	}
}

func GetMetricHandle(s abstract.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		t := (entity.MetricType)(r.PathValue("type"))
		n := r.PathValue("name")

		val, err := s.Get(t, n)
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
}

func UpdateMetricHandle(s abstract.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		t := (entity.MetricType)(r.PathValue("type"))
		n := r.PathValue("name")
		v := r.PathValue("value")

		err := s.CreateOrUpdate(t, n, v)
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
}
