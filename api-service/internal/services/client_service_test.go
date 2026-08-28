package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
)

type clientRepositoryStub struct {
	created     models.Client
	createInput models.Client
	clients     []models.Client
	filter      models.ClientFilter
	addressErr  error
}

func (stub *clientRepositoryStub) Create(_ context.Context, candidate models.Client) (models.Client, error) {
	stub.createInput = candidate
	if stub.created.ID != uuid.Nil() {
		return stub.created, nil
	}

	return candidate, nil
}

func (stub *clientRepositoryStub) List(_ context.Context, filter models.ClientFilter) ([]models.Client, error) {
	stub.filter = filter
	return stub.clients, nil
}

func (stub *clientRepositoryStub) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (stub *clientRepositoryStub) GetAddress(context.Context, uuid.UUID) (models.Address, error) {
	return models.Address{}, stub.addressErr
}

func (stub *clientRepositoryStub) UpdateAddress(context.Context, uuid.UUID, models.Address) (models.Address, error) {
	return models.Address{}, stub.addressErr
}

func TestCreateClientAssignsIDsAndRegistrationDate(t *testing.T) {
	repository := &clientRepositoryStub{}
	service := NewClientService(repository)
	input := models.ClientCreate{
		ClientName:    "Ivan",
		ClientSurname: "Ivanov",
		Birthday:      "1990-05-17",
		Gender:        "male",
		Address: models.AddressCreate{
			Country: "Russia",
			City:    "Moscow",
			Street:  "Tverskaya, 10",
		},
	}

	got, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.ID == uuid.Nil() || got.Address.ID == uuid.Nil() {
		t.Fatalf("Create() IDs = (%s, %s), want non-zero IDs", got.ID, got.Address.ID)
	}
	if _, err := time.Parse(time.DateOnly, got.RegistrationDate); err != nil {
		t.Fatalf("Create() registration date = %q, want YYYY-MM-DD: %v", got.RegistrationDate, err)
	}
	if got != repository.createInput {
		t.Fatalf("Create() = %#v, repository received %#v", got, repository.createInput)
	}
}

func TestGetClientsPassesFilterAndReturnsEmptyArray(t *testing.T) {
	repository := &clientRepositoryStub{}
	service := NewClientService(repository)
	name, surname, limit, offset := "Ivan", "Ivanov", 10, 2
	filter := models.ClientFilter{
		ClientName:    &name,
		ClientSurname: &surname,
		Limit:         &limit,
		Offset:        &offset,
	}

	got, err := service.GetAll(context.Background(), filter)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("GetAll() = %#v, want non-nil empty slice", got)
	}
	if repository.filter.ClientName == nil || *repository.filter.ClientName != name ||
		repository.filter.ClientSurname == nil || *repository.filter.ClientSurname != surname ||
		repository.filter.Limit == nil || *repository.filter.Limit != limit ||
		repository.filter.Offset == nil || *repository.filter.Offset != offset {
		t.Fatalf("repository filter = %#v, want %#v", repository.filter, filter)
	}
}

func TestClientNotFoundCanBeInspectedThroughServiceError(t *testing.T) {
	service := NewClientService(&clientRepositoryStub{addressErr: ErrClientNotFound})

	if _, err := service.GetAddress(context.Background(), uuid.New()); !errors.Is(err, ErrClientNotFound) {
		t.Fatalf("GetAddress() error = %v, want wrapped ErrClientNotFound", err)
	}
}
