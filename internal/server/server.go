package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server/middleware"
	"github.com/Mr-Filatik/go-metrics-collector/internal/service"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	router   *chi.Mux
	service  *service.Service
	conveyor *middleware.Conveyor
	log      logger.Logger
}

func NewServer(s *service.Service, hashKey string, l logger.Logger) *Server {
	srv := Server{
		router:   chi.NewRouter(),
		service:  s,
		conveyor: middleware.New(hashKey, l),
		log:      l,
	}
	srv.routes()
	return &srv
}

func (s *Server) routes() {
	s.router.Mount("/debug", http.DefaultServeMux)

	s.router.Handle("/ping", s.conveyor.MainConveyor(http.HandlerFunc(s.Ping)))
	s.router.Handle("/", s.conveyor.MainConveyor(http.HandlerFunc(s.GetAllMetrics)))
	s.router.Handle("/updates/", s.conveyor.MainConveyor(http.HandlerFunc(s.UpdateAllMetrics)))
	s.router.Handle("/value/", s.conveyor.MainConveyor(http.HandlerFunc(s.GetMetricJSON)))
	s.router.Handle("/update/", s.conveyor.MainConveyor(http.HandlerFunc(s.UpdateMetricJSON)))
	s.router.Handle("/value/{type}/{name}", s.conveyor.MainConveyor(http.HandlerFunc(s.GetMetric)))
	s.router.Handle("/update/{type}/{name}/{value}", s.conveyor.MainConveyor(http.HandlerFunc(s.UpdateMetric)))
}

func (s *Server) Start(serverAddress string, restore bool) {
	s.service.Start(restore)

	s.log.Info(
		"Start server",
		"endpoint", serverAddress,
		"restore data", restore,
	)
	err := http.ListenAndServe(serverAddress, s.router)
	if err != nil {
		log.Fatal(err)
	}

	s.service.Stop()
}

func (s *Server) Ping(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodGet)
	if !ok {
		return
	}

	err := s.service.Ping()
	if err != nil {
		s.serverResponceInternalServerError(w, err)
	}
}

func (s *Server) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodGet)
	if !ok {
		return
	}

	mArr, err := s.service.GetAll()
	if err != nil {
		s.serverResponceInternalServerError(w, err)
		return
	}
	s.serverResponceWithJSON(w, mArr)
}

func (s *Server) UpdateAllMetrics(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodPost)
	if !ok {
		return
	}

	metr, err := getMetricsFromJSON(r, true)
	if err != nil {
		s.serverResponceBadRequest(w, err)
		return
	}

	for _, m := range metr {
		_, err := s.service.CreateOrUpdate(m)
		if err != nil {
			if err.Error() == service.MetricNotFound || err.Error() == service.MetricUncorrect {
				s.serverResponceBadRequest(w, err)
				return
			}
			s.serverResponceInternalServerError(w, err)
			return
		}
	}
}

func (s *Server) GetMetric(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodGet)
	if !ok {
		return
	}

	metr, merr := getMetricFromRequest(r, false)
	if merr != nil {
		s.serverResponceBadRequest(w, merr)
		return
	}

	m, err := s.service.Get(metr.ID, metr.MType)
	if err != nil {
		if err.Error() == service.MetricNotFound {
			s.serverResponceNotFound(w, err)
			return
		}
		if err.Error() == service.MetricUncorrect {
			s.serverResponceBadRequest(w, err)
			return
		}
		s.serverResponceInternalServerError(w, err)
		return
	}

	val := ""
	if m.MType == entity.Gauge {
		val = strconv.FormatFloat(*m.Value, 'f', -1, 64)
	}
	if m.MType == entity.Counter {
		val = strconv.FormatInt(*m.Delta, 10)
	}
	if _, wErr := w.Write([]byte(val)); wErr != nil {
		s.serverResponceInternalServerError(w, wErr)
	}
}

func (s *Server) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodPost)
	if !ok {
		return
	}

	metr, err := getMetricFromJSON(r, false)
	if err != nil {
		s.serverResponceBadRequest(w, err)
		return
	}

	m, err := s.service.Get(metr.ID, metr.MType)
	if err != nil {
		if err.Error() == service.MetricNotFound {
			s.serverResponceNotFound(w, err)
			return
		}
		if err.Error() == service.MetricUncorrect {
			s.serverResponceBadRequest(w, err)
			return
		}
		s.serverResponceInternalServerError(w, err)
		return
	}

	s.serverResponceWithJSON(w, m)
}

func (s *Server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodPost)
	if !ok {
		return
	}

	metr, merr := getMetricFromRequest(r, true)
	if merr != nil {
		s.serverResponceBadRequest(w, merr)
		return
	}

	_, err := s.service.CreateOrUpdate(metr)
	if err != nil {
		if err.Error() == service.MetricNotFound || err.Error() == service.MetricUncorrect {
			s.serverResponceBadRequest(w, err)
			return
		}
		s.serverResponceInternalServerError(w, err)
		return
	}
}

func (s *Server) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodPost)
	if !ok {
		return
	}

	metr, err := getMetricFromJSON(r, true)
	if err != nil {
		s.serverResponceBadRequest(w, err)
		return
	}

	m, err := s.service.CreateOrUpdate(metr)
	if err != nil {
		if err.Error() == service.MetricNotFound || err.Error() == service.MetricUncorrect {
			s.serverResponceBadRequest(w, err)
			return
		}
		s.serverResponceInternalServerError(w, err)
		return
	}

	s.serverResponceWithJSON(w, m)
}

func getMetricFromRequest(r *http.Request, validateValue bool) (entity.Metrics, error) {
	metr := entity.Metrics{
		ID:    r.PathValue("name"),
		MType: r.PathValue("type"),
	}
	if metr.MType != entity.Gauge && metr.MType != entity.Counter {
		return metr, errors.New("incorrect metric type")
	}
	if validateValue {
		val := r.PathValue("value")
		if metr.MType == entity.Gauge {
			num, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return metr, errors.New("incorrect metric value for float")
			}
			metr.Value = &num
		}
		if metr.MType == entity.Counter {
			num, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return metr, errors.New("invalid metric value for int")
			}
			metr.Delta = &num
		}
	}
	return metr, nil
}

func getMetricsFromJSON(r *http.Request, validateValue bool) ([]entity.Metrics, error) {
	var metr []entity.Metrics
	var buf bytes.Buffer

	if _, err := buf.ReadFrom(r.Body); err != nil {
		return make([]entity.Metrics, 0), errors.New(err.Error())
	}

	if err := json.Unmarshal(buf.Bytes(), &metr); err != nil {
		return make([]entity.Metrics, 0), errors.New(err.Error())
	}

	for _, m := range metr {
		if m.MType != entity.Gauge && m.MType != entity.Counter {
			return metr, errors.New("incorrect metric type")
		}

		if validateValue && m.Delta == nil && m.Value == nil {
			return make([]entity.Metrics, 0), errors.New("invalid metric value or delta")
		}
	}

	return metr, nil
}

func getMetricFromJSON(r *http.Request, validateValue bool) (entity.Metrics, error) {
	var metr entity.Metrics
	var buf bytes.Buffer

	if _, err := buf.ReadFrom(r.Body); err != nil {
		return entity.Metrics{}, errors.New(err.Error())
	}

	if err := json.Unmarshal(buf.Bytes(), &metr); err != nil {
		return entity.Metrics{}, errors.New(err.Error())
	}

	if metr.MType != entity.Gauge && metr.MType != entity.Counter {
		return metr, errors.New("incorrect metric type")
	}

	if validateValue && metr.Delta == nil && metr.Value == nil {
		return entity.Metrics{}, errors.New("invalid metric value or delta")
	}

	return metr, nil
}

func (s *Server) validateRequestMethod(w http.ResponseWriter, current string, needed string) bool {
	if current != needed {
		s.log.Error(
			"Invalid request",
			errors.New("invalid request method"),
			"actual", current,
			"expected", needed,
		)

		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func (s *Server) serverResponceWithJSON(w http.ResponseWriter, v any) {
	res, err := json.Marshal(v)
	if err != nil {
		s.serverResponceInternalServerError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	if err != nil {
		s.serverResponceInternalServerError(w, err)
		return
	}
}

func (s *Server) serverResponceBadRequest(w http.ResponseWriter, err error) {
	s.log.Error("Bad request error (code 400)", err)
	http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
}

func (s *Server) serverResponceNotFound(w http.ResponseWriter, err error) {
	s.log.Error("Bad request error (code 404)", err)
	http.Error(w, "Error: "+err.Error(), http.StatusNotFound)
}

func (s *Server) serverResponceInternalServerError(w http.ResponseWriter, err error) {
	s.log.Error("Internal server error (code 500)", err)
	http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
}
