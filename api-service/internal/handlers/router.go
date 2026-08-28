package handlers

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewRouter(supplierService supplierService, clientServices ...clientService) http.Handler {
	var client clientService
	if len(clientServices) > 0 {
		client = clientServices[0]
	}

	return newRouter(supplierService, client, nil, nil)
}

func NewRouterWithProducts(
	supplier supplierService,
	client clientService,
	product productService,
	image imageService,
) http.Handler {
	return newRouter(supplier, client, product, image)
}

func newRouter(supplier supplierService, client clientService, product productService, image imageService) http.Handler {
	supplierHandler := &supplierHandler{supplier: supplier}

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

	if client != nil {
		clientHandler := &clientHandler{client: client}
		mux.HandleFunc("POST /api/v1/clients", clientHandler.create)
		mux.HandleFunc("GET /api/v1/clients", clientHandler.getAll)
		mux.HandleFunc("DELETE /api/v1/clients/{clientId}", clientHandler.delete)
		mux.HandleFunc("GET /api/v1/clients/{clientId}/address", clientHandler.getAddress)
		mux.HandleFunc("PATCH /api/v1/clients/{clientId}/address", clientHandler.updateAddress)
	}

	if product != nil {
		productHandler := &productHandler{product: product}
		mux.HandleFunc("POST /api/v1/products", productHandler.create)
		mux.HandleFunc("GET /api/v1/products", productHandler.getAvailable)
		mux.HandleFunc("GET /api/v1/products/{productId}", productHandler.getByID)
		mux.HandleFunc("DELETE /api/v1/products/{productId}", productHandler.delete)
		mux.HandleFunc("PATCH /api/v1/products/{productId}/stock", productHandler.decreaseStock)
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
