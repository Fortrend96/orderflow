package catalog

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	skuLength    = 8
	minNameRunes = 2
	maxNameRunes = 80
)

type SKU [skuLength]byte

type Product struct {
	SKU              SKU
	Name             string
	UnitPriceKopecks int64
}

// ParseSKU accepts exactly eight uppercase ASCII letters or digits.
func ParseSKU(raw string) (SKU, error) {
	var sku SKU

	if len(raw) != skuLength {
		return sku, fmt.Errorf(
			"SKU must contain exactly %d ASCII bytes",
			skuLength,
		)
	}

	for i := 0; i < len(raw); i++ {
		b := raw[i]

		letter := b >= 'A' && b <= 'Z'
		digit := b >= '0' && b <= '9'

		if !letter && !digit {
			return SKU{}, fmt.Errorf(
				"SKU contains invalid byte at offset %d",
				i,
			)
		}

		sku[i] = b
	}

	return sku, nil
}

func (sku SKU) String() string {
	return string(sku[:])
}

// NewProduct validates a product and normalizes surrounding whitespace.
func NewProduct(
	rawSKU string,
	rawName string,
	unitPriceKopecks int64,
) (Product, error) {
	sku, err := ParseSKU(rawSKU)

	if err != nil {
		return Product{}, fmt.Errorf(
			"invalid product SKU: %w",
			err,
		)
	}

	if !utf8.ValidString(rawName) {
		return Product{}, fmt.Errorf(
			"product name is not valid UTF-8",
		)
	}

	name := strings.TrimSpace(rawName)

	nameRunes := utf8.RuneCountInString(name)

	if nameRunes < minNameRunes ||
		nameRunes > maxNameRunes {
		return Product{}, fmt.Errorf(
			"product name must contain %d..%d Unicode code points",
			minNameRunes,
			maxNameRunes,
		)
	}

	if unitPriceKopecks < 0 {
		return Product{}, fmt.Errorf(
			"unit price must be non-negative",
		)
	}

	return Product{
		SKU:              sku,
		Name:             name,
		UnitPriceKopecks: unitPriceKopecks,
	}, nil
}
