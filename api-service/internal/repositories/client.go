package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
	"github.com/andrey1af/shop-api/api-service/internal/services"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClientRepository struct {
	pool *pgxpool.Pool
}

var _ services.ClientRepository = (*ClientRepository)(nil)

func NewClientRepository(pool *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{pool: pool}
}

func (repository *ClientRepository) Create(ctx context.Context, client models.Client) (models.Client, error) {
	birthday, err := time.Parse(time.DateOnly, client.Birthday)
	if err != nil {
		return models.Client{}, fmt.Errorf("parse client birthday: %w", err)
	}
	registrationDate, err := time.Parse(time.DateOnly, client.RegistrationDate)
	if err != nil {
		return models.Client{}, fmt.Errorf("parse client registration date: %w", err)
	}

	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Client{}, fmt.Errorf("begin create client transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO addresses (id, country, city, street)
		VALUES ($1, $2, $3, $4)`,
		client.Address.ID,
		client.Address.Country,
		client.Address.City,
		client.Address.Street,
	)
	if err != nil {
		return models.Client{}, fmt.Errorf("insert client address: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO clients (
			id, client_name, client_surname, birthday, gender, registration_date, address_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		client.ID,
		client.ClientName,
		client.ClientSurname,
		birthday,
		client.Gender,
		registrationDate,
		client.Address.ID,
	)
	if err != nil {
		return models.Client{}, fmt.Errorf("insert client: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Client{}, fmt.Errorf("commit create client transaction: %w", err)
	}

	return client, nil
}

func (repository *ClientRepository) List(ctx context.Context, filter models.ClientFilter) ([]models.Client, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT c.id,
		       c.client_name,
		       c.client_surname,
		       c.birthday::text,
		       c.gender,
		       c.registration_date::text,
		       a.id,
		       a.country,
		       a.city,
		       a.street
		FROM clients AS c
		JOIN addresses AS a ON a.id = c.address_id
		WHERE ($1::text IS NULL OR (c.client_name = $1 AND c.client_surname = $2))
		ORDER BY c.id
		LIMIT $3::bigint OFFSET $4::bigint`,
		filter.ClientName,
		filter.ClientSurname,
		filter.Limit,
		filter.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query clients: %w", err)
	}
	defer rows.Close()

	clients := make([]models.Client, 0)
	for rows.Next() {
		client, err := scanClient(rows)
		if err != nil {
			return nil, fmt.Errorf("scan client: %w", err)
		}

		clients = append(clients, client)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate clients: %w", err)
	}

	return clients, nil
}

func (repository *ClientRepository) Delete(ctx context.Context, clientID uuid.UUID) error {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin delete client transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var addressID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT address_id
		FROM clients
		WHERE id = $1
		FOR UPDATE`, clientID).Scan(&addressID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("query client address for deletion: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM clients WHERE id = $1`, clientID); err != nil {
		return fmt.Errorf("delete client: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM addresses WHERE id = $1`, addressID); err != nil {
		return fmt.Errorf("delete client address: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete client transaction: %w", err)
	}

	return nil
}

func (repository *ClientRepository) GetAddress(ctx context.Context, clientID uuid.UUID) (models.Address, error) {
	var address models.Address
	err := repository.pool.QueryRow(ctx, `
		SELECT a.id, a.country, a.city, a.street
		FROM addresses AS a
		JOIN clients AS c ON c.address_id = a.id
		WHERE c.id = $1`, clientID).Scan(
		&address.ID,
		&address.Country,
		&address.City,
		&address.Street,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Address{}, services.ErrClientNotFound
	}
	if err != nil {
		return models.Address{}, fmt.Errorf("query client address: %w", err)
	}

	return address, nil
}

func (repository *ClientRepository) UpdateAddress(ctx context.Context, clientID uuid.UUID, address models.Address) (models.Address, error) {
	var updated models.Address
	err := repository.pool.QueryRow(ctx, `
		UPDATE addresses AS a
		SET country = $2,
		    city = $3,
		    street = $4
		FROM clients AS c
		WHERE c.id = $1
		  AND a.id = c.address_id
		RETURNING a.id, a.country, a.city, a.street`,
		clientID,
		address.Country,
		address.City,
		address.Street,
	).Scan(
		&updated.ID,
		&updated.Country,
		&updated.City,
		&updated.Street,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Address{}, services.ErrClientNotFound
	}
	if err != nil {
		return models.Address{}, fmt.Errorf("update client address: %w", err)
	}

	return updated, nil
}

func scanClient(row rowScanner) (models.Client, error) {
	var client models.Client
	err := row.Scan(
		&client.ID,
		&client.ClientName,
		&client.ClientSurname,
		&client.Birthday,
		&client.Gender,
		&client.RegistrationDate,
		&client.Address.ID,
		&client.Address.Country,
		&client.Address.City,
		&client.Address.Street,
	)

	return client, err
}
