package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
	"github.com/andrey1af/shop-api/api-service/internal/services"
)

const maxImageSize = 10 << 20

type imageService interface {
	Create(context.Context, uuid.UUID, []byte) (models.ImageMetadata, error)
	GetByID(context.Context, uuid.UUID) (models.Image, error)
	GetByProductID(context.Context, uuid.UUID) (models.Image, error)
	Replace(context.Context, uuid.UUID, []byte) error
	Delete(context.Context, uuid.UUID) error
}

type imageHandler struct {
	image imageService
}

func (h *imageHandler) createForProduct(w http.ResponseWriter, r *http.Request) {
	productID, ok := parseProductID(w, r)
	if !ok {
		return
	}
	data, ok := readImage(w, r)
	if !ok {
		return
	}

	metadata, err := h.image.Create(r.Context(), productID, data)
	if errors.Is(err, services.ErrProductNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{Code: "PRODUCT_NOT_FOUND", Message: "Product not found"})
		return
	}
	if errors.Is(err, services.ErrImageAlreadyExists) {
		writeJSON(w, http.StatusConflict, errorResponse{Code: "IMAGE_ALREADY_EXISTS", Message: "Product already has an image"})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Location", "/api/v1/images/"+metadata.ID.String())
	writeJSON(w, http.StatusCreated, metadata)
}

func (h *imageHandler) getForProduct(w http.ResponseWriter, r *http.Request) {
	productID, ok := parseProductID(w, r)
	if !ok {
		return
	}

	image, err := h.image.GetByProductID(r.Context(), productID)
	if errors.Is(err, services.ErrImageNotFound) {
		writeError(w, http.StatusNotFound, "Image not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeImage(w, image)
}

func (h *imageHandler) getByID(w http.ResponseWriter, r *http.Request) {
	imageID, ok := parseImageID(w, r)
	if !ok {
		return
	}

	image, err := h.image.GetByID(r.Context(), imageID)
	if errors.Is(err, services.ErrImageNotFound) {
		writeError(w, http.StatusNotFound, "Image not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeImage(w, image)
}

func (h *imageHandler) replace(w http.ResponseWriter, r *http.Request) {
	imageID, ok := parseImageID(w, r)
	if !ok {
		return
	}
	data, ok := readImage(w, r)
	if !ok {
		return
	}

	err := h.image.Replace(r.Context(), imageID, data)
	if errors.Is(err, services.ErrImageNotFound) {
		writeError(w, http.StatusNotFound, "Image not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *imageHandler) delete(w http.ResponseWriter, r *http.Request) {
	imageID, ok := parseImageID(w, r)
	if !ok {
		return
	}

	if err := h.image.Delete(r.Context(), imageID); err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseImageID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	imageID, err := uuid.Parse(r.PathValue("imageId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid image ID")
		return uuid.Nil(), false
	}

	return imageID, true
}

func readImage(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	body := http.MaxBytesReader(w, r.Body, maxImageSize)
	data, err := io.ReadAll(body)
	if err != nil || len(data) == 0 {
		writeError(w, http.StatusBadRequest, "Image must contain between 1 byte and 10 MiB")
		return nil, false
	}

	return data, true
}

func writeImage(w http.ResponseWriter, image models.Image) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="image-%s.bin"`, image.ID))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(image.Data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(image.Data)
}
