package server

import (
	"encoding/json"
	"log"
	"net/http"

	config "github.com/Mr-Filatik/go-metrics-collector/internal/config/server"
	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server/middleware"
	"github.com/Mr-Filatik/go-metrics-collector/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	router  *chi.Mux
	storage *storage.Storage
}

func NewServer(s *storage.Storage) *Server {
	srv := Server{
		router:  chi.NewRouter(),
		storage: s,
	}
	srv.routes()
	return &srv
}

func (s *Server) routes() {
	s.router.Handle("/", middleware.MainConveyor(http.HandlerFunc(s.GetAllMetrics)))
	s.router.Handle("/value/{type}/{name}", middleware.MainConveyor(http.HandlerFunc(s.GetMetric)))
	s.router.Handle("/update/{type}/{name}/{value}", middleware.MainConveyor(http.HandlerFunc(s.UpdateMetric)))
}

func (s *Server) Start(conf config.Config) {
	log.Printf("Start server on endpoint %v.", conf.ServerAddress)
	err := http.ListenAndServe(conf.ServerAddress, s.router)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("Invalid request type %v, needed GET.", r.Method)
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	res, err := json.Marshal(s.storage.GetAll())
	if err != nil {
		reportServerError(w, err, false)
		return
	}

	_, wErr := w.Write(res)
	if wErr != nil {
		reportServerError(w, wErr, false)
	}
}

func (s *Server) GetMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("Invalid request type %v, needed GET.", r.Method)
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	t := entity.MetricType(r.PathValue("type"))
	n := r.PathValue("name")

	val, err := s.storage.Get(t, n)
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

func (s *Server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Invalid request type %v, needed POST.", r.Method)
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	t := entity.MetricType(r.PathValue("type"))
	n := r.PathValue("name")
	v := r.PathValue("value")

	err := s.storage.CreateOrUpdate(t, n, v)
	if err != nil {
		if storage.IsExpectedError(err) {
			reportServerError(w, err, true)
		} else {
			reportServerError(w, err, false)
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
