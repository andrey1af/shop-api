package models

import "github.com/google/uuid"

type Supplier struct {
	ID          uuid.UUID
	Name        string
	PhoneNumber string
	Address     Address
}

type SupplierCreate struct {
	Name        string        `json:"name"`
	PhoneNumber string        `json:"phone_number"`
	Address     AddressCreate `json:"address"`
}
