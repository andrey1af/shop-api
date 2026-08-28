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

type ImageRepository struct {
	pool *pgxpool.Pool
}

var _ services.ImageRepository = (*ImageRepository)(nil)

func NewImageRepository(pool *pgxpool.Pool) *ImageRepository {
	return &ImageRepository{pool: pool}
}

func (repository *ImageRepository) Create(ctx context.Context, image models.Image) (models.Image, error) {
	_, err := repository.pool.Exec(ctx, `
		INSERT INTO images (id, product_id, image)
		VALUES ($1, $2, $3)`, image.ID, image.ProductID, image.Data)
	if isUniqueViolation(err) {
		return models.Image{}, services.ErrImageAlreadyExists
	}
	if isForeignKeyViolation(err) {
		return models.Image{}, services.ErrProductNotFound
	}
	if err != nil {
		return models.Image{}, fmt.Errorf("insert image: %w", err)
	}

	return image, nil
}

func (repository *ImageRepository) GetByID(ctx context.Context, imageID uuid.UUID) (models.Image, error) {
	return repository.get(ctx, `
		SELECT id, product_id, image
		FROM images
		WHERE id = $1`, imageID)
}

func (repository *ImageRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (models.Image, error) {
	return repository.get(ctx, `
		SELECT id, product_id, image
		FROM images
		WHERE product_id = $1`, productID)
}

func (repository *ImageRepository) Replace(ctx context.Context, imageID uuid.UUID, data []byte) error {
	result, err := repository.pool.Exec(ctx, `UPDATE images SET image = $2 WHERE id = $1`, imageID, data)
	if err != nil {
		return fmt.Errorf("update image: %w", err)
	}
	if result.RowsAffected() == 0 {
		return services.ErrImageNotFound
	}

	return nil
}

func (repository *ImageRepository) Delete(ctx context.Context, imageID uuid.UUID) error {
	if _, err := repository.pool.Exec(ctx, `DELETE FROM images WHERE id = $1`, imageID); err != nil {
		return fmt.Errorf("delete image: %w", err)
	}

	return nil
}

func (repository *ImageRepository) get(ctx context.Context, query string, id uuid.UUID) (models.Image, error) {
	var image models.Image
	err := repository.pool.QueryRow(ctx, query, id).Scan(&image.ID, &image.ProductID, &image.Data)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Image{}, services.ErrImageNotFound
	}
	if err != nil {
		return models.Image{}, fmt.Errorf("query image: %w", err)
	}

	return image, nil
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError

	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
