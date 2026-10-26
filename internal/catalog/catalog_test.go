package catalog

import (
	"errors"
	"fmt"
	"testing"
)

func mustProduct(
	t *testing.T,
	rawSKU, name string,
	price int64,
) Product {
	t.Helper()

	product, err := NewProduct(rawSKU, name, price)

	if err != nil {
		t.Fatal(err)
	}

	return product
}

func mustCatalog(t *testing.T, capacity int) *Catalog {
	t.Helper()

	c, err := NewCatalog(capacity)

	if err != nil {
		t.Fatal(err)
	}

	return c
}

// requireIndexConsistent checks the catalog's internal invariant.
func requireIndexConsistent(t *testing.T, c *Catalog) {
	t.Helper()

	if len(c.products) != len(c.indexBySKU) {
		t.Fatalf(
			"products len=%d, index len=%d",
			len(c.products),
			len(c.indexBySKU),
		)
	}

	for i, p := range c.products {
		got, ok := c.indexBySKU[p.SKU]

		if !ok || got != i {
			t.Fatalf(
				"SKU %s: index=%d, found=%t, want=%d",
				p.SKU,
				got,
				ok,
				i,
			)
		}
	}
}

func TestNewCatalogRejectsNegativeCapacity(t *testing.T) {
	if _, err := NewCatalog(-1); err == nil {
		t.Fatal("expected negative capacity error")
	}
}

func TestCatalogAddFindAndDuplicate(t *testing.T) {
	c := mustCatalog(t, 1)

	first := mustProduct(
		t,
		"COFFEE01",
		"Кофе",
		12990,
	)

	if err := c.Add(first); err != nil {
		t.Fatal(err)
	}

	if got := c.Count(); got != 1 {
		t.Fatalf("count=%d, want 1", got)
	}

	actual, ok := c.Find(first.SKU)

	if !ok || actual != first {
		t.Fatalf(
			"find got=%+v, found=%t",
			actual,
			ok,
		)
	}

	if err := c.Add(first); !errors.Is(
		err,
		ErrProductExists,
	) {
		t.Fatalf("duplicate error=%v", err)
	}

	if got := c.Count(); got != 1 {
		t.Fatalf("count after duplicate=%d", got)
	}

	missing, err := ParseSKU("UNKNOWN1")

	if err != nil {
		t.Fatal(err)
	}

	if _, ok := c.Find(missing); ok {
		t.Fatal("unexpectedly found missing product")
	}

	requireIndexConsistent(t, c)
}

func TestCatalogRejectsInvalidProduct(t *testing.T) {
	c := mustCatalog(t, 0)

	if err := c.Add(Product{}); err == nil {
		t.Fatal("expected invalid product error")
	}

	if c.Count() != 0 {
		t.Fatal("invalid product added")
	}

	requireIndexConsistent(t, c)
}

func TestCatalogListIsIndependent(t *testing.T) {
	c := mustCatalog(t, 2)

	first := mustProduct(
		t,
		"COFFEE01",
		"Кофе",
		12990,
	)

	second := mustProduct(
		t,
		"TEA00001",
		"Чай",
		5500,
	)

	if err := c.Add(first); err != nil {
		t.Fatal(err)
	}

	snapshot := c.List()

	snapshot[0].Name = "Изменено"

	if err := c.Add(second); err != nil {
		t.Fatal(err)
	}

	if len(snapshot) != 1 {
		t.Fatalf(
			"snapshot len=%d",
			len(snapshot),
		)
	}

	got, ok := c.Find(first.SKU)

	if !ok || got.Name != "Кофе" {
		t.Fatalf(
			"internal state changed through List: %+v",
			got,
		)
	}

	requireIndexConsistent(t, c)
}

func TestCatalogRemoveMaintainsIndexes(t *testing.T) {
	c := mustCatalog(t, 3)

	products := []Product{
		mustProduct(t, "COFFEE01", "Кофе", 12990),
		mustProduct(t, "TEA00001", "Чай", 5500),
		mustProduct(t, "MILK0001", "Молоко", 8800),
		mustProduct(t, "BREAD001", "Хлеб", 3200),
	}

	for _, p := range products {
		if err := c.Add(p); err != nil {
			t.Fatal(err)
		}
	}

	requireIndexConsistent(t, c)

	if !c.Remove(products[1].SKU) {
		t.Fatal("remove middle failed")
	}

	expected := []Product{
		products[0],
		products[2],
		products[3],
	}

	for i, p := range c.List() {
		if p != expected[i] {
			t.Fatalf(
				"after middle remove index %d: %+v",
				i,
				p,
			)
		}
	}

	requireIndexConsistent(t, c)

	if !c.Remove(products[0].SKU) {
		t.Fatal("remove first failed")
	}

	requireIndexConsistent(t, c)

	if !c.Remove(products[3].SKU) {
		t.Fatal("remove last failed")
	}

	requireIndexConsistent(t, c)

	if c.Remove(products[3].SKU) {
		t.Fatal("second remove must return false")
	}

	if err := c.Add(products[1]); err != nil {
		t.Fatal(err)
	}

	if err := c.Add(products[3]); err != nil {
		t.Fatal(err)
	}

	requireIndexConsistent(t, c)

	if len(c.products) != 3 ||
		c.products[0] != products[2] ||
		c.products[1] != products[1] ||
		c.products[2] != products[3] {
		t.Fatalf(
			"insertion order unexpected: %+v",
			c.List(),
		)
	}
}

func TestCatalogClearsRemovedTail(t *testing.T) {
	c := mustCatalog(t, 3)

	p := mustProduct(
		t,
		"COFFEE01",
		"Кофе",
		12990,
	)

	if err := c.Add(p); err != nil {
		t.Fatal(err)
	}

	if !c.Remove(p.SKU) {
		t.Fatal("remove failed")
	}

	if c.products[:cap(c.products)][0] != (Product{}) {
		t.Fatal("tail not cleared")
	}

	requireIndexConsistent(t, c)
}

func TestCatalogUpdatePrice(t *testing.T) {
	c := mustCatalog(t, 0)

	p := mustProduct(
		t,
		"COFFEE01",
		"Кофе",
		12990,
	)

	if err := c.Add(p); err != nil {
		t.Fatal(err)
	}

	if err := c.UpdatePrice(p.SKU, -10); err == nil {
		t.Fatal("expected price validation error")
	}

	got, _ := c.Find(p.SKU)

	if got.UnitPriceKopecks != 12990 {
		t.Fatal("price changed after invalid update")
	}

	missing, err := ParseSKU("UNKNOWN1")

	if err != nil {
		t.Fatal(err)
	}

	if err := c.UpdatePrice(
		missing,
		0,
	); !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("missing SKU error=%v", err)
	}

	if err := c.UpdatePrice(p.SKU, 0); err != nil {
		t.Fatal(err)
	}

	got, ok := c.Find(p.SKU)

	if !ok || got.UnitPriceKopecks != 0 {
		t.Fatalf(
			"updated price=%d, found=%t",
			got.UnitPriceKopecks,
			ok,
		)
	}

	requireIndexConsistent(t, c)
}

func TestCatalogZeroValue(t *testing.T) {
	var c Catalog

	product := mustProduct(
		t,
		"COFFEE01",
		"Кофе",
		1,
	)

	if _, found := c.Find(product.SKU); found {
		t.Fatal("unexpected zero value hit")
	}

	if c.Remove(product.SKU) {
		t.Fatal("unexpected zero value remove")
	}

	if err := c.Add(product); err != nil {
		t.Fatal(err)
	}

	if _, found := c.Find(product.SKU); !found {
		t.Fatal("zero value catalog failed")
	}

	requireIndexConsistent(t, &c)
}

func TestCatalogGrows(t *testing.T) {
	c := mustCatalog(t, 1)

	for i := 0; i < 30; i++ {
		sku := fmt.Sprintf("P%07d", i)

		if err := c.Add(
			mustProduct(t, sku, "Товар", int64(i)),
		); err != nil {
			t.Fatal(err)
		}
	}

	if c.Count() != 30 {
		t.Fatalf("count=%d", c.Count())
	}

	requireIndexConsistent(t, c)

	for i := 0; i < 30; i++ {
		sku, _ := ParseSKU(
			fmt.Sprintf("P%07d", i),
		)

		p, ok := c.Find(sku)

		if !ok || p.UnitPriceKopecks != int64(i) {
			t.Fatalf("cannot find %s", sku)
		}
	}
}

func BenchmarkCatalogFind(b *testing.B) {
	c, err := NewCatalog(10000)

	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 10000; i++ {
		p, err := NewProduct(
			fmt.Sprintf("P%07d", i),
			"Товар",
			int64(i),
		)

		if err != nil {
			b.Fatal(err)
		}

		if err := c.Add(p); err != nil {
			b.Fatal(err)
		}
	}

	sku, err := ParseSKU("P0009999")

	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, found := c.Find(sku)

		if !found {
			b.Fatal("product not found")
		}
	}
}

func BenchmarkLinearFind(b *testing.B) {
	products := make([]Product, 0, 10000)

	for i := 0; i < 10000; i++ {
		p, err := NewProduct(
			fmt.Sprintf("P%07d", i),
			"Товар",
			int64(i),
		)

		if err != nil {
			b.Fatal(err)
		}

		products = append(products, p)
	}

	sku, err := ParseSKU("P0009999")

	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		found := false

		for _, p := range products {
			if p.SKU == sku {
				found = true
				break
			}
		}

		if !found {
			b.Fatal("product not found")
		}
	}
}
