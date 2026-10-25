package catalog

import (
	"errors"
	"fmt"
)

var ErrProductExists = errors.New("product already exists")

type Catalog struct {
	products []Product
}

func NewCatalog(initialCapacity int) (*Catalog, error) {
	if initialCapacity < 0 {
		return nil, fmt.Errorf(
			"initial capacity must be non-negative: %d",
			initialCapacity,
		)
	}

	return &Catalog{
		products: make([]Product, 0, initialCapacity),
	}, nil
}

func (c *Catalog) Add(product Product) error {
	validated, err := NewProduct(
		product.SKU.String(),
		product.Name,
		product.UnitPriceKopecks,
	)

	if err != nil {
		return fmt.Errorf("invalid product: %w", err)
	}

	for _, current := range c.products {
		if current.SKU == validated.SKU {
			return fmt.Errorf(
				"%w: %s",
				ErrProductExists,
				validated.SKU.String(),
			)
		}
	}

	c.products = append(c.products, validated)

	return nil
}

func (c *Catalog) Find(sku SKU) (Product, bool) {
	for _, product := range c.products {
		if product.SKU == sku {
			return product, true
		}
	}

	return Product{}, false
}

// List returns a separate slice so callers cannot modify internal storage.
func (c *Catalog) List() []Product {
	result := make([]Product, len(c.products))

	copy(result, c.products)

	return result
}

func (c *Catalog) Remove(sku SKU) bool {
	for i, product := range c.products {
		if product.SKU == sku {
			copy(c.products[i:], c.products[i+1:])

			last := len(c.products) - 1

			c.products[last] = Product{}
			c.products = c.products[:last]

			return true
		}
	}

	return false
}

func (c *Catalog) Count() int {
	return len(c.products)
}
