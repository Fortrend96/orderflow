package checkout

import (
	"errors"
	"math"
	"testing"

	"github.com/Fortrend96/orderflow/internal/catalog"
)

// This fake has a pointer receiver:
// only *fakeFinder implements ProductFinder.
type fakeFinder struct {
	products map[catalog.SKU]catalog.Product
	calls    int
}

func (f *fakeFinder) Find(
	sku catalog.SKU,
) (catalog.Product, bool) {
	f.calls++

	product, found := f.products[sku]

	return product, found
}

// This fake has a value receiver:
// both fixedFinder and *fixedFinder implement ProductFinder.
type fixedFinder struct {
	product catalog.Product
}

func (f fixedFinder) Find(
	sku catalog.SKU,
) (catalog.Product, bool) {
	if f.product.SKU() != sku {
		return catalog.Product{}, false
	}

	return f.product, true
}

// Compile-time checks of implicit interface
// implementation and method sets.
var (
	_ ProductFinder = (*catalog.Catalog)(nil)
	_ ProductFinder = (*fakeFinder)(nil)
	_ ProductFinder = fixedFinder{}
	_ ProductFinder = (*fixedFinder)(nil)
)

func mustTestProduct(
	t *testing.T,
	sku, name string,
	price int64,
) catalog.Product {
	t.Helper()

	product, err := catalog.NewProduct(
		sku,
		name,
		price,
	)

	if err != nil {
		t.Fatal(err)
	}

	return product
}

func TestNewServiceRejectsNilInterface(t *testing.T) {
	service, err := NewService(nil)

	if err == nil || service != nil {
		t.Fatalf(
			"service=%v, err=%v; want nil service and error",
			service,
			err,
		)
	}
}

func TestServiceQuoteWithFake(t *testing.T) {
	coffee := mustTestProduct(
		t,
		"COFFEE01",
		"Кофе",
		12990,
	)

	free := mustTestProduct(
		t,
		"FREE0001",
		"Подарок",
		0,
	)

	finder := &fakeFinder{
		products: map[catalog.SKU]catalog.Product{
			coffee.SKU(): coffee,
			free.SKU():   free,
		},
	}

	service, err := NewService(finder)

	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		sku      catalog.SKU
		quantity int64
		want     int64
		wantErr  bool
	}{
		{
			name:     "regular order",
			sku:      coffee.SKU(),
			quantity: 3,
			want:     38970,
		},
		{
			name:     "zero price",
			sku:      free.SKU(),
			quantity: 2,
			want:     0,
		},
		{
			name:     "zero quantity",
			sku:      coffee.SKU(),
			quantity: 0,
			wantErr:  true,
		},
		{
			name:     "negative quantity",
			sku:      coffee.SKU(),
			quantity: -2,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quote, err := service.Quote(
				tt.sku,
				tt.quantity,
			)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"Quote error=%v, wantErr=%t",
					err,
					tt.wantErr,
				)
			}

			if err != nil {
				if quote != (Quote{}) {
					t.Fatalf(
						"error result should be zero Quote, got %+v",
						quote,
					)
				}

				return
			}

			if quote.SKU != tt.sku ||
				quote.Quantity != tt.quantity ||
				quote.TotalKopecks != tt.want {
				t.Fatalf(
					"unexpected quote: %+v",
					quote,
				)
			}
		})
	}

	if finder.calls != len(tests) {
		t.Fatalf(
			"Find called %d times, want %d",
			finder.calls,
			len(tests),
		)
	}
}

func TestServiceQuoteNotFound(t *testing.T) {
	finder := &fakeFinder{
		products: make(
			map[catalog.SKU]catalog.Product,
		),
	}

	service, err := NewService(finder)

	if err != nil {
		t.Fatal(err)
	}

	sku, err := catalog.ParseSKU("MISSING1")

	if err != nil {
		t.Fatal(err)
	}

	quote, err := service.Quote(sku, 2)

	if !errors.Is(
		err,
		catalog.ErrProductNotFound,
	) {
		t.Fatalf(
			"error=%v, want ErrProductNotFound",
			err,
		)
	}

	if quote != (Quote{}) {
		t.Fatalf(
			"missing product should return zero Quote, got %+v",
			quote,
		)
	}

	if finder.calls != 1 {
		t.Fatalf(
			"Find called %d times, want 1",
			finder.calls,
		)
	}
}

func TestServiceQuoteOverflow(t *testing.T) {
	product := mustTestProduct(
		t,
		"MAX00001",
		"Товар",
		math.MaxInt64,
	)

	service, err := NewService(
		fixedFinder{product: product},
	)

	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.Quote(
		product.SKU(),
		2,
	); err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestServiceAcceptsValueReceiverImplementation(
	t *testing.T,
) {
	product := mustTestProduct(
		t,
		"COFFEE01",
		"Кофе",
		100,
	)

	service, err := NewService(
		fixedFinder{product: product},
	)

	if err != nil {
		t.Fatal(err)
	}

	quote, err := service.Quote(
		product.SKU(),
		2,
	)

	if err != nil || quote.TotalKopecks != 200 {
		t.Fatalf(
			"quote=%+v, err=%v",
			quote,
			err,
		)
	}
}

func TestServiceWithRealCatalog(t *testing.T) {
	store, err := catalog.NewCatalog(1)

	if err != nil {
		t.Fatal(err)
	}

	coffee := mustTestProduct(
		t,
		"COFFEE01",
		"Кофе",
		12990,
	)

	if err := store.Add(coffee); err != nil {
		t.Fatal(err)
	}

	originalService, err := NewService(store)

	if err != nil {
		t.Fatal(err)
	}

	clone := store.Clone()

	cloneService, err := NewService(clone)

	if err != nil {
		t.Fatal(err)
	}

	if err := clone.UpdatePrice(
		coffee.SKU(),
		20000,
	); err != nil {
		t.Fatal(err)
	}

	original, err := originalService.Quote(
		coffee.SKU(),
		2,
	)

	if err != nil {
		t.Fatal(err)
	}

	copied, err := cloneService.Quote(
		coffee.SKU(),
		2,
	)

	if err != nil {
		t.Fatal(err)
	}

	if original.TotalKopecks != 25980 ||
		copied.TotalKopecks != 40000 {
		t.Fatalf(
			"original=%d, clone=%d",
			original.TotalKopecks,
			copied.TotalKopecks,
		)
	}
}

func TestTypedNilInterfaceIsNotNil(t *testing.T) {
	var pointer *fakeFinder

	var finder ProductFinder = pointer

	if finder == nil {
		t.Fatal(
			"interface with dynamic type *fakeFinder must not be nil",
		)
	}

	// Do not call finder.Find:
	// the pointer receiver would dereference nil.
}
