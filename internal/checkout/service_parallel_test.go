package checkout

import (
	"errors"
	"math"
	"testing"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/pricing"
)

// newIsolatedService uses t.Helper for useful failure locations and t.Cleanup
// to check the fixture's invariant even when a subtest fails early.
func newIsolatedService(t *testing.T, price int64) (*Service, catalog.SKU) {
	t.Helper()
	store, err := catalog.NewCatalog(1)
	if err != nil {
		t.Fatal(err)
	}
	product := mustTestProduct(t, "COFFEE01", "Кофе", price)
	if err := store.Add(product); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if store.Count() != 1 {
			t.Errorf("fixture catalog mutated: count=%d", store.Count())
		}
		got, found := store.Find(product.SKU())
		if !found || got != product {
			t.Error("Quote changed the product in its catalog")
		}
	})
	service, err := NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	return service, product.SKU()
}

func TestQuoteParallelIndependentSubtests(t *testing.T) {
	cases := []struct {
		name     string
		price    int64
		quantity int64
		want     int64
		wantErr  error
	}{
		{name: "regular", price: 12990, quantity: 3, want: 38970},
		{name: "free", price: 0, quantity: 3, want: 0},
		{name: "zero quantity", price: 10, quantity: 0, wantErr: pricing.ErrInvalidQuantity},
		{name: "negative quantity", price: 10, quantity: -1, wantErr: pricing.ErrInvalidQuantity},
		{name: "overflow", price: math.MaxInt64, quantity: 2, wantErr: pricing.ErrTotalOverflow},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			service, sku := newIsolatedService(t, tc.price)
			quote, err := service.Quote(sku, tc.quantity)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) || quote != (Quote{}) {
					t.Fatalf("got quote=%+v err=%v; want %v", quote, err, tc.wantErr)
				}
				return
			}
			if err != nil || quote.TotalKopecks != tc.want {
				t.Fatalf("got quote=%+v err=%v; want total=%d", quote, err, tc.want)
			}
		})
	}
}
