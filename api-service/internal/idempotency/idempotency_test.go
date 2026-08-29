package idempotency

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type memoryBackend struct {
	hash     string
	response *Response
	begins   int
}

func (backend *memoryBackend) Begin(_ context.Context, _ string, hash string) (Decision, error) {
	backend.begins++
	if backend.hash == "" {
		backend.hash = hash
		return Decision{Reservation: &memoryReservation{backend: backend}}, nil
	}
	if backend.hash != hash {
		return Decision{Conflict: true}, nil
	}

	replay := *backend.response
	return Decision{Replay: &replay}, nil
}

type memoryReservation struct {
	backend *memoryBackend
}

func (reservation *memoryReservation) Complete(_ context.Context, response Response) error {
	reservation.backend.response = &response
	return nil
}

func (*memoryReservation) Abort(context.Context) {}

func TestRequireStoresAndReplaysResponse(t *testing.T) {
	backend := &memoryBackend{}
	handlerCalls := 0
	handler := Require(backend, func(w http.ResponseWriter, _ *http.Request) {
		handlerCalls++
		w.Header().Set("Location", "/api/v1/products/created")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"created"}`))
	})

	first := performRequest(t, handler, "key-1", []byte(`{"name":"first"}`))
	second := performRequest(t, handler, "key-1", []byte(`{"name":"first"}`))

	if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
		t.Fatalf("statuses = %d/%d, want 201/201", first.Code, second.Code)
	}
	if first.Body.String() != second.Body.String() {
		t.Fatalf("replayed body = %q, want %q", second.Body.String(), first.Body.String())
	}
	if second.Header().Get("Location") != "/api/v1/products/created" {
		t.Fatalf("replayed Location = %q", second.Header().Get("Location"))
	}
	if handlerCalls != 1 {
		t.Fatalf("handler calls = %d, want 1", handlerCalls)
	}
}

func TestRequireRejectsReusedKeyWithDifferentRequest(t *testing.T) {
	backend := &memoryBackend{}
	handler := Require(backend, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	_ = performRequest(t, handler, "key-1", []byte(`{"name":"first"}`))
	conflict := performRequest(t, handler, "key-1", []byte(`{"name":"second"}`))

	if conflict.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", conflict.Code)
	}
	if !bytes.Contains(conflict.Body.Bytes(), []byte("IDEMPOTENCY_KEY_REUSED")) {
		t.Fatalf("body = %q", conflict.Body.String())
	}
}

func TestRequireRejectsMissingKey(t *testing.T) {
	backend := &memoryBackend{}
	handler := Require(backend, func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler was called")
	})

	response := performRequest(t, handler, "", []byte(`{}`))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
	if backend.begins != 0 {
		t.Fatalf("backend begins = %d, want 0", backend.begins)
	}
}

func performRequest(t *testing.T, handler http.HandlerFunc, key string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/products", io.NopCloser(bytes.NewReader(body)))
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	response := httptest.NewRecorder()
	handler(response, request)
	return response
}
