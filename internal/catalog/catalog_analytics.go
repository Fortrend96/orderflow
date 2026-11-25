package catalog

import "github.com/Fortrend96/orderflow/internal/collections"

// PriceGroups returns independent groups of products by price in kopecks.
// Products inside each group retain insertion order. Map key order is undefined.
func (c *Catalog) PriceGroups() map[int64][]Product {
	return collections.GroupBy(
		c.products,
		Product.UnitPriceKopecks,
	)
}

// DistinctNames returns product names in first-appearance order.
// It avoids allocating an intermediate []Product slice.
func (c *Catalog) DistinctNames() []string {
	names := make([]string, 0, len(c.products))

	var seen collections.Set[string]

	for _, product := range c.products {
		name := product.Name()

		if seen.Add(name) {
			names = append(names, name)
		}
	}

	return names
}
