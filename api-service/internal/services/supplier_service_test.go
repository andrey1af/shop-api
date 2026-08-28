package services

import (
	"context"
	"errors"
	"testing"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
)

type repositoryStub struct {
	created        models.Supplier
	createErr      error
	all            []models.Supplier
	allErr         error
	byID           models.Supplier
	byIDErr        error
	deleteErr      error
	address        models.Address
	addressErr     error
	updated        models.Address
	updateErr      error
	createCalls    int
	updateCalls    int
	createInput    models.Supplier
	updateInput    models.Address
	lastSupplierID uuid.UUID
}

func (stub *repositoryStub) Create(_ context.Context, candidate models.Supplier) (models.Supplier, error) {
	stub.createCalls++
	stub.createInput = candidate
	if stub.createErr != nil {
		return models.Supplier{}, stub.createErr
	}
	if stub.created.ID != uuid.Nil() {
		return stub.created, nil
	}

	return candidate, nil
}

func (stub *repositoryStub) List(context.Context) ([]models.Supplier, error) {
	return stub.all, stub.allErr
}

func (stub *repositoryStub) GetByID(_ context.Context, supplierID uuid.UUID) (models.Supplier, error) {
	stub.lastSupplierID = supplierID
	return stub.byID, stub.byIDErr
}

func (stub *repositoryStub) Delete(_ context.Context, supplierID uuid.UUID) error {
	stub.lastSupplierID = supplierID
	return stub.deleteErr
}

func (stub *repositoryStub) GetAddress(_ context.Context, supplierID uuid.UUID) (models.Address, error) {
	stub.lastSupplierID = supplierID
	return stub.address, stub.addressErr
}

func (stub *repositoryStub) UpdateAddress(_ context.Context, supplierID uuid.UUID, address models.Address) (models.Address, error) {
	stub.updateCalls++
	stub.lastSupplierID = supplierID
	stub.updateInput = address
	return stub.updated, stub.updateErr
}

func TestCreatePassesSupplierToRepository(t *testing.T) {
	repository := &repositoryStub{}
	service := NewSupplierService(repository)

	input := validSupplierCreate()

	got, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repository.createCalls != 1 {
		t.Fatalf("repository Create() calls = %d, want 1", repository.createCalls)
	}
	if got != repository.createInput {
		t.Fatalf("Create() = %#v, repository received %#v", got, repository.createInput)
	}
	if got.ID == uuid.Nil() {
		t.Fatal("Create() did not assign supplier ID")
	}
	if got.Address.ID == uuid.Nil() {
		t.Fatal("Create() did not assign address ID")
	}
	if got.Name != input.Name || got.PhoneNumber != input.PhoneNumber {
		t.Fatalf("repository supplier fields = %#v, want input %#v", got, input)
	}
	if got.Address.Country != input.Address.Country || got.Address.City != input.Address.City || got.Address.Street != input.Address.Street {
		t.Fatalf("repository address = %#v, want input %#v", got.Address, input.Address)
	}
}

func TestGetAllReturnsNonNilEmptySlice(t *testing.T) {
	service := NewSupplierService(&repositoryStub{})

	got, err := service.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("GetAll() = %#v, want non-nil empty slice", got)
	}
}

func TestRepositoryErrorsCanBeInspectedThroughServiceErrors(t *testing.T) {
	supplierID := uuid.New()
	deleteErr := errors.New("delete supplier")
	repository := &repositoryStub{
		byIDErr:   ErrSupplierNotFound,
		deleteErr: deleteErr,
		updateErr: ErrSupplierNotFound,
	}
	service := NewSupplierService(repository)

	if _, err := service.GetByID(context.Background(), supplierID); !errors.Is(err, ErrSupplierNotFound) {
		t.Fatalf("GetByID() error = %v, want wrapped ErrSupplierNotFound", err)
	}
	if err := service.Delete(context.Background(), supplierID); !errors.Is(err, deleteErr) {
		t.Fatalf("Delete() error = %v, want wrapped repository error", err)
	}
	if _, err := service.UpdateAddress(context.Background(), supplierID, validSupplier().Address); !errors.Is(err, ErrSupplierNotFound) {
		t.Fatalf("UpdateAddress() error = %v, want wrapped ErrSupplierNotFound", err)
	}
}

func TestUpdateAddressPassesInputToRepository(t *testing.T) {
	supplierID := uuid.New()
	want := models.Address{ID: uuid.New(), Country: "Russia", City: "Moscow", Street: "Tverskaya, 1"}
	repository := &repositoryStub{updated: want}
	service := NewSupplierService(repository)

	input := models.Address{Country: "Russia", City: "Moscow", Street: "Tverskaya, 1"}
	got, err := service.UpdateAddress(context.Background(), supplierID, input)
	if err != nil {
		t.Fatalf("UpdateAddress() error = %v", err)
	}
	if got != want {
		t.Fatalf("UpdateAddress() = %#v, want %#v", got, want)
	}
	if repository.updateCalls != 1 || repository.lastSupplierID != supplierID {
		t.Fatalf("repository UpdateAddress() calls = %d, supplier ID = %s", repository.updateCalls, repository.lastSupplierID)
	}
	if repository.updateInput != input {
		t.Fatalf("repository UpdateAddress() input = %#v, want %#v", repository.updateInput, input)
	}
}

func validSupplier() models.Supplier {
	return models.Supplier{
		Name:        "Supplier",
		PhoneNumber: "+7 999 123-45-67",
		Address: models.Address{
			Country: "Russia",
			City:    "Moscow",
			Street:  "Tverskaya, 1",
		},
	}
}

func validSupplierCreate() models.SupplierCreate {
	return models.SupplierCreate{
		Name:        "Supplier",
		PhoneNumber: "+7 999 123-45-67",
		Address: models.AddressCreate{
			Country: "Russia",
			City:    "Moscow",
			Street:  "Tverskaya, 1",
		},
	}
}
