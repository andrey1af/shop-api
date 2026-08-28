//go:build e2e

package supplier_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
	"uuid"
)

type address struct {
	ID      uuid.UUID `json:"id"`
	Country string    `json:"country"`
	City    string    `json:"city"`
	Street  string    `json:"street"`
}

type supplier struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phone_number"`
	Address     address   `json:"address"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiClient struct {
	baseURL string
	client  *http.Client
}

func TestSupplierLifecycle(t *testing.T) {
	api := newAPIClient(t)

	invalidResponse := api.doJSON(t, http.MethodPost, "/api/v1/suppliers", []byte(`{"name":`))
	requireError(t, invalidResponse, http.StatusBadRequest, "INVALID_REQUEST")

	uniqueSuffix := time.Now().UTC().Format("20060102150405.000000000")
	createPayload := map[string]any{
		"name":         "E2E Supplier " + uniqueSuffix,
		"phone_number": "+7 999 000-00-01",
		"address": map[string]string{
			"country": "Russia",
			"city":    "Moscow",
			"street":  "E2E Initial Street, 1",
		},
	}

	createResponse := api.doJSON(t, http.MethodPost, "/api/v1/suppliers", createPayload)
	created := decodeResponse[supplier](t, createResponse, http.StatusCreated)
	if created.ID == uuid.Nil() || created.Address.ID == uuid.Nil() {
		t.Fatalf("create response must contain supplier and address IDs: %#v", created)
	}
	if created.Name != createPayload["name"] || created.PhoneNumber != createPayload["phone_number"] {
		t.Fatalf("create response = %#v, want name and phone from %#v", created, createPayload)
	}
	wantLocation := "/api/v1/suppliers/" + created.ID.String()
	if got := createResponse.Header.Get("Location"); got != wantLocation {
		t.Fatalf("Location = %q, want %q", got, wantLocation)
	}

	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}
		response := api.doJSON(t, http.MethodDelete, wantLocation, nil)
		response.Body.Close()
	})

	listResponse := api.doJSON(t, http.MethodGet, "/api/v1/suppliers", nil)
	suppliers := decodeResponse[[]supplier](t, listResponse, http.StatusOK)
	if !slices.ContainsFunc(suppliers, func(candidate supplier) bool { return candidate.ID == created.ID }) {
		t.Fatalf("created supplier %s is absent from supplier list", created.ID)
	}

	getResponse := api.doJSON(t, http.MethodGet, wantLocation, nil)
	got := decodeResponse[supplier](t, getResponse, http.StatusOK)
	if got != created {
		t.Fatalf("GET supplier = %#v, want %#v", got, created)
	}

	addressPath := wantLocation + "/address"
	addressResponse := api.doJSON(t, http.MethodGet, addressPath, nil)
	gotAddress := decodeResponse[address](t, addressResponse, http.StatusOK)
	if gotAddress != created.Address {
		t.Fatalf("GET address = %#v, want %#v", gotAddress, created.Address)
	}

	updatePayload := map[string]string{
		"country": "Russia",
		"city":    "Saint Petersburg",
		"street":  "E2E Updated Street, 25",
	}
	updateResponse := api.doJSON(t, http.MethodPatch, addressPath, updatePayload)
	updatedAddress := decodeResponse[address](t, updateResponse, http.StatusOK)
	if updatedAddress.ID != created.Address.ID ||
		updatedAddress.Country != updatePayload["country"] ||
		updatedAddress.City != updatePayload["city"] ||
		updatedAddress.Street != updatePayload["street"] {
		t.Fatalf("PATCH address = %#v, want ID %s and payload %#v", updatedAddress, created.Address.ID, updatePayload)
	}

	getUpdatedResponse := api.doJSON(t, http.MethodGet, wantLocation, nil)
	updatedSupplier := decodeResponse[supplier](t, getUpdatedResponse, http.StatusOK)
	if updatedSupplier.Address != updatedAddress {
		t.Fatalf("supplier address after PATCH = %#v, want %#v", updatedSupplier.Address, updatedAddress)
	}

	invalidIDResponse := api.doJSON(t, http.MethodGet, "/api/v1/suppliers/not-a-uuid", nil)
	requireError(t, invalidIDResponse, http.StatusBadRequest, "INVALID_REQUEST")

	missingResponse := api.doJSON(t, http.MethodGet, "/api/v1/suppliers/"+uuid.New().String(), nil)
	requireError(t, missingResponse, http.StatusNotFound, "NOT_FOUND")

	deleteResponse := api.doJSON(t, http.MethodDelete, wantLocation, nil)
	requireNoContent(t, deleteResponse)
	deleted = true

	repeatedDeleteResponse := api.doJSON(t, http.MethodDelete, wantLocation, nil)
	requireNoContent(t, repeatedDeleteResponse)

	deletedResponse := api.doJSON(t, http.MethodGet, wantLocation, nil)
	requireError(t, deletedResponse, http.StatusNotFound, "NOT_FOUND")
}

func newAPIClient(t *testing.T) *apiClient {
	t.Helper()

	baseURL := strings.TrimRight(os.Getenv("E2E_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	api := &apiClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}

	deadline := time.Now().Add(30 * time.Second)
	for {
		response, err := api.client.Get(baseURL + "/api/health/live")
		if err == nil && response.StatusCode == http.StatusOK {
			response.Body.Close()
			return api
		}
		if response != nil {
			response.Body.Close()
		}
		if time.Now().After(deadline) {
			t.Fatalf("API at %s did not become ready within 30s: %v", baseURL, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (api *apiClient) doJSON(t *testing.T, method, path string, payload any) *http.Response {
	t.Helper()

	var body io.Reader
	if payload != nil {
		if raw, ok := payload.([]byte); ok {
			body = bytes.NewReader(raw)
		} else {
			encoded, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("encode %s %s payload: %v", method, path, err)
			}
			body = bytes.NewReader(encoded)
		}
	}

	request, err := http.NewRequest(method, api.baseURL+path, body)
	if err != nil {
		t.Fatalf("create %s %s request: %v", method, path, err)
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := api.client.Do(request)
	if err != nil {
		t.Fatalf("send %s %s request: %v", method, path, err)
	}

	return response
}

func decodeResponse[T any](t *testing.T, response *http.Response, wantStatus int) T {
	t.Helper()
	defer response.Body.Close()

	body := readBody(t, response)
	if response.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", response.StatusCode, wantStatus, body)
	}
	if contentType := response.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("decode response %q: %v", body, err)
	}
	return result
}

func requireError(t *testing.T, response *http.Response, wantStatus int, wantCode string) {
	t.Helper()

	got := decodeResponse[errorResponse](t, response, wantStatus)
	if got.Code != wantCode || got.Message == "" {
		t.Fatalf("error response = %#v, want code %q and a message", got, wantCode)
	}
}

func requireNoContent(t *testing.T, response *http.Response) {
	t.Helper()
	defer response.Body.Close()

	body := readBody(t, response)
	if response.StatusCode != http.StatusNoContent || len(body) != 0 {
		t.Fatalf("response = (%d, %q), want (%d, empty body)", response.StatusCode, body, http.StatusNoContent)
	}
}

func readBody(t *testing.T, response *http.Response) []byte {
	t.Helper()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return body
}
