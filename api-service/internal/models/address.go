package models

import "github.com/google/uuid"

type Address struct {
	ID      uuid.UUID
	Country string
	City    string
	Street  string
}

type AddressCreate struct {
	Country string `json:"country"`
	City    string `json:"city"`
	Street  string `json:"street"`
}
