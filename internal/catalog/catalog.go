package catalog

import (
	"errors"
	"fmt"
)

var (
	ErrProductExists   = errors.New("product already exists")
	ErrProductNotFound = errors.New("product not found")
)

// Catalog stores products in insertion order and indexes them by SKU.
// Catalog is not safe for concurrent access.
type Catalog struct {
	products   []Product
	indexBySKU map[SKU]int
}

func NewCatalog(initialCapacity int) (*Catalog, error) {
	if initialCapacity < 0 {
		return nil, fmt.Errorf(
			"initial capacity must be non-negative: %d",
			initialCapacity,
		)
	}

	return &Catalog{
		products:   make([]Product, 0, initialCapacity),
		indexBySKU: make(map[SKU]int, initialCapacity),
	}, nil
}

func (c *Catalog) Add(product Product) error {
	validated, err := NewProduct(
		product.SKU.String(),
		product.Name,
		product.UnitPriceKopecks,
	)

	if err != nil {
		return fmt.Errorf(
			"invalid product: %w",
			err,
		)
	}

	if _, exists := c.indexBySKU[validated.SKU]; exists {
		return fmt.Errorf(
			"%w: %s",
			ErrProductExists,
			validated.SKU.String(),
		)
	}

	if c.indexBySKU == nil {
		c.indexBySKU = make(map[SKU]int)
	}

	position := len(c.products)

	c.products = append(c.products, validated)
	c.indexBySKU[validated.SKU] = position

	return nil
}

func (c *Catalog) Find(sku SKU) (Product, bool) {
	position, found := c.indexBySKU[sku]

	if !found {
		return Product{}, false
	}

	return c.products[position], true
}

func (c *Catalog) List() []Product {
	result := make([]Product, len(c.products))

	copy(result, c.products)

	return result
}

func (c *Catalog) Remove(sku SKU) bool {
	position, found := c.indexBySKU[sku]

	if !found {
		return false
	}

	copy(
		c.products[position:],
		c.products[position+1:],
	)

	last := len(c.products) - 1

	c.products[last] = Product{}
	c.products = c.products[:last]

	delete(c.indexBySKU, sku)

	for i := position; i < len(c.products); i++ {
		c.indexBySKU[c.products[i].SKU] = i
	}

	return true

}

func (c *Catalog) UpdatePrice(
	sku SKU,
	newPriceKopecks int64,
) error {
	if newPriceKopecks < 0 {
		return fmt.Errorf(
			"unit price must be non-negative: %d",
			newPriceKopecks,
		)
	}

	position, found := c.indexBySKU[sku]

	if !found {
		return fmt.Errorf(
			"%w: %s",
			ErrProductNotFound,
			sku.String(),
		)
	}

	c.products[position].UnitPriceKopecks = newPriceKopecks

	return nil
}

func (c *Catalog) Count() int {
	return len(c.products)
}
