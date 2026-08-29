package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/andrey1af/shop-api/api-service/internal/idempotency"
)

type readinessChecker interface {
	Ping(context.Context) error
}

type RouterInfrastructure struct {
	Readiness   readinessChecker
	Idempotency idempotency.Backend
	OpenAPI     []byte
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewRouter(supplierService supplierService, clientServices ...clientService) http.Handler {
	var client clientService
	if len(clientServices) > 0 {
		client = clientServices[0]
	}

	return newRouter(supplierService, client, nil, nil, RouterInfrastructure{})
}

func NewRouterWithProducts(
	supplier supplierService,
	client clientService,
	product productService,
	image imageService,
) http.Handler {
	return newRouter(supplier, client, product, image, RouterInfrastructure{})
}

func NewApplicationRouter(
	supplier supplierService,
	client clientService,
	product productService,
	image imageService,
	infrastructure RouterInfrastructure,
) http.Handler {
	return newRouter(supplier, client, product, image, infrastructure)
}

func newRouter(
	supplier supplierService,
	client clientService,
	product productService,
	image imageService,
	infrastructure RouterInfrastructure,
) http.Handler {
	supplierHandler := &supplierHandler{supplier: supplier}

	mux := http.NewServeMux()

	// health and API documentation
	mux.HandleFunc("GET /api/v1/health", healthCheck(infrastructure.Readiness, http.StatusInternalServerError))
	mux.HandleFunc("GET /api/v1/health/live", healthLive)
	mux.HandleFunc("GET /api/v1/health/ready", healthCheck(infrastructure.Readiness, http.StatusServiceUnavailable))
	// Backward-compatible liveness route used by existing deployments.
	mux.HandleFunc("GET /api/health/live", healthLive)
	if len(infrastructure.OpenAPI) > 0 {
		mux.HandleFunc("GET /swagger/index.html", swaggerUI)
		mux.HandleFunc("GET /swagger/openapi.yaml", serveOpenAPI(infrastructure.OpenAPI))
	}
	// supplier
	mux.HandleFunc("POST /api/v1/suppliers", idempotent(infrastructure.Idempotency, supplierHandler.create))
	mux.HandleFunc("GET /api/v1/suppliers", supplierHandler.getAll)
	mux.HandleFunc("GET /api/v1/suppliers/{supplierId}", supplierHandler.getByID)
	mux.HandleFunc("DELETE /api/v1/suppliers/{supplierId}", supplierHandler.delete)
	mux.HandleFunc("GET /api/v1/suppliers/{supplierId}/address", supplierHandler.getAddress)
	mux.HandleFunc("PATCH /api/v1/suppliers/{supplierId}/address", supplierHandler.updateAddress)

	if client != nil {
		clientHandler := &clientHandler{client: client}
		mux.HandleFunc("POST /api/v1/clients", idempotent(infrastructure.Idempotency, clientHandler.create))
		mux.HandleFunc("GET /api/v1/clients", clientHandler.getAll)
		mux.HandleFunc("DELETE /api/v1/clients/{clientId}", clientHandler.delete)
		mux.HandleFunc("GET /api/v1/clients/{clientId}/address", clientHandler.getAddress)
		mux.HandleFunc("PATCH /api/v1/clients/{clientId}/address", clientHandler.updateAddress)
	}

	if product != nil {
		productHandler := &productHandler{product: product}
		mux.HandleFunc("POST /api/v1/products", idempotent(infrastructure.Idempotency, productHandler.create))
		mux.HandleFunc("GET /api/v1/products", productHandler.getAvailable)
		mux.HandleFunc("GET /api/v1/products/{productId}", productHandler.getByID)
		mux.HandleFunc("DELETE /api/v1/products/{productId}", productHandler.delete)
		mux.HandleFunc("PATCH /api/v1/products/{productId}/stock", idempotent(infrastructure.Idempotency, productHandler.decreaseStock))
	}

	if image != nil {
		imageHandler := &imageHandler{image: image}
		mux.HandleFunc("POST /api/v1/products/{productId}/image", imageHandler.createForProduct)
		mux.HandleFunc("GET /api/v1/products/{productId}/image", imageHandler.getForProduct)
		mux.HandleFunc("GET /api/v1/images/{imageId}", imageHandler.getByID)
		mux.HandleFunc("PUT /api/v1/images/{imageId}", imageHandler.replace)
		mux.HandleFunc("DELETE /api/v1/images/{imageId}", imageHandler.delete)
	}
	return mux
}

func idempotent(backend idempotency.Backend, handler http.HandlerFunc) http.HandlerFunc {
	if backend == nil {
		return handler
	}
	return idempotency.Require(backend, handler)
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
