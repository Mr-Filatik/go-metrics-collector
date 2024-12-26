package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	"github.com/Mr-Filatik/go-metrics-collector/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {

}

func TestGetAllMetrics(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		path               string
		parameters         map[string]string
		mockCreateOrUpdate func(t string, n, v string) error
		expectedStatus     int
		expectedBody       string
	}{
		{
			name:           "Request empty list",
			method:         http.MethodGet,
			path:           "/",
			parameters:     map[string]string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "[]",
		},
	}

	log := logger.New()
	defer log.Close()
	repo := repository.New()
	stor := storage.New(repo, log)
	serv := &Server{
		storage: stor,
		log:     log,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			w := httptest.NewRecorder()

			serv.GetAllMetrics(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			require.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestGetMetric(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		path               string
		parameters         map[string]string
		mockCreateOrUpdate func(t string, n, v string) error
		expectedStatus     int
		expectedBody       string
	}{
		{
			name:           "Invalid metric type",
			method:         http.MethodGet,
			path:           "/value/abracadabra/testAbracadabra",
			parameters:     map[string]string{"type": "abracadabra", "name": "testAbracadabra"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Error: invalid metric type\n",
		},
		{
			name:           "Invalid metric name for gauge type",
			method:         http.MethodGet,
			path:           "/value/gauge/testGauge",
			parameters:     map[string]string{"type": "gauge", "name": "testGauge"},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Error: metric not found\n",
		},
		{
			name:           "Invalid metric name for counter type",
			method:         http.MethodGet,
			path:           "/value/counter/testCounter",
			parameters:     map[string]string{"type": "counter", "name": "testCounter"},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Error: metric not found\n",
		},
	}

	log := logger.New()
	defer log.Close()
	repo := repository.New()
	stor := storage.New(repo, log)
	serv := &Server{
		storage: stor,
		log:     log,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			req.SetPathValue("type", tt.parameters["type"])
			req.SetPathValue("name", tt.parameters["name"])
			w := httptest.NewRecorder()

			serv.GetMetric(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			require.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestUpdateMetric(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		path               string
		parameters         map[string]string
		mockCreateOrUpdate func(t string, n, v string) error
		expectedStatus     int
		expectedBody       string
	}{
		{
			name:           "Valid gauge metric type",
			method:         http.MethodPost,
			path:           "/update/gauge/testCounter/10",
			parameters:     map[string]string{"type": "gauge", "name": "testGauge", "value": "10"},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Valid counter metric type",
			method:         http.MethodPost,
			path:           "/update/counter/testCounter/10",
			parameters:     map[string]string{"type": "counter", "name": "testCounter", "value": "10"},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Invalid metric type",
			method:         http.MethodPost,
			path:           "/update/abracadabra/testAbracadabra/10",
			parameters:     map[string]string{"type": "abracadabra", "name": "testAbracadabra", "value": "10"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Error: invalid metric type\n",
		},
	}

	log := logger.New()
	defer log.Close()
	repo := repository.New()
	stor := storage.New(repo, log)
	serv := &Server{
		storage: stor,
		log:     log,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			req.SetPathValue("type", tt.parameters["type"])
			req.SetPathValue("name", tt.parameters["name"])
			req.SetPathValue("value", tt.parameters["value"])
			w := httptest.NewRecorder()

			serv.UpdateMetric(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			require.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}
