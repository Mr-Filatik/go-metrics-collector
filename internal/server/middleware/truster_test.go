package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type subnetTestInfo struct {
	header  string
	trusted string
}

func TestWithTrustSubnet(t *testing.T) {
	mockLog := &mockLogger{}
	conveyor := New(mockLog)

	tests := []struct {
		name               string
		inputSubnetInfo    subnetTestInfo
		expectedStatusCode int
	}{
		{
			name: "trusted subnet",
			inputSubnetInfo: subnetTestInfo{
				header:  "sub.net",
				trusted: "sub.net",
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "no trusted subnet",
			inputSubnetInfo: subnetTestInfo{
				header:  "no.sub.net",
				trusted: "sub.net",
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "off trusted subnet",
			inputSubnetInfo: subnetTestInfo{
				header:  "sub.net",
				trusted: "",
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
			req.Header.Set("X-Real-IP", tt.inputSubnetInfo.header)

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ip := r.Header.Get("X-Real-IP")
				assert.NotEmpty(t, ip)
				w.WriteHeader(http.StatusOK)
			})

			handler := conveyor.WithTrustSubnet(next, tt.inputSubnetInfo.trusted)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatusCode, rec.Code)
		})
	}
}
