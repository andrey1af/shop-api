package models

import "uuid"

type Supplier struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phone_number"`
	Address     Address   `json:"address"`
}

type SupplierCreate struct {
	Name        string        `json:"name"`
	PhoneNumber string        `json:"phone_number"`
	Address     AddressCreate `json:"address"`
}
