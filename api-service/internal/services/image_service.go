package services

import (
	"context"
	"errors"
	"fmt"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
)

var (
	ErrImageNotFound      = errors.New("image not found")
	ErrImageAlreadyExists = errors.New("product already has an image")
)

type ImageRepository interface {
	Create(context.Context, models.Image) (models.Image, error)
	GetByID(context.Context, uuid.UUID) (models.Image, error)
	GetByProductID(context.Context, uuid.UUID) (models.Image, error)
	Replace(context.Context, uuid.UUID, []byte) error
	Delete(context.Context, uuid.UUID) error
}

type ImageService struct {
	repository ImageRepository
}

func NewImageService(repository ImageRepository) *ImageService {
	return &ImageService{repository: repository}
}

func (service *ImageService) Create(ctx context.Context, productID uuid.UUID, data []byte) (models.ImageMetadata, error) {
	image := models.Image{ID: uuid.New(), ProductID: productID, Data: data}
	created, err := service.repository.Create(ctx, image)
	if err != nil {
		return models.ImageMetadata{}, fmt.Errorf("create product %s image: %w", productID, err)
	}

	return models.ImageMetadata{ID: created.ID, ProductID: created.ProductID}, nil
}

func (service *ImageService) GetByID(ctx context.Context, imageID uuid.UUID) (models.Image, error) {
	image, err := service.repository.GetByID(ctx, imageID)
	if err != nil {
		return models.Image{}, fmt.Errorf("get image %s: %w", imageID, err)
	}

	return image, nil
}

func (service *ImageService) GetByProductID(ctx context.Context, productID uuid.UUID) (models.Image, error) {
	image, err := service.repository.GetByProductID(ctx, productID)
	if err != nil {
		return models.Image{}, fmt.Errorf("get product %s image: %w", productID, err)
	}

	return image, nil
}

func (service *ImageService) Replace(ctx context.Context, imageID uuid.UUID, data []byte) error {
	if err := service.repository.Replace(ctx, imageID, data); err != nil {
		return fmt.Errorf("replace image %s: %w", imageID, err)
	}

	return nil
}

func (service *ImageService) Delete(ctx context.Context, imageID uuid.UUID) error {
	if err := service.repository.Delete(ctx, imageID); err != nil {
		return fmt.Errorf("delete image %s: %w", imageID, err)
	}

	return nil
}
