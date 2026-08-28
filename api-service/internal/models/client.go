package models

import "uuid"

type Client struct {
	ID               uuid.UUID `json:"id"`
	ClientName       string    `json:"client_name"`
	ClientSurname    string    `json:"client_surname"`
	Birthday         string    `json:"birthday"`
	Gender           string    `json:"gender"`
	RegistrationDate string    `json:"registration_date"`
	Address          Address   `json:"address"`
}

type ClientCreate struct {
	ClientName       string        `json:"client_name"`
	ClientSurname    string        `json:"client_surname"`
	Birthday         string        `json:"birthday"`
	Gender           string        `json:"gender"`
	RegistrationDate string        `json:"registration_date,omitempty"`
	Address          AddressCreate `json:"address"`
}

type ClientFilter struct {
	ClientName    *string
	ClientSurname *string
	Limit         *int
	Offset        *int
}
