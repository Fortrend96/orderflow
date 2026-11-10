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

// Product is a validated product value. Its fields cannot be changed by callers.
type Product struct {
	sku              SKU
	name             string
	unitPriceKopecks int64
}

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
		sku:              sku,
		name:             name,
		unitPriceKopecks: unitPriceKopecks,
	}, nil
}

func (p Product) SKU() SKU {
	return p.sku
}

func (p Product) Name() string {
	return p.name
}

func (p Product) UnitPriceKopecks() int64 {
	return p.unitPriceKopecks
}

// WithPrice returns a validated product copy with a different unit price.
func (p Product) WithPrice(
	newPriceKopecks int64,
) (Product, error) {
	return NewProduct(
		p.sku.String(),
		p.name,
		newPriceKopecks,
	)
}
