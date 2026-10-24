package catalog

import (
	"strings"
	"testing"
)

func TestParseSKU(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantError bool
	}{
		{
			name:  "letters and digits",
			input: "COFFEE01",
			want:  "COFFEE01",
		},
		{
			name:  "all digits",
			input: "12345678",
			want:  "12345678",
		},
		{
			name:      "too short",
			input:     "ABC1234",
			wantError: true,
		},
		{
			name:      "too long",
			input:     "ABC123456",
			wantError: true,
		},
		{
			name:      "lowercase",
			input:     "coffee01",
			wantError: true,
		},
		{
			name:      "hyphen",
			input:     "COFF-001",
			wantError: true,
		},
		{
			name:      "non ASCII",
			input:     "КОФЕ1234",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSKU(tt.input)

			if (err != nil) != tt.wantError {
				t.Fatalf(
					"ParseSKU(%q) error=%v, wantError=%t",
					tt.input,
					err,
					tt.wantError,
				)
			}

			if err != nil {
				return
			}

			if got.String() != tt.want {
				t.Errorf(
					"ParseSKU(%q)=%q, want %q",
					tt.input,
					got.String(),
					tt.want,
				)
			}
		})
	}
}

func TestNewProduct(t *testing.T) {
	tests := []struct {
		name        string
		sku         string
		productName string
		price       int64
		wantName    string
		wantError   bool
	}{
		{
			name:        "Cyrillic name",
			sku:         "COFFEE01",
			productName: "Кофе",
			price:       12990,
			wantName:    "Кофе",
		},
		{
			name:        "trim whitespace",
			sku:         "COFFEE01",
			productName: "  Чай  ",
			price:       0,
			wantName:    "Чай",
		},
		{
			name:        "empty name",
			sku:         "COFFEE01",
			productName: "  ",
			price:       100,
			wantError:   true,
		},
		{
			name:        "short name",
			sku:         "COFFEE01",
			productName: "Я",
			price:       100,
			wantError:   true,
		},
		{
			name:        "80 runes",
			sku:         "COFFEE01",
			productName: strings.Repeat("Я", 80),
			price:       100,
			wantName:    strings.Repeat("Я", 80),
		},
		{
			name:        "81 runes",
			sku:         "COFFEE01",
			productName: strings.Repeat("Я", 81),
			price:       100,
			wantError:   true,
		},
		{
			name:        "invalid UTF8",
			sku:         "COFFEE01",
			productName: "bad\xff",
			price:       100,
			wantError:   true,
		},
		{
			name:        "invalid SKU",
			sku:         "coffee01",
			productName: "Кофе",
			price:       100,
			wantError:   true,
		},
		{
			name:        "negative price",
			sku:         "COFFEE01",
			productName: "Кофе",
			price:       -1,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := NewProduct(
				tt.sku,
				tt.productName,
				tt.price,
			)

			if (err != nil) != tt.wantError {
				t.Fatalf(
					"NewProduct(%q, %q, %d) error=%v, wantError=%t",
					tt.sku,
					tt.productName,
					tt.price,
					err,
					tt.wantError,
				)
			}

			if err != nil {
				return
			}

			if product.Name != tt.wantName {
				t.Errorf(
					"got name %q, want %q",
					product.Name,
					tt.wantName,
				)
			}

			if product.SKU.String() != tt.sku {
				t.Errorf(
					"got SKU %q, want %q",
					product.SKU.String(),
					tt.sku,
				)
			}

			if product.UnitPriceKopecks != tt.price {
				t.Errorf(
					"got price %d, want %d",
					product.UnitPriceKopecks,
					tt.price,
				)
			}
		})
	}
}

func TestSKUArrayHasValueSemantics(t *testing.T) {
	original, err := ParseSKU("COFFEE01")

	if err != nil {
		t.Fatal(err)
	}

	copied := original
	copied[0] = 'T'

	if original.String() != "COFFEE01" {
		t.Errorf(
			"original unexpectedly changed: %q",
			original.String(),
		)
	}

	if copied.String() != "TOFFEE01" {
		t.Errorf(
			"copy is %q, want TOFFEE01",
			copied.String(),
		)
	}
}
