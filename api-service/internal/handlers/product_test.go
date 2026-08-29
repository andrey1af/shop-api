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
		Price:          testPointer(54990.0),
		AvailableStock: testPointer(int64(12)),
		LastUpdateDate: "2026-08-28",
		SupplierID:     uuid.New(),
	}
	if !validProductCreate(candidate) {
		t.Fatalf("validProductCreate() rejected %#v", candidate)
	}
}

func TestValidProductCreateRejectsInvalidValues(t *testing.T) {
	valid := models.ProductCreate{
		Name:           "Product",
		Category:       "Category",
		Price:          testPointer(1.0),
		AvailableStock: testPointer(int64(0)),
		SupplierID:     uuid.New(),
	}

	tests := map[string]models.ProductCreate{
		"blank name":       {Category: valid.Category, Price: valid.Price, AvailableStock: valid.AvailableStock, SupplierID: valid.SupplierID},
		"missing price":    {Name: valid.Name, Category: valid.Category, AvailableStock: valid.AvailableStock, SupplierID: valid.SupplierID},
		"missing stock":    {Name: valid.Name, Category: valid.Category, Price: valid.Price, SupplierID: valid.SupplierID},
		"negative price":   {Name: valid.Name, Category: valid.Category, Price: testPointer(-1.0), AvailableStock: valid.AvailableStock, SupplierID: valid.SupplierID},
		"negative stock":   {Name: valid.Name, Category: valid.Category, Price: valid.Price, AvailableStock: testPointer(int64(-1)), SupplierID: valid.SupplierID},
		"missing supplier": {Name: valid.Name, Category: valid.Category, Price: valid.Price, AvailableStock: valid.AvailableStock},
		"invalid date":     {Name: valid.Name, Category: valid.Category, Price: valid.Price, AvailableStock: valid.AvailableStock, SupplierID: valid.SupplierID, LastUpdateDate: "28-08-2026"},
	}

	for name, candidate := range tests {
		t.Run(name, func(t *testing.T) {
			if validProductCreate(candidate) {
				t.Fatalf("validProductCreate() accepted %#v", candidate)
			}
		})
	}
}

func testPointer[T any](value T) *T {
	return &value
}
