package repositories

import (
	"context"
	"errors"
	"fmt"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
	"github.com/andrey1af/shop-api/api-service/internal/services"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SupplierRepository struct {
	pool *pgxpool.Pool
}

var _ services.SupplierRepository = (*SupplierRepository)(nil)

func NewSupplierRepository(pool *pgxpool.Pool) *SupplierRepository {
	return &SupplierRepository{pool: pool}
}

func (repository *SupplierRepository) Create(ctx context.Context, supplier models.Supplier) (models.Supplier, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Supplier{}, fmt.Errorf("begin create supplier transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO addresses (id, country, city, street)
		VALUES ($1, $2, $3, $4)`,
		supplier.Address.ID,
		supplier.Address.Country,
		supplier.Address.City,
		supplier.Address.Street,
	)
	if err != nil {
		return models.Supplier{}, fmt.Errorf("insert supplier address: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO suppliers (id, name, phone_number, address_id)
		VALUES ($1, $2, $3, $4)`,
		supplier.ID,
		supplier.Name,
		supplier.PhoneNumber,
		supplier.Address.ID,
	)
	if err != nil {
		return models.Supplier{}, fmt.Errorf("insert supplier: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Supplier{}, fmt.Errorf("commit create supplier transaction: %w", err)
	}

	return supplier, nil
}

func (repository *SupplierRepository) List(ctx context.Context) ([]models.Supplier, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT s.id,
		       s.name,
		       s.phone_number,
		       a.id,
		       a.country,
		       a.city,
		       a.street
		FROM suppliers AS s
		JOIN addresses AS a ON a.id = s.address_id
		ORDER BY s.id`)
	if err != nil {
		return nil, fmt.Errorf("query suppliers: %w", err)
	}
	defer rows.Close()

	suppliers := make([]models.Supplier, 0)
	for rows.Next() {
		supplier, err := scanSupplier(rows)
		if err != nil {
			return nil, fmt.Errorf("scan supplier: %w", err)
		}

		suppliers = append(suppliers, supplier)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate suppliers: %w", err)
	}

	return suppliers, nil
}

func (repository *SupplierRepository) GetByID(ctx context.Context, supplierID uuid.UUID) (models.Supplier, error) {
	supplier, err := scanSupplier(repository.pool.QueryRow(ctx, `
		SELECT s.id,
		       s.name,
		       s.phone_number,
		       a.id,
		       a.country,
		       a.city,
		       a.street
		FROM suppliers AS s
		JOIN addresses AS a ON a.id = s.address_id
		WHERE s.id = $1`, supplierID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Supplier{}, services.ErrSupplierNotFound
	}
	if err != nil {
		return models.Supplier{}, fmt.Errorf("query supplier by ID: %w", err)
	}

	return supplier, nil
}

func (repository *SupplierRepository) Delete(ctx context.Context, supplierID uuid.UUID) error {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin delete supplier transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var addressID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT address_id
		FROM suppliers
		WHERE id = $1
		FOR UPDATE`, supplierID).Scan(&addressID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("query supplier address for deletion: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM suppliers WHERE id = $1`, supplierID); err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("%w: %v", services.ErrSupplierInUse, err)
		}

		return fmt.Errorf("delete supplier: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM addresses WHERE id = $1`, addressID); err != nil {
		return fmt.Errorf("delete supplier address: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete supplier transaction: %w", err)
	}

	return nil
}

func (repository *SupplierRepository) GetAddress(ctx context.Context, supplierID uuid.UUID) (models.Address, error) {
	var address models.Address
	err := repository.pool.QueryRow(ctx, `
		SELECT a.id, a.country, a.city, a.street
		FROM addresses AS a
		JOIN suppliers AS s ON s.address_id = a.id
		WHERE s.id = $1`, supplierID).Scan(
		&address.ID,
		&address.Country,
		&address.City,
		&address.Street,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Address{}, services.ErrSupplierNotFound
	}
	if err != nil {
		return models.Address{}, fmt.Errorf("query supplier address: %w", err)
	}

	return address, nil
}

func (repository *SupplierRepository) UpdateAddress(ctx context.Context, supplierID uuid.UUID, address models.Address) (models.Address, error) {
	var updated models.Address
	err := repository.pool.QueryRow(ctx, `
		UPDATE addresses AS a
		SET country = $2,
		    city = $3,
		    street = $4
		FROM suppliers AS s
		WHERE s.id = $1
		  AND a.id = s.address_id
		RETURNING a.id, a.country, a.city, a.street`,
		supplierID,
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
		return models.Address{}, services.ErrSupplierNotFound
	}
	if err != nil {
		return models.Address{}, fmt.Errorf("update supplier address: %w", err)
	}

	return updated, nil
}

type rowScanner interface {
	Scan(...any) error
}

func scanSupplier(row rowScanner) (models.Supplier, error) {
	var supplier models.Supplier
	err := row.Scan(
		&supplier.ID,
		&supplier.Name,
		&supplier.PhoneNumber,
		&supplier.Address.ID,
		&supplier.Address.Country,
		&supplier.Address.City,
		&supplier.Address.Street,
	)

	return supplier, err
}

func isForeignKeyViolation(err error) bool {
	var postgresError *pgconn.PgError

	return errors.As(err, &postgresError) && postgresError.Code == "23503"
}
