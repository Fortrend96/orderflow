package catalog

import "testing"

func TestCatalogCloneIndependent(t *testing.T) {
	original := mustCatalog(t, 2)
	coffee := mustProduct(t, "COFFEE01", "Кофе", 12990)
	tea := mustProduct(t, "TEA00001", "Чай", 5500)
	milk := mustProduct(t, "MILK0001", "Молоко", 8800)

	for _, p := range []Product{coffee, tea, milk} {
		if err := original.Add(p); err != nil {
			t.Fatal(err)
		}
	}

	clone := original.Clone()
	if clone == original {
		t.Fatal("Clone returned the original pointer")
	}
	if &clone.products[0] == &original.products[0] {
		t.Fatal("Clone reused the original backing array")
	}

	requireIndexConsistent(t, original)
	requireIndexConsistent(t, clone)

	for i, p := range original.List() {
		if clone.products[i] != p {
			t.Fatalf("clone product %d differs from original", i)
		}
	}

	if err := clone.UpdatePrice(coffee.SKU(), 20000); err != nil {
		t.Fatal(err)
	}
	if !clone.Remove(tea.SKU()) {
		t.Fatal("could not remove tea from clone")
	}
	bread := mustProduct(t, "BREAD001", "Хлеб", 3200)
	if err := clone.Add(bread); err != nil {
		t.Fatal(err)
	}

	gotOriginal, found := original.Find(coffee.SKU())
	if !found || gotOriginal.UnitPriceKopecks() != 12990 {
		t.Fatal("updating clone affected original")
	}
	if original.Count() != 3 || clone.Count() != 3 {
		t.Fatalf("counts: original=%d clone=%d", original.Count(), clone.Count())
	}
	if _, found := original.Find(tea.SKU()); !found {
		t.Fatal("removing from clone deleted tea from original")
	}
	if _, found := original.Find(bread.SKU()); found {
		t.Fatal("adding to clone changed original")
	}

	expected := []SKU{coffee.SKU(), milk.SKU(), bread.SKU()}
	for i, product := range clone.List() {
		if product.SKU() != expected[i] {
			t.Fatalf("clone order at %d: %s", i, product.SKU())
		}
	}

	requireIndexConsistent(t, original)
	requireIndexConsistent(t, clone)
}

func TestCatalogCloneEmptyAndZeroValue(t *testing.T) {
	var zero Catalog
	clone := zero.Clone()

	if clone == nil || clone.Count() != 0 || clone.indexBySKU == nil {
		t.Fatal("Clone must return an initialized empty catalog")
	}

	coffee := mustProduct(t, "COFFEE01", "Кофе", 100)
	if err := clone.Add(coffee); err != nil {
		t.Fatal(err)
	}
	if zero.Count() != 0 {
		t.Fatal("mutating clone affected zero catalog")
	}

	requireIndexConsistent(t, &zero)
	requireIndexConsistent(t, clone)
}

func TestShallowCatalogCopySharesStorage(t *testing.T) {
	original := mustCatalog(t, 1)
	coffee := mustProduct(t, "COFFEE01", "Кофе", 12990)
	if err := original.Add(coffee); err != nil {
		t.Fatal(err)
	}

	// Educational anti-example: copying the struct does not clone its storage.
	shallow := *original
	if err := shallow.UpdatePrice(coffee.SKU(), 15000); err != nil {
		t.Fatal(err)
	}

	got, found := original.Find(coffee.SKU())
	if !found || got.UnitPriceKopecks() != 15000 {
		t.Fatal("expected shallow copy to share underlying Product storage")
	}
}
