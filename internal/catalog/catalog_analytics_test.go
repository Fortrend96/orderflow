package catalog

import (
	"slices"
	"testing"
)

func TestCatalogPriceGroupsAndDistinctNames(t *testing.T) {
	store := mustCatalog(t, 4)
	coffee1 := mustProduct(t, "COFFEE01", "Кофе", 12990)
	tea := mustProduct(t, "TEA00001", "Чай", 5500)
	coffee2 := mustProduct(t, "COFFEE02", "Кофе", 12990)
	milk := mustProduct(t, "MILK0001", "Молоко", 5500)
	for _, p := range []Product{coffee1, tea, coffee2, milk} {
		if err := store.Add(p); err != nil {
			t.Fatal(err)
		}
	}
	if names := store.DistinctNames(); !slices.Equal(names, []string{"Кофе", "Чай", "Молоко"}) {
		t.Fatalf("names=%v", names)
	}
	groups := store.PriceGroups()
	if len(groups) != 2 || !slices.Equal(groups[12990], []Product{coffee1, coffee2}) ||
		!slices.Equal(groups[5500], []Product{tea, milk}) {
		t.Fatalf("groups=%v", groups)
	}

	// Result slices and map must be independent of the Catalog.
	groups[12990][0] = milk
	delete(groups, 5500)
	names := store.DistinctNames()
	names[0] = "Modified"
	got, found := store.Find(coffee1.SKU())
	if !found || got != coffee1 || store.Count() != 4 {
		t.Fatal("mutating analytics results changed Catalog")
	}
	if gotNames := store.DistinctNames(); gotNames[0] != "Кофе" {
		t.Fatalf("names unexpectedly changed: %v", gotNames)
	}
	requireIndexConsistent(t, store)
}

func TestCatalogAnalyticsZeroValue(t *testing.T) {
	var store Catalog
	if names := store.DistinctNames(); names == nil || len(names) != 0 {
		t.Fatalf("names=%v; want independent empty slice", names)
	}
	groups := store.PriceGroups()
	if groups == nil || len(groups) != 0 {
		t.Fatalf("groups=%v; want initialized empty map", groups)
	}
}
