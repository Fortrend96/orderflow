package catalog

import "github.com/Fortrend96/orderflow/internal/collections"

// PriceGroups returns independent groups of products by price in kopecks.
// Products inside each group retain insertion order. Map key order is undefined.
func (c *Catalog) PriceGroups() map[int64][]Product {
	return collections.GroupBy(c.products, Product.UnitPriceKopecks)
}

// DistinctNames returns names in the order of their first appearance.
// Products with the same name but different SKUs contribute only one name.
func (c *Catalog) DistinctNames() []string {
	unique := collections.UniqueBy(c.products, Product.Name)
	names := make([]string, 0, len(unique))
	for _, product := range unique {
		names = append(names, product.Name())
	}
	return names
}
