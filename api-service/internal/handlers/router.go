package handlers

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewRouter(supplierService supplierService) http.Handler {
	supplierHandler := &supplierHandler{supplier: supplierService}

	mux := http.NewServeMux()

	// health
	mux.HandleFunc("GET /api/health/live", healthLive)
	// supplier
	mux.HandleFunc("POST /api/v1/suppliers", supplierHandler.create)
	mux.HandleFunc("GET /api/v1/suppliers", supplierHandler.getAll)
	mux.HandleFunc("GET /api/v1/suppliers/{supplierId}", supplierHandler.getByID)
	mux.HandleFunc("DELETE /api/v1/suppliers/{supplierId}", supplierHandler.delete)
	mux.HandleFunc("GET /api/v1/suppliers/{supplierId}/address", supplierHandler.getAddress)
	mux.HandleFunc("PATCH /api/v1/suppliers/{supplierId}/address", supplierHandler.updateAddress)
	return mux
}

func decodeJSON(response http.ResponseWriter, request *http.Request, target any) error {
	decoder := json.NewDecoder(
		http.MaxBytesReader(response, request.Body, 1<<20),
	)
	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{
		Code:    errorCode(status),
		Message: message,
	})
}

func errorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "INVALID_REQUEST"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	default:
		return "INTERNAL_ERROR"
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
