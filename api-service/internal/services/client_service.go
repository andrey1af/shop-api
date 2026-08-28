package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
)

var ErrClientNotFound = errors.New("client not found")

type ClientRepository interface {
	Create(context.Context, models.Client) (models.Client, error)
	List(context.Context, models.ClientFilter) ([]models.Client, error)
	Delete(context.Context, uuid.UUID) error
	GetAddress(context.Context, uuid.UUID) (models.Address, error)
	UpdateAddress(context.Context, uuid.UUID, models.Address) (models.Address, error)
}

type ClientService struct {
	repository ClientRepository
}

func NewClientService(repository ClientRepository) *ClientService {
	return &ClientService{repository: repository}
}

func (service *ClientService) Create(ctx context.Context, candidate models.ClientCreate) (models.Client, error) {
	registrationDate := candidate.RegistrationDate
	if registrationDate == "" {
		registrationDate = time.Now().Format(time.DateOnly)
	}

	client := models.Client{
		ID:               uuid.New(),
		ClientName:       candidate.ClientName,
		ClientSurname:    candidate.ClientSurname,
		Birthday:         candidate.Birthday,
		Gender:           candidate.Gender,
		RegistrationDate: registrationDate,
		Address: models.Address{
			ID:      uuid.New(),
			Country: candidate.Address.Country,
			City:    candidate.Address.City,
			Street:  candidate.Address.Street,
		},
	}

	created, err := service.repository.Create(ctx, client)
	if err != nil {
		return models.Client{}, fmt.Errorf("create client: %w", err)
	}

	return created, nil
}

func (service *ClientService) GetAll(ctx context.Context, filter models.ClientFilter) ([]models.Client, error) {
	clients, err := service.repository.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get clients: %w", err)
	}

	if clients == nil {
		return []models.Client{}, nil
	}

	return clients, nil
}

func (service *ClientService) Delete(ctx context.Context, clientID uuid.UUID) error {
	if err := service.repository.Delete(ctx, clientID); err != nil {
		return fmt.Errorf("delete client %s: %w", clientID, err)
	}

	return nil
}

func (service *ClientService) GetAddress(ctx context.Context, clientID uuid.UUID) (models.Address, error) {
	address, err := service.repository.GetAddress(ctx, clientID)
	if err != nil {
		return models.Address{}, fmt.Errorf("get client %s address: %w", clientID, err)
	}

	return address, nil
}

func (service *ClientService) UpdateAddress(ctx context.Context, clientID uuid.UUID, address models.Address) (models.Address, error) {
	updated, err := service.repository.UpdateAddress(ctx, clientID, address)
	if err != nil {
		return models.Address{}, fmt.Errorf("update client %s address: %w", clientID, err)
	}

	return updated, nil
}
