package server

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	_ "net/http/pprof"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server/middleware"
	"github.com/Mr-Filatik/go-metrics-collector/internal/service"
	"github.com/go-chi/chi/v5"
)

// HTTPServer представляет HTTP-сервер приложения.
// Использует chi как маршрутизатор, service для бизнес-логики,
// conveyor для обработки данных и logger для логирования.
type HTTPServer struct {
	router      *chi.Mux             // роутер
	service     *service.Service     // сервис с основной логикой
	conveyor    *middleware.Conveyor // конвейер для middleware
	log         logger.Logger        // логгер
	http.Server                      // сервер
}

type HTTPServerConfig struct {
	PrivateRsaKey *rsa.PrivateKey
	Service       *service.Service
	Address       string
	HashKey       string
	TrustedSubnet string
}

// NewHTTPServer создаёт и инициализирует новый экзепляр *Server.
//
// Параметры:
//   - conf: конфиг сервера
func NewHTTPServer(ctx context.Context, conf *HTTPServerConfig, log logger.Logger) *HTTPServer {
	log.Info("HTTPServer creating...")

	srv := &HTTPServer{
		Server: http.Server{
			Addr: conf.Address,
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
		},
		router:   chi.NewRouter(),
		service:  conf.Service,
		conveyor: middleware.New(log),
		log:      log,
	}
	srv.registerMiddlewares(conf.HashKey, conf.PrivateRsaKey, conf.TrustedSubnet)
	srv.registerRoutes()

	log.Info("HTTPServer create is successfull")
	return srv
}

func (s *HTTPServer) Start(ctx context.Context) error {
	s.log.Info(
		"HTTPServer starting...",
		"address", s.Server.Addr,
	)

	go func() {
		if err := s.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error("Error in HTTPServer", err)
		}
	}()

	s.log.Info("HTTPServer start is successfull")
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.log.Info("HTTPServer shutdown starting...")
	err := s.Server.Shutdown(ctx)
	if err != nil {
		s.log.Error("HTTPServer shutdown error", err)
		return fmt.Errorf("HTTPServer shutdown error: %w", err)
	}
	s.log.Info("HTTPServer shutdown is successfull")
	return nil
}

func (s *HTTPServer) Close() error {
	s.log.Info("HTTPServer close starting...")
	err := s.Server.Close()
	if err != nil {
		s.log.Error("HTTPServer close error", err)
		return fmt.Errorf("HTTPServer close error: %w", err)
	}
	s.log.Info("HTTPServer close is successfull")
	return nil
}

func (s *HTTPServer) registerMiddlewares(hashKey string, privateKey *rsa.PrivateKey, ts string) {
	ms := []middleware.Middleware{
		func(h http.Handler) http.Handler {
			return s.conveyor.WithLogging(h)
		},
		func(h http.Handler) http.Handler {
			return s.conveyor.WithTrustSubnet(h, ts)
		},
		func(h http.Handler) http.Handler {
			return s.conveyor.WithDecryption(h, privateKey)
		},
		func(h http.Handler) http.Handler {
			return s.conveyor.WithCompressedGzip(h)
		},
	}

	if hashKey != "" {
		ms = append(ms, func(h http.Handler) http.Handler {
			return s.conveyor.WithHashValidation(h, hashKey)
		})
	}

	s.conveyor.RegisterMiddlewares(ms...)
}

func (s *HTTPServer) registerRoutes() {
	s.router.Mount("/debug", http.DefaultServeMux)

	s.router.Handle("/ping", s.conveyor.Middlewares(http.HandlerFunc(s.Ping)))
	s.router.Handle("/", s.conveyor.Middlewares(http.HandlerFunc(s.GetAllMetrics)))
	s.router.Handle("/updates/", s.conveyor.Middlewares(http.HandlerFunc(s.UpdateAllMetrics)))
	s.router.Handle("/value/", s.conveyor.Middlewares(http.HandlerFunc(s.GetMetricJSON)))
	s.router.Handle("/update/", s.conveyor.Middlewares(http.HandlerFunc(s.UpdateMetricJSON)))
	s.router.Handle("/value/{type}/{name}", s.conveyor.Middlewares(http.HandlerFunc(s.GetMetric)))
	s.router.Handle("/update/{type}/{name}/{value}", s.conveyor.Middlewares(http.HandlerFunc(s.UpdateMetric)))

	s.Handler = s.router
}

// Ping проверяет доступность сервера.
//
// Параметры:
//   - w: ResponseWriter
//   - r: запрос
func (s *HTTPServer) Ping(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodGet)
	if !ok {
		return
	}

	err := s.service.Ping(r.Context())
	if err != nil {
		s.serverResponceInternalServerError(w, err)
	}
}

// GetAllMetrics запрашивает получение всех метрик.
//
// Параметры:
//   - w: ResponseWriter
//   - r: запрос
func (s *HTTPServer) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodGet)
	if !ok {
		return
	}

	mArr, err := s.service.GetAll(r.Context())
	if err != nil {
		s.serverResponceInternalServerError(w, err)
		return
	}
	s.serverResponceWithJSON(w, mArr)
}

// UpdateAllMetrics обновление всех метрик.
//
// Параметры:
//   - w: ResponseWriter
//   - r: запрос
func (s *HTTPServer) UpdateAllMetrics(w http.ResponseWriter, r *http.Request) {
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
		_, err := s.service.CreateOrUpdate(r.Context(), m)
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

// GetMetric получение одной метрики.
//
// Параметры:
//   - w: ResponseWriter
//   - r: запрос
func (s *HTTPServer) GetMetric(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodGet)
	if !ok {
		return
	}

	metr, merr := getMetricFromRequest(r, false)
	if merr != nil {
		s.serverResponceBadRequest(w, merr)
		return
	}

	m, err := s.service.Get(r.Context(), metr.ID, metr.MType)
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

// GetMetricJSON получение одной метрики в формате JSON.
//
// Параметры:
//   - w: ResponseWriter
//   - r: запрос
func (s *HTTPServer) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodPost)
	if !ok {
		return
	}

	metr, err := getMetricFromJSON(r, false)
	if err != nil {
		s.serverResponceBadRequest(w, err)
		return
	}

	m, err := s.service.Get(r.Context(), metr.ID, metr.MType)
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

// UpdateMetric обновление значения одной метрики.
//
// Параметры:
//   - w: ResponseWriter
//   - r: запрос
func (s *HTTPServer) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodPost)
	if !ok {
		return
	}

	metr, merr := getMetricFromRequest(r, true)
	if merr != nil {
		s.serverResponceBadRequest(w, merr)
		return
	}

	_, err := s.service.CreateOrUpdate(r.Context(), metr)
	if err != nil {
		if err.Error() == service.MetricNotFound || err.Error() == service.MetricUncorrect {
			s.serverResponceBadRequest(w, err)
			return
		}
		s.serverResponceInternalServerError(w, err)
		return
	}
}

// UpdateMetricJSON обновление значения одной метрики в формате JSON.
//
// Параметры:
//   - w: ResponseWriter
//   - r: запрос
func (s *HTTPServer) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	ok := s.validateRequestMethod(w, r.Method, http.MethodPost)
	if !ok {
		return
	}

	metr, err := getMetricFromJSON(r, true)
	if err != nil {
		s.serverResponceBadRequest(w, err)
		return
	}

	m, err := s.service.CreateOrUpdate(r.Context(), metr)
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

func (s *HTTPServer) validateRequestMethod(w http.ResponseWriter, current string, needed string) bool {
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

func (s *HTTPServer) serverResponceWithJSON(w http.ResponseWriter, v any) {
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

func (s *HTTPServer) serverResponceBadRequest(w http.ResponseWriter, err error) {
	s.log.Error("Bad request error (code 400)", err)
	http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
}

func (s *HTTPServer) serverResponceNotFound(w http.ResponseWriter, err error) {
	s.log.Error("Bad request error (code 404)", err)
	http.Error(w, "Error: "+err.Error(), http.StatusNotFound)
}

func (s *HTTPServer) serverResponceInternalServerError(w http.ResponseWriter, err error) {
	s.log.Error("Internal server error (code 500)", err)
	http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
}
