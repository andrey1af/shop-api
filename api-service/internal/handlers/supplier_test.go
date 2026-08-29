package handlers

import (
	"testing"

	"github.com/andrey1af/shop-api/api-service/internal/models"
)

func TestValidSupplierCreate(t *testing.T) {
	valid := models.SupplierCreate{
		Name:        "Supplier",
		PhoneNumber: "+7 999 000-00-01",
		Address: models.AddressCreate{
			Country: "Russia",
			City:    "Moscow",
			Street:  "Tverskaya, 1",
		},
	}
	if !validSupplierCreate(valid) {
		t.Fatalf("validSupplierCreate() rejected %#v", valid)
	}

	tests := map[string]models.SupplierCreate{
		"missing name":  {PhoneNumber: valid.PhoneNumber, Address: valid.Address},
		"short phone":   {Name: valid.Name, PhoneNumber: "1234", Address: valid.Address},
		"long phone":    {Name: valid.Name, PhoneNumber: "1234567890123456789012345678901", Address: valid.Address},
		"blank country": {Name: valid.Name, PhoneNumber: valid.PhoneNumber, Address: models.AddressCreate{City: "Moscow", Street: "Street"}},
	}
	for name, candidate := range tests {
		t.Run(name, func(t *testing.T) {
			if validSupplierCreate(candidate) {
				t.Fatalf("validSupplierCreate() accepted %#v", candidate)
			}
		})
	}
}
