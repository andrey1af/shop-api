package handlers

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
	"github.com/andrey1af/shop-api/api-service/internal/services"
)

type clientService interface {
	Create(context.Context, models.ClientCreate) (models.Client, error)
	GetAll(context.Context, models.ClientFilter) ([]models.Client, error)
	Delete(context.Context, uuid.UUID) error
	GetAddress(context.Context, uuid.UUID) (models.Address, error)
	UpdateAddress(context.Context, uuid.UUID, models.Address) (models.Address, error)
}

type clientHandler struct {
	client clientService
}

func (h *clientHandler) create(w http.ResponseWriter, r *http.Request) {
	var candidate models.ClientCreate
	if err := decodeJSON(w, r, &candidate); err != nil || !validClientCreate(candidate) {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	client, err := h.client.Create(r.Context(), candidate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Location", "/api/v1/clients/"+client.ID.String())
	writeJSON(w, http.StatusCreated, client)
}

func (h *clientHandler) getAll(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseClientFilter(w, r)
	if !ok {
		return
	}

	clients, err := h.client.GetAll(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, clients)
}

func (h *clientHandler) delete(w http.ResponseWriter, r *http.Request) {
	clientID, ok := parseClientID(w, r)
	if !ok {
		return
	}

	if err := h.client.Delete(r.Context(), clientID); err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *clientHandler) getAddress(w http.ResponseWriter, r *http.Request) {
	clientID, ok := parseClientID(w, r)
	if !ok {
		return
	}

	address, err := h.client.GetAddress(r.Context(), clientID)
	if errors.Is(err, services.ErrClientNotFound) {
		writeError(w, http.StatusNotFound, "Client not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, address)
}

func (h *clientHandler) updateAddress(w http.ResponseWriter, r *http.Request) {
	clientID, ok := parseClientID(w, r)
	if !ok {
		return
	}

	var candidate models.AddressCreate
	if err := decodeJSON(w, r, &candidate); err != nil || !validAddressCreate(candidate) {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	address, err := h.client.UpdateAddress(r.Context(), clientID, models.Address{
		Country: candidate.Country,
		City:    candidate.City,
		Street:  candidate.Street,
	})
	if errors.Is(err, services.ErrClientNotFound) {
		writeError(w, http.StatusNotFound, "Client not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, address)
}

func parseClientID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	clientID, err := uuid.Parse(r.PathValue("clientId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid client ID")
		return uuid.Nil(), false
	}

	return clientID, true
}

func parseClientFilter(w http.ResponseWriter, r *http.Request) (models.ClientFilter, bool) {
	query := r.URL.Query()
	name, hasName := query["client_name"]
	surname, hasSurname := query["client_surname"]
	if hasName != hasSurname ||
		(hasName && (len(name) != 1 || len(surname) != 1 || !validText(name[0], 100) || !validText(surname[0], 100))) {
		writeError(w, http.StatusBadRequest, "client_name and client_surname must be provided together")
		return models.ClientFilter{}, false
	}

	filter := models.ClientFilter{}
	if hasName {
		filter.ClientName = &name[0]
		filter.ClientSurname = &surname[0]
	}

	var ok bool
	filter.Limit, ok = parseOptionalInt(query["limit"], 1)
	if !ok {
		writeError(w, http.StatusBadRequest, "Invalid limit")
		return models.ClientFilter{}, false
	}
	filter.Offset, ok = parseOptionalInt(query["offset"], 0)
	if !ok {
		writeError(w, http.StatusBadRequest, "Invalid offset")
		return models.ClientFilter{}, false
	}

	return filter, true
}

func parseOptionalInt(values []string, minimum int) (*int, bool) {
	if len(values) == 0 {
		return nil, true
	}
	if len(values) != 1 {
		return nil, false
	}

	value, err := strconv.Atoi(values[0])
	if err != nil || value < minimum || value > math.MaxInt32 {
		return nil, false
	}

	return &value, true
}

func validClientCreate(candidate models.ClientCreate) bool {
	if !validText(candidate.ClientName, 100) ||
		!validText(candidate.ClientSurname, 100) ||
		!validText(candidate.Gender, 50) ||
		!validAddressCreate(candidate.Address) {
		return false
	}

	birthday, err := time.Parse(time.DateOnly, candidate.Birthday)
	if err != nil || birthday.Format(time.DateOnly) != candidate.Birthday {
		return false
	}

	registrationDateValue := candidate.RegistrationDate
	if registrationDateValue == "" {
		registrationDateValue = time.Now().Format(time.DateOnly)
	}

	registrationDate, err := time.Parse(time.DateOnly, registrationDateValue)
	return err == nil &&
		registrationDate.Format(time.DateOnly) == registrationDateValue &&
		!birthday.After(registrationDate)
}

func validAddressCreate(candidate models.AddressCreate) bool {
	return validText(candidate.Country, 100) &&
		validText(candidate.City, 100) &&
		validText(candidate.Street, 255)
}

func validText(value string, maxLength int) bool {
	return strings.TrimSpace(value) != "" && utf8.RuneCountInString(value) <= maxLength
}
