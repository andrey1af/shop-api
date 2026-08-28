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

type ProductRepository struct {
	pool *pgxpool.Pool
}

var _ services.ProductRepository = (*ProductRepository)(nil)

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (repository *ProductRepository) Create(ctx context.Context, product models.Product) (models.Product, error) {
	lastUpdateDate, err := time.Parse(time.DateOnly, product.LastUpdateDate)
	if err != nil {
		return models.Product{}, fmt.Errorf("parse product last update date: %w", err)
	}

	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Product{}, fmt.Errorf("begin create product transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var supplierExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM suppliers WHERE id = $1)`, product.SupplierID).Scan(&supplierExists); err != nil {
		return models.Product{}, fmt.Errorf("check product supplier: %w", err)
	}
	if !supplierExists {
		return models.Product{}, services.ErrSupplierNotFound
	}

	var categoryID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO product_categories (id, name)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, uuid.New(), product.Category).Scan(&categoryID)
	if err != nil {
		return models.Product{}, fmt.Errorf("upsert product category: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO products (
			id, name, category_id, price, available_stock, last_update_date, supplier_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		product.ID,
		product.Name,
		categoryID,
		product.Price,
		product.AvailableStock,
		lastUpdateDate,
		product.SupplierID,
	)
	if isForeignKeyViolation(err) {
		return models.Product{}, services.ErrSupplierNotFound
	}
	if err != nil {
		return models.Product{}, fmt.Errorf("insert product: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Product{}, fmt.Errorf("commit create product transaction: %w", err)
	}

	product.ImageID = nil
	return product, nil
}

func (repository *ProductRepository) ListAvailable(ctx context.Context) ([]models.Product, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT p.id,
		       p.name,
		       pc.name,
		       p.price,
		       p.available_stock,
		       p.last_update_date::text,
		       p.supplier_id,
		       i.id
		FROM products AS p
		JOIN product_categories AS pc ON pc.id = p.category_id
		LEFT JOIN images AS i ON i.product_id = p.id
		WHERE p.available_stock > 0
		ORDER BY p.id`)
	if err != nil {
		return nil, fmt.Errorf("query available products: %w", err)
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scan available product: %w", err)
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate available products: %w", err)
	}

	return products, nil
}

func (repository *ProductRepository) GetByID(ctx context.Context, productID uuid.UUID) (models.Product, error) {
	product, err := scanProduct(repository.pool.QueryRow(ctx, `
		SELECT p.id,
		       p.name,
		       pc.name,
		       p.price,
		       p.available_stock,
		       p.last_update_date::text,
		       p.supplier_id,
		       i.id
		FROM products AS p
		JOIN product_categories AS pc ON pc.id = p.category_id
		LEFT JOIN images AS i ON i.product_id = p.id
		WHERE p.id = $1`, productID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Product{}, services.ErrProductNotFound
	}
	if err != nil {
		return models.Product{}, fmt.Errorf("query product by ID: %w", err)
	}

	return product, nil
}

func (repository *ProductRepository) Delete(ctx context.Context, productID uuid.UUID) error {
	if _, err := repository.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, productID); err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	return nil
}

func (repository *ProductRepository) DecreaseStock(ctx context.Context, productID uuid.UUID, quantity int64) (models.Product, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Product{}, fmt.Errorf("begin decrease stock transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var availableStock int64
	err = tx.QueryRow(ctx, `
		SELECT available_stock
		FROM products
		WHERE id = $1
		FOR UPDATE`, productID).Scan(&availableStock)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Product{}, services.ErrProductNotFound
	}
	if err != nil {
		return models.Product{}, fmt.Errorf("lock product stock: %w", err)
	}
	if availableStock < quantity {
		return models.Product{}, services.ErrInsufficientStock
	}

	product, err := scanProduct(tx.QueryRow(ctx, `
		UPDATE products AS p
		SET available_stock = p.available_stock - $2,
		    last_update_date = CURRENT_DATE
		FROM product_categories AS pc
		WHERE p.id = $1
		  AND pc.id = p.category_id
		RETURNING p.id,
		          p.name,
		          pc.name,
		          p.price,
		          p.available_stock,
		          p.last_update_date::text,
		          p.supplier_id,
		          (SELECT i.id FROM images AS i WHERE i.product_id = p.id)`,
		productID,
		quantity,
	))
	if err != nil {
		return models.Product{}, fmt.Errorf("update product stock: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Product{}, fmt.Errorf("commit decrease stock transaction: %w", err)
	}

	return product, nil
}

func scanProduct(row rowScanner) (models.Product, error) {
	var product models.Product
	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Category,
		&product.Price,
		&product.AvailableStock,
		&product.LastUpdateDate,
		&product.SupplierID,
		&product.ImageID,
	)

	return product, err
}
