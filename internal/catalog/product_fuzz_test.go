package catalog

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzParseSKU checks the public parsing contract on arbitrary input bytes.
func FuzzParseSKU(f *testing.F) {
	for _, seed := range []string{
		"COFFEE01", "TEA00001", "", "short", "coffee01",
		"ABCDEF_1", "ABCDEFGH", "12345678", "Кофе123", "AB\xffCD123",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 1024 {
			return // Keep test executions bounded, including generated inputs.
		}
		sku, err := ParseSKU(raw)
		if err != nil {
			if !errors.Is(err, ErrInvalidSKU) {
				t.Fatalf("ParseSKU(%q): unexpected error category: %v", raw, err)
			}
			if sku != (SKU{}) {
				t.Fatalf("ParseSKU(%q) returned a nonzero SKU on error: %v", raw, sku)
			}
			return
		}

		if len(raw) != 8 || sku.String() != raw {
			t.Fatalf("successful ParseSKU did not round-trip %q: %q", raw, sku.String())
		}
		for _, b := range []byte(raw) {
			if !((b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')) {
				t.Fatalf("successful ParseSKU accepted invalid byte %q in %q", b, raw)
			}
		}
		parsedAgain, err := ParseSKU(sku.String())
		if err != nil || parsedAgain != sku {
			t.Fatalf("round-trip failed: %v, %v", parsedAgain, err)
		}
	})
}

// FuzzNewProduct checks successful values and the category of rejected input.
func FuzzNewProduct(f *testing.F) {
	f.Add("COFFEE01", "Кофе", int64(12990))
	f.Add("COFFEE01", " Кофе ", int64(0))
	f.Add("COFFEE01", "\xff", int64(100))
	f.Add("coffee01", "Coffee", int64(100))
	f.Add("TEA00001", "Чай", int64(-1))
	f.Add("TEA00001", "", int64(5))

	f.Fuzz(func(t *testing.T, rawSKU, rawName string, price int64) {
		if len(rawSKU) > 1024 || len(rawName) > 1024 {
			return
		}
		product, err := NewProduct(rawSKU, rawName, price)
		if err != nil {
			if product != (Product{}) {
				t.Fatalf("NewProduct returned nonzero product on error: %+v", product)
			}
			if !errors.Is(err, ErrInvalidSKU) && !errors.Is(err, ErrInvalidName) && !errors.Is(err, ErrInvalidPrice) {
				t.Fatalf("unexpected error category: %v", err)
			}
			return
		}

		if product.SKU().String() != rawSKU {
			t.Fatalf("SKU changed: %q != %q", product.SKU(), rawSKU)
		}
		if product.Name() != strings.TrimSpace(rawName) {
			t.Fatalf("name not trimmed: %q", product.Name())
		}
		if !utf8.ValidString(product.Name()) || utf8.RuneCountInString(product.Name()) < 2 ||
			utf8.RuneCountInString(product.Name()) > 80 {
			t.Fatalf("invalid successful product name: %q", product.Name())
		}
		if product.UnitPriceKopecks() != price || price < 0 {
			t.Fatalf("invalid successful price: %d", product.UnitPriceKopecks())
		}
	})
}
