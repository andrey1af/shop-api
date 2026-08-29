package handlers

import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
	"github.com/andrey1af/shop-api/api-service/internal/services"
)

type productService interface {
	Create(context.Context, models.ProductCreate) (models.Product, error)
	GetAvailable(context.Context) ([]models.Product, error)
	GetByID(context.Context, uuid.UUID) (models.Product, error)
	Delete(context.Context, uuid.UUID) error
	DecreaseStock(context.Context, uuid.UUID, int64) (models.Product, error)
}

type productHandler struct {
	product productService
}

func (h *productHandler) create(w http.ResponseWriter, r *http.Request) {
	var candidate models.ProductCreate
	if err := decodeJSON(w, r, &candidate); err != nil || !validProductCreate(candidate) {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	product, err := h.product.Create(r.Context(), candidate)
	if errors.Is(err, services.ErrSupplierNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{Code: "SUPPLIER_NOT_FOUND", Message: "Supplier not found"})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Location", "/api/v1/products/"+product.ID.String())
	writeJSON(w, http.StatusCreated, product)
}

func (h *productHandler) getAvailable(w http.ResponseWriter, r *http.Request) {
	products, err := h.product.GetAvailable(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, products)
}

func (h *productHandler) getByID(w http.ResponseWriter, r *http.Request) {
	productID, ok := parseProductID(w, r)
	if !ok {
		return
	}

	product, err := h.product.GetByID(r.Context(), productID)
	if errors.Is(err, services.ErrProductNotFound) {
		writeError(w, http.StatusNotFound, "Product not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func (h *productHandler) delete(w http.ResponseWriter, r *http.Request) {
	productID, ok := parseProductID(w, r)
	if !ok {
		return
	}

	if err := h.product.Delete(r.Context(), productID); err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *productHandler) decreaseStock(w http.ResponseWriter, r *http.Request) {
	productID, ok := parseProductID(w, r)
	if !ok {
		return
	}

	var candidate models.StockDecrease
	if err := decodeJSON(w, r, &candidate); err != nil || candidate.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	product, err := h.product.DecreaseStock(r.Context(), productID, candidate.Quantity)
	if errors.Is(err, services.ErrProductNotFound) {
		writeError(w, http.StatusNotFound, "Product not found")
		return
	}
	if errors.Is(err, services.ErrInsufficientStock) {
		writeJSON(w, http.StatusConflict, errorResponse{Code: "INSUFFICIENT_STOCK", Message: "Insufficient stock"})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func parseProductID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	productID, err := uuid.Parse(r.PathValue("productId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid product ID")
		return uuid.Nil(), false
	}

	return productID, true
}

func validProductCreate(candidate models.ProductCreate) bool {
	if !validText(candidate.Name, 255) ||
		!validText(candidate.Category, 100) ||
		candidate.Price == nil ||
		candidate.AvailableStock == nil ||
		*candidate.Price < 0 || math.IsInf(*candidate.Price, 0) || math.IsNaN(*candidate.Price) ||
		*candidate.Price > 9_999_999_999.99 ||
		*candidate.AvailableStock < 0 ||
		candidate.SupplierID == uuid.Nil() {
		return false
	}

	if candidate.LastUpdateDate == "" {
		return true
	}
	lastUpdateDate, err := time.Parse(time.DateOnly, candidate.LastUpdateDate)
	return err == nil && lastUpdateDate.Format(time.DateOnly) == candidate.LastUpdateDate
}
