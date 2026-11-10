package catalog

import (
	"fmt"
	"testing"
)

// Global sinks keep the benchmark results observable to the compiler.
var (
	valueSink   Product
	pointerSink *Product
	listSink    []Product
	catalogSink *Catalog
)

func makeProductValue() Product {
	return Product{
		sku:              SKU{'C', 'O', 'F', 'F', 'E', 'E', '0', '1'},
		name:             "Кофе",
		unitPriceKopecks: 12990,
	}
}

func makeProductPointer() *Product {
	product := Product{
		sku:              SKU{'C', 'O', 'F', 'F', 'E', 'E', '0', '1'},
		name:             "Кофе",
		unitPriceKopecks: 12990,
	}
	return &product
}

func BenchmarkReturnValue(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		valueSink = makeProductValue()
	}
}

func BenchmarkReturnPointer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pointerSink = makeProductPointer()
	}
}

func makeBenchmarkCatalog(b *testing.B, count int) *Catalog {
	b.Helper()
	c, err := NewCatalog(count)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < count; i++ {
		product, err := NewProduct(
			fmt.Sprintf("P%07d", i),
			"Товар",
			int64(i),
		)
		if err != nil {
			b.Fatal(err)
		}
		if err := c.Add(product); err != nil {
			b.Fatal(err)
		}
	}
	return c
}

func BenchmarkCatalogList(b *testing.B) {
	c := makeBenchmarkCatalog(b, 1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		listSink = c.List()
	}
}

func BenchmarkCatalogClone(b *testing.B) {
	c := makeBenchmarkCatalog(b, 1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		catalogSink = c.Clone()
	}
}
