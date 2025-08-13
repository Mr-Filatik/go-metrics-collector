package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mr-Filatik/go-metrics-collector/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(r.URL.Path))
}

func TestRegisterMiddlewares(t *testing.T) {
	mockLog := &testutil.MockLogger{}
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
	mockLog := &testutil.MockLogger{}
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
	mockLog := &testutil.MockLogger{}
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
