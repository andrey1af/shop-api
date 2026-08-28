package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
)

type productRepositoryStub struct {
	createInput models.Product
	products    []models.Product
	getErr      error
	stockErr    error
}

func (stub *productRepositoryStub) Create(_ context.Context, product models.Product) (models.Product, error) {
	stub.createInput = product
	return product, nil
}

func (stub *productRepositoryStub) ListAvailable(context.Context) ([]models.Product, error) {
	return stub.products, nil
}

func (stub *productRepositoryStub) GetByID(context.Context, uuid.UUID) (models.Product, error) {
	return models.Product{}, stub.getErr
}

func (stub *productRepositoryStub) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (stub *productRepositoryStub) DecreaseStock(context.Context, uuid.UUID, int64) (models.Product, error) {
	return models.Product{}, stub.stockErr
}

func TestCreateProductAssignsIDAndLastUpdateDate(t *testing.T) {
	repository := &productRepositoryStub{}
	service := NewProductService(repository)
	input := models.ProductCreate{
		Name:           "Refrigerator",
		Category:       "Appliances",
		Price:          54990,
		AvailableStock: 12,
		SupplierID:     uuid.New(),
	}

	got, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.ID == uuid.Nil() {
		t.Fatal("Create() did not assign product ID")
	}
	if _, err := time.Parse(time.DateOnly, got.LastUpdateDate); err != nil {
		t.Fatalf("Create() last update date = %q: %v", got.LastUpdateDate, err)
	}
	if got != repository.createInput {
		t.Fatalf("Create() = %#v, repository received %#v", got, repository.createInput)
	}
}

func TestGetAvailableProductsReturnsEmptyArray(t *testing.T) {
	service := NewProductService(&productRepositoryStub{})

	got, err := service.GetAvailable(context.Background())
	if err != nil {
		t.Fatalf("GetAvailable() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("GetAvailable() = %#v, want non-nil empty slice", got)
	}
}

func TestProductErrorsCanBeInspectedThroughService(t *testing.T) {
	repository := &productRepositoryStub{getErr: ErrProductNotFound, stockErr: ErrInsufficientStock}
	service := NewProductService(repository)

	if _, err := service.GetByID(context.Background(), uuid.New()); !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("GetByID() error = %v, want ErrProductNotFound", err)
	}
	if _, err := service.DecreaseStock(context.Background(), uuid.New(), 1); !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("DecreaseStock() error = %v, want ErrInsufficientStock", err)
	}
}
