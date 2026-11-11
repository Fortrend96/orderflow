package catalog

import (
	"errors"
	"testing"
)

func TestProductValidationErrorKinds(t *testing.T) {
	tests := []struct {
		name  string
		sku   string
		text  string
		price int64
		want  error
	}{
		{
			name:  "short SKU",
			sku:   "SHORT",
			text:  "Кофе",
			price: 100,
			want:  ErrInvalidSKU,
		},
		{
			name:  "bad SKU character",
			sku:   "coffee01",
			text:  "Кофе",
			price: 100,
			want:  ErrInvalidSKU,
		},
		{
			name:  "invalid UTF-8",
			sku:   "COFFEE01",
			text:  "x\xff",
			price: 100,
			want:  ErrInvalidName,
		},
		{
			name:  "short name",
			sku:   "COFFEE01",
			text:  "Я",
			price: 100,
			want:  ErrInvalidName,
		},
		{
			name:  "negative price",
			sku:   "COFFEE01",
			text:  "Кофе",
			price: -10,
			want:  ErrInvalidPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProduct(
				tt.sku,
				tt.text,
				tt.price,
			)

			if !errors.Is(err, tt.want) {
				t.Fatalf(
					"NewProduct returned %v; errors.Is(..., %v) = false",
					err,
					tt.want,
				)
			}
		})
	}
}

func TestCatalogAddAndUpdateWrapValidationErrors(
	t *testing.T,
) {
	store, err := NewCatalog(1)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Add(Product{}); !errors.Is(
		err,
		ErrInvalidSKU,
	) {
		t.Fatalf(
			"Add zero Product: %v; want ErrInvalidSKU",
			err,
		)
	}

	product, err := NewProduct(
		"COFFEE01",
		"Кофе",
		12990,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Add(product); err != nil {
		t.Fatal(err)
	}

	if err := store.Add(product); !errors.Is(
		err,
		ErrProductExists,
	) {
		t.Fatalf(
			"Add duplicate: %v; want ErrProductExists",
			err,
		)
	}

	if err := store.UpdatePrice(
		product.SKU(),
		-1,
	); !errors.Is(err, ErrInvalidPrice) {
		t.Fatalf(
			"UpdatePrice(-1): %v; want ErrInvalidPrice",
			err,
		)
	}

	stored, found := store.Find(product.SKU())
	if !found || stored.UnitPriceKopecks() != 12990 {
		t.Fatal("failed update changed the stored price")
	}

	missing, err := ParseSKU("MISSING1")
	if err != nil {
		t.Fatal(err)
	}

	if err := store.UpdatePrice(
		missing,
		100,
	); !errors.Is(err, ErrProductNotFound) {
		t.Fatalf(
			"UpdatePrice missing: %v; want ErrProductNotFound",
			err,
		)
	}
}
