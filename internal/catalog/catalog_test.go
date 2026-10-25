package catalog

import (
	"errors"
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

func TestNewCatalogRejectsNegativeCapacity(t *testing.T) {
	_, err := NewCatalog(-1)

	if err == nil {
		t.Fatal("expected an error for negative capacity")
	}
}

func TestCatalogAddFindAndDuplicate(t *testing.T) {
	store, err := NewCatalog(1)

	if err != nil {
		t.Fatal(err)
	}

	product := mustProduct(
		t,
		"COFFEE01",
		"Кофе",
		12990,
	)

	if err := store.Add(product); err != nil {
		t.Fatal(err)
	}

	if store.Count() != 1 {
		t.Fatalf(
			"count = %d, want 1",
			store.Count(),
		)
	}

	got, ok := store.Find(product.SKU)

	if !ok || got != product {
		t.Fatalf(
			"got = %+v, found = %t",
			got,
			ok,
		)
	}

	if err := store.Add(product); !errors.Is(
		err,
		ErrProductExists,
	) {
		t.Fatalf(
			"duplicate add error = %v, want ErrProductExists",
			err,
		)
	}

	if store.Count() != 1 {
		t.Fatalf(
			"count = %d, want 1",
			store.Count(),
		)
	}
}

func TestCatalogRejectsInvalidProduct(t *testing.T) {
	store, _ := NewCatalog(0)

	if err := store.Add(Product{}); err == nil {
		t.Fatal("expected invalid product error")
	}

	if store.Count() != 0 {
		t.Fatalf(
			"count = %d, want 0",
			store.Count(),
		)
	}
}

func TestCatalogListIsIndependent(t *testing.T) {
	store, _ := NewCatalog(2)

	first := mustProduct(
		t, "COFFEE01", "Кофе", 12990,
	)

	second := mustProduct(
		t, "TEA00001", "Чай", 5500,
	)

	if err := store.Add(first); err != nil {
		t.Fatal(err)
	}

	snapshot := store.List()

	snapshot[0].Name = "Изменённое название"

	if err := store.Add(second); err != nil {
		t.Fatal(err)
	}

	if len(snapshot) != 1 {
		t.Fatalf(
			"snapshot len = %d, want 1",
			len(snapshot),
		)
	}

	got, found := store.Find(first.SKU)

	if !found || got.Name != "Кофе" {
		t.Fatalf(
			"catalog changed through snapshot: %+v",
			got,
		)
	}
}

func TestCatalogRemove(t *testing.T) {
	store, _ := NewCatalog(3)

	a := mustProduct(t, "COFFEE01", "Кофе", 12990)
	b := mustProduct(t, "TEA00001", "Чай", 5500)
	c := mustProduct(t, "MILK0001", "Молоко", 8800)

	for _, p := range []Product{a, b, c} {
		if err := store.Add(p); err != nil {
			t.Fatal(err)
		}
	}

	if !store.Remove(b.SKU) {
		t.Fatal("expected to remove second product")
	}

	if store.Count() != 2 {
		t.Fatalf(
			"count = %d, want 2",
			store.Count(),
		)
	}

	got := store.List()

	if got[0] != a || got[1] != c {
		t.Fatalf(
			"wrong remaining order: %+v",
			got,
		)
	}

	if store.Remove(b.SKU) {
		t.Fatal("expected false when removing again")
	}

	if _, found := store.Find(b.SKU); found {
		t.Fatal("removed product still found")
	}

	if store.products[:cap(store.products)][2] != (Product{}) {
		t.Fatalf(
			"unused tail not cleared: %+v",
			store.products[:cap(store.products)][2],
		)
	}

	if !store.Remove(a.SKU) ||
		!store.Remove(c.SKU) ||
		store.Count() != 0 {
		t.Fatal("failed to remove remaining products")
	}
}

func TestCatalogGrows(t *testing.T) {
	store, err := NewCatalog(1)

	if err != nil {
		t.Fatal(err)
	}

	first := mustProduct(
		t, "COFFEE01", "Кофе", 12990,
	)

	second := mustProduct(
		t, "TEA00001", "Чай", 5500,
	)

	if err := store.Add(first); err != nil {
		t.Fatal(err)
	}

	if err := store.Add(second); err != nil {
		t.Fatal(err)
	}

	if store.Count() != 2 {
		t.Fatalf(
			"count = %d, want 2",
			store.Count(),
		)
	}

	if _, found := store.Find(second.SKU); !found {
		t.Fatal("second product not found after growth")
	}
}
