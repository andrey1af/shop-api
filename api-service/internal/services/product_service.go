package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type ProductRepository interface {
	Create(context.Context, models.Product) (models.Product, error)
	ListAvailable(context.Context) ([]models.Product, error)
	GetByID(context.Context, uuid.UUID) (models.Product, error)
	Delete(context.Context, uuid.UUID) error
	DecreaseStock(context.Context, uuid.UUID, int64) (models.Product, error)
}

type ProductService struct {
	repository ProductRepository
}

func NewProductService(repository ProductRepository) *ProductService {
	return &ProductService{repository: repository}
}

func (service *ProductService) Create(ctx context.Context, candidate models.ProductCreate) (models.Product, error) {
	lastUpdateDate := candidate.LastUpdateDate
	if lastUpdateDate == "" {
		lastUpdateDate = time.Now().Format(time.DateOnly)
	}

	product := models.Product{
		ID:             uuid.New(),
		Name:           candidate.Name,
		Category:       candidate.Category,
		Price:          candidate.Price,
		AvailableStock: candidate.AvailableStock,
		LastUpdateDate: lastUpdateDate,
		SupplierID:     candidate.SupplierID,
		ImageID:        candidate.ImageID,
	}

	created, err := service.repository.Create(ctx, product)
	if err != nil {
		return models.Product{}, fmt.Errorf("create product: %w", err)
	}

	return created, nil
}

func (service *ProductService) GetAvailable(ctx context.Context) ([]models.Product, error) {
	products, err := service.repository.ListAvailable(ctx)
	if err != nil {
		return nil, fmt.Errorf("get available products: %w", err)
	}
	if products == nil {
		return []models.Product{}, nil
	}

	return products, nil
}

func (service *ProductService) GetByID(ctx context.Context, productID uuid.UUID) (models.Product, error) {
	product, err := service.repository.GetByID(ctx, productID)
	if err != nil {
		return models.Product{}, fmt.Errorf("get product %s: %w", productID, err)
	}

	return product, nil
}

func (service *ProductService) Delete(ctx context.Context, productID uuid.UUID) error {
	if err := service.repository.Delete(ctx, productID); err != nil {
		return fmt.Errorf("delete product %s: %w", productID, err)
	}

	return nil
}

func (service *ProductService) DecreaseStock(ctx context.Context, productID uuid.UUID, quantity int64) (models.Product, error) {
	product, err := service.repository.DecreaseStock(ctx, productID, quantity)
	if err != nil {
		return models.Product{}, fmt.Errorf("decrease product %s stock: %w", productID, err)
	}

	return product, nil
}
