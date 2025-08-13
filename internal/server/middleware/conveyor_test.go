package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/stretchr/testify/assert"
)

type mockLogger struct {
	debugCalled []string
	infoCalled  []string
	errorCalled []string
	warnCalled  []string
}

func (m *mockLogger) Log(log logger.LogLevel, msg string, keyvals ...interface{}) {
	m.debugCalled = append(m.debugCalled, msg)
}

func (m *mockLogger) Debug(msg string, keyvals ...interface{}) {
	m.debugCalled = append(m.debugCalled, msg)
}

func (m *mockLogger) Info(msg string, keyvals ...interface{}) {
	var parts []string
	parts = append(parts, msg)
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}
		var value string
		if i+1 < len(keyvals) {
			value = fmt.Sprint(keyvals[i+1])
		} else {
			value = "<no-value>"
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	m.infoCalled = append(m.infoCalled, strings.Join(parts, " "))
}

func (m *mockLogger) Error(msg string, err error, keyvals ...interface{}) {
	m.errorCalled = append(m.errorCalled, msg)
}

func (m *mockLogger) Warn(msg string, keyvals ...interface{}) {
	m.warnCalled = append(m.warnCalled, msg)
}

func (m *mockLogger) Close() {}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(r.URL.Path))
}

func TestRegisterMiddlewares(t *testing.T) {
	mockLog := &mockLogger{}
	conveyor := New(mockLog)

	m1 := func(next http.Handler) http.Handler {
		return next
	}
	m2 := func(next http.Handler) http.Handler {
		return next
	}

	conveyor.RegisterMiddlewares(m1, m2)

	assert.Len(t, conveyor.middlewares, 2)
}

func TestMiddlewares_Replaces(t *testing.T) {
	mockLog := &mockLogger{}
	conveyor := New(mockLog)

	var count1 int
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count1++
			next.ServeHTTP(w, r)
		})
	}

	var count2 int
	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count2++
			next.ServeHTTP(w, r)
		})
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)

	conveyor.RegisterMiddlewares(m1)
	handler1 := conveyor.Middlewares(http.HandlerFunc(echoHandler))

	handler1.ServeHTTP(rec, req)
	assert.Equal(t, 1, count1)
	assert.Equal(t, 0, count2)

	conveyor.RegisterMiddlewares(m2)
	handler2 := conveyor.Middlewares(http.HandlerFunc(echoHandler))

	handler2.ServeHTTP(rec, req)
	handler2.ServeHTTP(rec, req)
	assert.Equal(t, 1, count1)
	assert.Equal(t, 2, count2)

	conveyor.RegisterMiddlewares()
	handler3 := conveyor.Middlewares(http.HandlerFunc(echoHandler))

	handler3.ServeHTTP(rec, req)
	assert.Equal(t, 1, count1)
	assert.Equal(t, 2, count2)
}

func TestMiddlewares_Order(t *testing.T) {
	mockLog := &mockLogger{}
	conveyor := New(mockLog)

	var order []string

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1-start")
			next.ServeHTTP(w, r)
			order = append(order, "m1-end")
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2-start")
			next.ServeHTTP(w, r)
			order = append(order, "m2-end")
		})
	}

	conveyor.RegisterMiddlewares(m1, m2)

	handler := conveyor.Middlewares(http.HandlerFunc(echoHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)

	handler.ServeHTTP(rec, req)

	expected := []string{"m1-start", "m2-start", "m2-end", "m1-end"}
	assert.Equal(t, expected, order)
}
