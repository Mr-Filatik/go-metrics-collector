package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/repository"
	"github.com/Mr-Filatik/go-metrics-collector/cmd/server/storage"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
)

func GetAllMetricsHandle(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("Invalid request type %v, needed GET.", r.Method)
			http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		res, err := json.Marshal(s.GetAll())
		if err != nil {
			reportServerError(w, err, false)
			return
		}

		_, wErr := w.Write(res)
		if wErr != nil {
			reportServerError(w, wErr, false)
		}
	}
}

func GetMetricHandle(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("Invalid request type %v, needed GET.", r.Method)
			http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		t := entity.MetricType(r.PathValue("type"))
		n := r.PathValue("name")

		val, err := s.Get(t, n)
		if err != nil {
			if storage.IsExpectedError(err) {
				if err.Error() == repository.ErrorMetricNotFound {
					log.Printf("Server error: %v.", err.Error())
					http.Error(w, "Error: "+err.Error(), http.StatusNotFound)
					return
				}
				reportServerError(w, err, true)
			} else {
				reportServerError(w, err, false)
			}
		}

		_, wErr := w.Write([]byte(val))
		if wErr != nil {
			reportServerError(w, wErr, false)
		}
	}
}

func UpdateMetricHandle(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("Invalid request type %v, needed POST.", r.Method)
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		t := entity.MetricType(r.PathValue("type"))
		n := r.PathValue("name")
		v := r.PathValue("value")

		err := s.CreateOrUpdate(t, n, v)
		if err != nil {
			if storage.IsExpectedError(err) {
				reportServerError(w, err, true)
			} else {
				reportServerError(w, err, false)
			}
		}
	}
}

func reportServerError(w http.ResponseWriter, e error, isExpected bool) {
	if isExpected {
		log.Printf("Server error: %v.", e.Error())
		http.Error(w, "Error: "+e.Error(), http.StatusBadRequest)
	} else {
		log.Printf("Unexpected server error: %v.", e.Error())
		http.Error(w, "Unexpected error.", http.StatusInternalServerError)
	}
}
