package models

import "uuid"

type Image struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	Data      []byte
}

type ImageMetadata struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
}
