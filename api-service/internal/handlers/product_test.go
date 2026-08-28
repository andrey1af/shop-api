package handlers

import (
	"testing"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
)

func TestValidProductCreate(t *testing.T) {
	candidate := models.ProductCreate{
		Name:           "Refrigerator",
		Category:       "Appliances",
		Price:          54990,
		AvailableStock: 12,
		LastUpdateDate: "2026-08-28",
		SupplierID:     uuid.New(),
	}
	if !validProductCreate(candidate) {
		t.Fatalf("validProductCreate() rejected %#v", candidate)
	}
}

func TestValidProductCreateRejectsInvalidValues(t *testing.T) {
	valid := models.ProductCreate{Name: "Product", Category: "Category", Price: 1, SupplierID: uuid.New()}
	imageID := uuid.New()

	tests := map[string]models.ProductCreate{
		"blank name":       {Category: valid.Category, Price: valid.Price, SupplierID: valid.SupplierID},
		"negative price":   {Name: valid.Name, Category: valid.Category, Price: -1, SupplierID: valid.SupplierID},
		"negative stock":   {Name: valid.Name, Category: valid.Category, Price: valid.Price, AvailableStock: -1, SupplierID: valid.SupplierID},
		"missing supplier": {Name: valid.Name, Category: valid.Category, Price: valid.Price},
		"invalid date":     {Name: valid.Name, Category: valid.Category, Price: valid.Price, SupplierID: valid.SupplierID, LastUpdateDate: "28-08-2026"},
		"prelinked image":  {Name: valid.Name, Category: valid.Category, Price: valid.Price, SupplierID: valid.SupplierID, ImageID: &imageID},
	}

	for name, candidate := range tests {
		t.Run(name, func(t *testing.T) {
			if validProductCreate(candidate) {
				t.Fatalf("validProductCreate() accepted %#v", candidate)
			}
		})
	}
}
