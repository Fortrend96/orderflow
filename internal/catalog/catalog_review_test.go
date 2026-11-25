package catalog

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/Fortrend96/orderflow/internal/collections"
)

func TestCatalogRejectsInvalidCapacityWithSentinel(t *testing.T) {
	store, err := NewCatalog(-7)

	if store != nil || !errors.Is(err, ErrInvalidCapacity) {
		t.Fatalf(
			"NewCatalog(-7) = %v, %v; want nil, ErrInvalidCapacity",
			store,
			err,
		)
	}
}

func TestDistinctNamesPreservesOrderWithManyDuplicates(t *testing.T) {
	store, err := NewCatalog(200)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 200; i++ {
		product, err := NewProduct(
			fmt.Sprintf("P%07d", i),
			fmt.Sprintf("Товар %02d", i%13),
			int64(i),
		)
		if err != nil {
			t.Fatal(err)
		}

		if err := store.Add(product); err != nil {
			t.Fatal(err)
		}
	}

	want := make([]string, 13)

	for i := range want {
		want[i] = fmt.Sprintf("Товар %02d", i)
	}

	got := store.DistinctNames()

	if !slices.Equal(got, want) {
		t.Fatalf(
			"DistinctNames()=%v, want %v",
			got,
			want,
		)
	}

	got[0] = "Изменено"

	if store.DistinctNames()[0] != want[0] {
		t.Fatal(
			"result must not expose mutable Catalog storage",
		)
	}
}

var distinctNamesSink []string

// The old approach first allocated []Product through UniqueBy, then []string.
// We keep that implementation only as a measured benchmark baseline.
func oldDistinctNames(products []Product) []string {
	unique := collections.UniqueBy(
		products,
		Product.Name,
	)

	names := make([]string, 0, len(unique))

	for _, p := range unique {
		names = append(names, p.Name())
	}

	return names
}

func BenchmarkDistinctNames(b *testing.B) {
	store, err := NewCatalog(1000)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		product, err := NewProduct(
			fmt.Sprintf("P%07d", i),
			fmt.Sprintf("Товар %02d", i%50),
			int64(i),
		)
		if err != nil {
			b.Fatal(err)
		}

		if err := store.Add(product); err != nil {
			b.Fatal(err)
		}
	}

	b.Run("single_pass", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			distinctNamesSink = store.DistinctNames()
		}
	})

	b.Run("two_pass_old", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			distinctNamesSink = oldDistinctNames(store.products)
		}
	})
}
