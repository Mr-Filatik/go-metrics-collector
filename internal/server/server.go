package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	config "github.com/Mr-Filatik/go-metrics-collector/internal/server/config"
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
	s.router.Handle("/value/", middleware.MainConveyor(http.HandlerFunc(s.GetMetricJSON)))
	s.router.Handle("/update/", middleware.MainConveyor(http.HandlerFunc(s.UpdateMetricJSON)))
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
	checkRequestMethod(w, r.Method, http.MethodGet)

	serverResponceWithJSON(w, s.storage.GetAll())
}

func (s *Server) GetMetric(w http.ResponseWriter, r *http.Request) {
	checkRequestMethod(w, r.Method, http.MethodGet)

	t := entity.MetricType(r.PathValue("type"))
	n := r.PathValue("name")

	val, err := s.storage.Get(t, n)
	if err != nil {
		if err.Error() == repository.ErrorMetricNotFound {
			serverResponceError(w, err, http.StatusNotFound)
			return
		}
		reportServerError(w, err, storage.IsExpectedError(err))
	}

	_, wErr := w.Write([]byte(val))
	if wErr != nil {
		reportServerError(w, wErr, false)
	}
}

func (s *Server) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	checkRequestMethod(w, r.Method, http.MethodPost)

	var metr entity.Metrics
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &metr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t := entity.MetricType(metr.MType)
	n := metr.ID

	val, err := s.storage.Get(t, n)
	if err != nil {
		if storage.IsExpectedError(err) {
			if err.Error() == repository.ErrorMetricNotFound {
				serverResponceError(w, err, http.StatusNotFound)
				return
			}
			reportServerError(w, err, true)
		} else {
			reportServerError(w, err, false)
		}
	}
	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		reportServerError(w, err, storage.IsExpectedError(err))
	}

	respMetr := entity.Metrics{
		ID:    metr.ID,
		MType: metr.MType,
	}
	if metr.MType == string(entity.Counter) {
		num := int64(*metr.Value)
		respMetr.Delta = &num
	}
	if metr.MType == string(entity.Gauge) {
		respMetr.Value = metr.Value
	}
	serverResponceWithJSON(w, respMetr)
}

func (s *Server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	checkRequestMethod(w, r.Method, http.MethodPost)

	t := entity.MetricType(r.PathValue("type"))
	n := r.PathValue("name")
	v := r.PathValue("value")

	err := s.storage.CreateOrUpdate(t, n, v)
	if err != nil {
		reportServerError(w, err, storage.IsExpectedError(err))
	}
}

func (s *Server) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	checkRequestMethod(w, r.Method, http.MethodPost)

	var metr entity.Metrics
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &metr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t := entity.MetricType(metr.MType)
	n := metr.ID

	if metr.Delta == nil && metr.Value == nil {
		http.Error(w, "invalid values", http.StatusBadRequest)
		return
	}
	var val float64
	if metr.Value != nil {
		val = *metr.Value
	}
	if metr.Delta != nil {
		val = float64(*metr.Delta)
	}

	err = s.storage.CreateOrUpdate(t, n, strconv.FormatFloat(val, 'f', -1, 64))
	if err != nil {
		reportServerError(w, err, storage.IsExpectedError(err))
	}
	newsval, err := s.storage.Get(t, n)
	if err == nil {
		if nv, nerr := strconv.ParseFloat(newsval, 64); nerr == nil {
			metr.Value = &nv
		}
	}

	respMetr := entity.Metrics{
		ID:    metr.ID,
		MType: metr.MType,
	}
	if metr.MType == string(entity.Counter) {
		num := int64(*metr.Value)
		respMetr.Delta = &num
	}
	if metr.MType == string(entity.Gauge) {
		respMetr.Value = metr.Value
	}
	serverResponceWithJSON(w, respMetr)
}

func checkRequestMethod(w http.ResponseWriter, current string, needed string) {
	if current != needed {
		log.Printf("Invalid request type %v, needed %v.", current, needed)
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}
}

func serverResponceWithJSON(w http.ResponseWriter, v any) {
	res, err := json.Marshal(v)
	if err != nil {
		serverResponceError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	if err != nil {
		serverResponceError(w, err, http.StatusInternalServerError)
		return
	}
}

func serverResponceError(w http.ResponseWriter, err error, status int) {
	log.Printf("Server error: %v.", err.Error())
	http.Error(w, "Error: "+err.Error(), status)
}

func reportServerError(w http.ResponseWriter, e error, isExpected bool) {
	if isExpected {
		serverResponceError(w, e, http.StatusBadRequest)
	} else {
		log.Printf("Unexpected server error: %v.", e.Error())
		http.Error(w, "Unexpected error.", http.StatusInternalServerError)
	}
}
