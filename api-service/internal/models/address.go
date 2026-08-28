package models

import "uuid"

type Address struct {
	ID      uuid.UUID `json:"id"`
	Country string    `json:"country"`
	City    string    `json:"city"`
	Street  string    `json:"street"`
}

type AddressCreate struct {
	Country string `json:"country"`
	City    string `json:"city"`
	Street  string `json:"street"`
}
