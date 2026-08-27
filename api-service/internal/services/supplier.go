package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/andrey1af/shop-api/api-service/internal/models"
	"github.com/google/uuid"
)

var (
	ErrSupplierNotFound = errors.New("supplier not found")
	ErrSupplierInUse    = errors.New("supplier is used by products")
)

type SupplierRepository interface {
	Create(context.Context, models.Supplier) (models.Supplier, error)
	List(context.Context) ([]models.Supplier, error)
	GetByID(context.Context, uuid.UUID) (models.Supplier, error)
	Delete(context.Context, uuid.UUID) error
	GetAddress(context.Context, uuid.UUID) (models.Address, error)
	UpdateAddress(context.Context, uuid.UUID, models.Address) (models.Address, error)
}

type SupplierService struct {
	repository SupplierRepository
}

func NewSupplierService(repository SupplierRepository) *SupplierService {
	return &SupplierService{repository: repository}
}

func (service *SupplierService) Create(ctx context.Context, candidate models.SupplierCreate) (models.Supplier, error) {
	addressID := uuid.New()
	supplier := models.Supplier{
		ID:          uuid.New(),
		Name:        candidate.Name,
		PhoneNumber: candidate.PhoneNumber,
		Address: models.Address{
			ID:      addressID,
			Country: candidate.Address.Country,
			City:    candidate.Address.City,
			Street:  candidate.Address.Street,
		},
	}

	created, err := service.repository.Create(ctx, supplier)
	if err != nil {
		return models.Supplier{}, fmt.Errorf("create supplier: %w", err)
	}

	return created, nil
}

func (service *SupplierService) GetAll(ctx context.Context) ([]models.Supplier, error) {
	suppliers, err := service.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all suppliers: %w", err)
	}

	if suppliers == nil {
		return []models.Supplier{}, nil
	}

	return suppliers, nil
}

func (service *SupplierService) GetByID(ctx context.Context, supplierID uuid.UUID) (models.Supplier, error) {
	if supplierID == uuid.Nil {
		return models.Supplier{}, ErrSupplierNotFound
	}

	supplier, err := service.repository.GetByID(ctx, supplierID)
	if err != nil {
		return models.Supplier{}, fmt.Errorf("get supplier %s: %w", supplierID, err)
	}

	return supplier, nil
}

func (service *SupplierService) Delete(ctx context.Context, supplierID uuid.UUID) error {
	if err := service.repository.Delete(ctx, supplierID); err != nil {
		return fmt.Errorf("delete supplier %s: %w", supplierID, err)
	}

	return nil
}

func (service *SupplierService) GetAddress(ctx context.Context, supplierID uuid.UUID) (models.Address, error) {
	address, err := service.repository.GetAddress(ctx, supplierID)
	if err != nil {
		return models.Address{}, fmt.Errorf("get supplier %s address: %w", supplierID, err)
	}

	return address, nil
}

func (service *SupplierService) UpdateAddress(ctx context.Context, supplierID uuid.UUID, address models.Address) (models.Address, error) {
	updated, err := service.repository.UpdateAddress(ctx, supplierID, address)
	if err != nil {
		return models.Address{}, fmt.Errorf("update supplier %s address: %w", supplierID, err)
	}

	return updated, nil
}
