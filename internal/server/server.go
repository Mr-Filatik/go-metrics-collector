package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	config "github.com/Mr-Filatik/go-metrics-collector/internal/server/config"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server/middleware"
	"github.com/Mr-Filatik/go-metrics-collector/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	router          *chi.Mux
	storage         *storage.Storage
	conveyor        *middleware.Conveyor
	log             logger.Logger
	syncSaveEnabled bool
}

func NewServer(s *storage.Storage, l logger.Logger) *Server {
	srv := Server{
		router:   chi.NewRouter(),
		storage:  s,
		conveyor: middleware.New(l),
		log:      l,
	}
	srv.routes()
	return &srv
}

func (s *Server) routes() {
	s.router.Handle("/", s.conveyor.MainConveyor(http.HandlerFunc(s.GetAllMetrics)))
	s.router.Handle("/value/", s.conveyor.MainConveyor(http.HandlerFunc(s.GetMetricJSON)))
	s.router.Handle("/update/", s.conveyor.MainConveyor(http.HandlerFunc(s.UpdateMetricJSON)))
	s.router.Handle("/value/{type}/{name}", s.conveyor.MainConveyor(http.HandlerFunc(s.GetMetric)))
	s.router.Handle("/update/{type}/{name}/{value}", s.conveyor.MainConveyor(http.HandlerFunc(s.UpdateMetric)))
}

func (s *Server) Start(conf config.Config) {
	s.syncSaveEnabled = conf.StoreInterval == 0

	if conf.Restore {
		err := s.storage.LoadData()
		if err != nil {
			s.log.Error("The data from the file was not loaded", err)
		}
	}

	if !s.syncSaveEnabled {
		go s.saveDataWithInterval(conf.StoreInterval)
	}

	s.log.Info(
		"Start server",
		"endpoint", conf.ServerAddress,
		"file path", conf.FileStoragePath,
		"interval", conf.StoreInterval,
		"restore data", conf.Restore,
	)
	err := http.ListenAndServe(conf.ServerAddress, s.router)
	if err != nil {
		log.Fatal(err)
	}

	serr := s.storage.SaveData()
	if serr != nil {
		s.log.Error("No data was saved after the server was stopped", err)
	}
}

func (s *Server) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodGet)
	if !ok {
		return
	}

	mArr, err := s.storage.GetAll()
	if err != nil {
		s.serverResponceError(w, err, http.StatusInternalServerError)
	}
	s.serverResponceWithJSON(w, mArr)
}

func (s *Server) GetMetric(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodGet)
	if !ok {
		return
	}

	m, err := s.storage.Get(r.PathValue("name"), r.PathValue("type"))
	if err != nil {
		if err.Error() == repository.ErrorMetricNotFound {
			s.serverResponceError(w, err, http.StatusNotFound)
			return
		}
		s.reportServerError(w, err, storage.IsExpectedError(err))
		return
	}

	if m.MType == entity.Gauge {
		val := strconv.FormatFloat(*m.Value, 'f', -1, 64)
		_, wErr := w.Write([]byte(val))
		if wErr != nil {
			s.reportServerError(w, wErr, false)
		}
	}
	if m.MType == entity.Counter {
		val := strconv.FormatInt(*m.Delta, 10)
		_, wErr := w.Write([]byte(val))
		if wErr != nil {
			s.reportServerError(w, wErr, false)
		}
	}
}

func (s *Server) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodPost)
	if !ok {
		return
	}

	metr, err := getMetricFromJSON(r, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err := s.storage.Get(metr.ID, metr.MType)
	if err != nil {
		if storage.IsExpectedError(err) {
			if err.Error() == repository.ErrorMetricNotFound {
				s.serverResponceError(w, err, http.StatusNotFound)
				return
			}
			s.reportServerError(w, err, true)
		} else {
			s.reportServerError(w, err, false)
		}
		return
	}

	s.serverResponceWithJSON(w, m)
}

func (s *Server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodPost)
	if !ok {
		return
	}

	val := r.PathValue("value")
	metr := entity.Metrics{
		ID:    r.PathValue("name"),
		MType: r.PathValue("type"),
	}
	if metr.MType == entity.Gauge {
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			s.reportServerError(w, err, true)
			return
		}
		metr.Value = &num
	}
	if metr.MType == entity.Counter {
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			s.reportServerError(w, err, true)
			return
		}
		metr.Delta = &num
	}

	_, err := s.storage.CreateOrUpdate(metr)
	if err != nil {
		s.reportServerError(w, err, storage.IsExpectedError(err))
	}
	if s.syncSaveEnabled {
		serr := s.storage.SaveData()
		if serr != nil {
			http.Error(w, "Unexpected error with file.", http.StatusInternalServerError)
			return
		}
		s.log.Info(
			"Save in file is DONE",
			"time", time.Now(),
		)
	}
}

func (s *Server) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodPost)
	if !ok {
		return
	}

	metr, err := getMetricFromJSON(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err := s.storage.CreateOrUpdate(metr)
	if err != nil {
		s.reportServerError(w, err, storage.IsExpectedError(err))
		return
	}
	if s.syncSaveEnabled {
		serr := s.storage.SaveData()
		if serr != nil {
			http.Error(w, "Unexpected error with file.", http.StatusInternalServerError)
			return
		}
		s.log.Info(
			"Save in file is DONE",
			"time", time.Now(),
		)
	}

	s.serverResponceWithJSON(w, m)
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

	if validateValue && metr.Delta == nil && metr.Value == nil {
		return entity.Metrics{}, errors.New("invalid value or delta")
	}

	return metr, nil
}

func (s *Server) validateRequestMethod(w http.ResponseWriter, current string, needed string) bool {
	if current != needed {
		s.log.Info(
			"Invalid request method",
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
		s.serverResponceError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	if err != nil {
		s.serverResponceError(w, err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) serverResponceError(w http.ResponseWriter, err error, status int) {
	s.log.Info(
		"Server error",
		"error", err.Error(),
	)
	http.Error(w, "Error: "+err.Error(), status)
}

func (s *Server) reportServerError(w http.ResponseWriter, e error, isExpected bool) {
	if isExpected {
		s.serverResponceError(w, e, http.StatusBadRequest)
	} else {
		s.log.Info(
			"Unexpected server error",
			"error", e.Error(),
		)
		http.Error(w, "Unexpected error.", http.StatusInternalServerError)
	}
}

func (s *Server) saveDataWithInterval(interval int64) {
	t := time.Tick(time.Duration(interval) * time.Second)

	for range t {
		serr := s.storage.SaveData()
		if serr != nil {
			s.log.Info(
				"Save in file ERROR",
			)
		}
		s.log.Info(
			"Save in file is DONE",
			"time", time.Now(),
		)
	}
}
