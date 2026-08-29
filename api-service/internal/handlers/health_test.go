package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type readinessStub struct {
	err error
}

func (stub readinessStub) Ping(context.Context) error {
	return stub.err
}

func TestHealthLiveMatchesContract(t *testing.T) {
	response := httptest.NewRecorder()
	healthLive(response, httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil))

	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("response = (%d, %q)", response.Code, response.Body.String())
	}
}

func TestHealthReadiness(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{name: "ready", status: http.StatusOK, body: `"status":"ok"`},
		{name: "unavailable", err: errors.New("database unavailable"), status: http.StatusServiceUnavailable, body: `"status":"unavailable"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler := healthCheck(readinessStub{err: test.err}, http.StatusServiceUnavailable)
			handler(response, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))

			if response.Code != test.status || !strings.Contains(response.Body.String(), test.body) {
				t.Fatalf("response = (%d, %q)", response.Code, response.Body.String())
			}
		})
	}
}

func TestSwaggerEndpoints(t *testing.T) {
	uiResponse := httptest.NewRecorder()
	swaggerUI(uiResponse, httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil))
	if uiResponse.Code != http.StatusOK || !strings.Contains(uiResponse.Body.String(), "/swagger/openapi.yaml") {
		t.Fatalf("Swagger UI response = (%d, %q)", uiResponse.Code, uiResponse.Body.String())
	}

	document := []byte("openapi: 3.0.3\n")
	documentResponse := httptest.NewRecorder()
	serveOpenAPI(document)(documentResponse, httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil))
	if documentResponse.Code != http.StatusOK || documentResponse.Body.String() != string(document) {
		t.Fatalf("OpenAPI response = (%d, %q)", documentResponse.Code, documentResponse.Body.String())
	}
}
