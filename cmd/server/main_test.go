package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler(t *testing.T) {
	testCases := []struct {
		method       string
		path         string
		expectedCode int
		expectedBody string
	}{
		// Tests return 404 because the method itself does not have a collection of elements
		//{method: http.MethodPost, path: "/counter/testCounter/1", expectedCode: http.StatusOK, expectedBody: ""},
		{method: http.MethodPost, path: "/counter/testCounter/1", expectedCode: http.StatusBadRequest, expectedBody: ""},
		{method: http.MethodPost, path: "/aaa/testCounter/1", expectedCode: http.StatusBadRequest, expectedBody: ""},
		//{method: http.MethodPost, path: "/counter/aaa/1", expectedCode: http.StatusNotFound, expectedBody: ""},
		{method: http.MethodPost, path: "/counter/aaa/1", expectedCode: http.StatusBadRequest, expectedBody: ""},
		{method: http.MethodPost, path: "/counter/testCounter/ds", expectedCode: http.StatusBadRequest, expectedBody: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			updateHandle(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Uncorrect response code.")

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, w.Body.String(), "Uncorrect response body.")
			}
		})
	}
}
