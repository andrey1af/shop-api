package services

import (
	"context"
	"errors"
	"testing"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
)

type imageRepositoryStub struct {
	created     models.Image
	createInput models.Image
	replaceErr  error
}

func (stub *imageRepositoryStub) Create(_ context.Context, image models.Image) (models.Image, error) {
	stub.createInput = image
	if stub.created.ID != uuid.Nil() {
		return stub.created, nil
	}
	return image, nil
}

func (stub *imageRepositoryStub) GetByID(context.Context, uuid.UUID) (models.Image, error) {
	return models.Image{}, nil
}

func (stub *imageRepositoryStub) GetByProductID(context.Context, uuid.UUID) (models.Image, error) {
	return models.Image{}, nil
}

func (stub *imageRepositoryStub) Replace(context.Context, uuid.UUID, []byte) error {
	return stub.replaceErr
}

func (stub *imageRepositoryStub) Delete(context.Context, uuid.UUID) error {
	return nil
}

func TestCreateImageAssignsIDAndReturnsMetadata(t *testing.T) {
	repository := &imageRepositoryStub{}
	service := NewImageService(repository)
	productID := uuid.New()
	data := []byte("image data")

	metadata, err := service.Create(context.Background(), productID, data)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if metadata.ID == uuid.Nil() || metadata.ProductID != productID {
		t.Fatalf("Create() metadata = %#v", metadata)
	}
	if repository.createInput.ID != metadata.ID || repository.createInput.ProductID != productID || string(repository.createInput.Data) != string(data) {
		t.Fatalf("repository image = %#v", repository.createInput)
	}
}

func TestImageNotFoundCanBeInspectedThroughService(t *testing.T) {
	service := NewImageService(&imageRepositoryStub{replaceErr: ErrImageNotFound})

	if err := service.Replace(context.Background(), uuid.New(), []byte("new image")); !errors.Is(err, ErrImageNotFound) {
		t.Fatalf("Replace() error = %v, want ErrImageNotFound", err)
	}
}
