package models

import "uuid"

type Product struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Category       string     `json:"category"`
	Price          float64    `json:"price"`
	AvailableStock int64      `json:"available_stock"`
	LastUpdateDate string     `json:"last_update_date"`
	SupplierID     uuid.UUID  `json:"supplier_id"`
	ImageID        *uuid.UUID `json:"image_id"`
}

type ProductCreate struct {
	Name           string     `json:"name"`
	Category       string     `json:"category"`
	Price          float64    `json:"price"`
	AvailableStock int64      `json:"available_stock"`
	LastUpdateDate string     `json:"last_update_date,omitempty"`
	SupplierID     uuid.UUID  `json:"supplier_id"`
	ImageID        *uuid.UUID `json:"image_id,omitempty"`
}

type StockDecrease struct {
	Quantity int64 `json:"quantity"`
}
