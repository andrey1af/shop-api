package handlers

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
	"github.com/andrey1af/shop-api/api-service/internal/services"
)

type supplierService interface {
	Create(context.Context, models.SupplierCreate) (models.Supplier, error)
	GetAll(context.Context) ([]models.Supplier, error)
	GetByID(context.Context, uuid.UUID) (models.Supplier, error)
	Delete(context.Context, uuid.UUID) error
	GetAddress(context.Context, uuid.UUID) (models.Address, error)
	UpdateAddress(context.Context, uuid.UUID, models.Address) (models.Address, error)
}

type supplierHandler struct {
	supplier supplierService
}

func (h *supplierHandler) create(w http.ResponseWriter, r *http.Request) {
	var supplierCreate models.SupplierCreate

	if err := decodeJSON(w, r, &supplierCreate); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	supplier, err := h.supplier.Create(r.Context(), supplierCreate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, supplier)
}

func (h *supplierHandler) getAll(w http.ResponseWriter, r *http.Request) {
	suppliers, err := h.supplier.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, suppliers)
}

func (h *supplierHandler) getByID(w http.ResponseWriter, r *http.Request) {
	supplierID, ok := parseSupplierID(w, r)
	if !ok {
		return
	}

	supplier, err := h.supplier.GetByID(r.Context(), supplierID)
	if errors.Is(err, services.ErrSupplierNotFound) {
		writeError(w, http.StatusNotFound, "Supplier not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, supplier)
}

func (h *supplierHandler) delete(w http.ResponseWriter, r *http.Request) {
	supplierID, ok := parseSupplierID(w, r)
	if !ok {
		return
	}

	err := h.supplier.Delete(r.Context(), supplierID)
	if errors.Is(err, services.ErrSupplierInUse) {
		writeJSON(w, http.StatusConflict, errorResponse{
			Code:    "SUPPLIER_IN_USE",
			Message: "Supplier is used by products",
		})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *supplierHandler) getAddress(w http.ResponseWriter, r *http.Request) {
	supplierID, ok := parseSupplierID(w, r)
	if !ok {
		return
	}

	address, err := h.supplier.GetAddress(r.Context(), supplierID)
	if errors.Is(err, services.ErrSupplierNotFound) {
		writeError(w, http.StatusNotFound, "Supplier not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, address)
}

func (h *supplierHandler) updateAddress(w http.ResponseWriter, r *http.Request) {
	supplierID, ok := parseSupplierID(w, r)
	if !ok {
		return
	}

	var addressCreate models.AddressCreate
	if err := decodeJSON(w, r, &addressCreate); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	address, err := h.supplier.UpdateAddress(r.Context(), supplierID, models.Address{
		Country: addressCreate.Country,
		City:    addressCreate.City,
		Street:  addressCreate.Street,
	})
	if errors.Is(err, services.ErrSupplierNotFound) {
		writeError(w, http.StatusNotFound, "Supplier not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, address)
}

func parseSupplierID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	supplierID, err := uuid.Parse(r.PathValue("supplierId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid supplier ID")
		return uuid.Nil(), false
	}

	return supplierID, true
}
