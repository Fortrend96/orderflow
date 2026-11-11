package checkout

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/pricing"
)

func TestQuoteErrorSupportsIsAndAs(t *testing.T) {
	product := mustTestProduct(
		t,
		"COFFEE01",
		"Кофе",
		12990,
	)

	service, err := NewService(
		fixedFinder{product: product},
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		sku      catalog.SKU
		quantity int64
		want     error
	}{
		{
			name:     "missing product",
			sku:      mustSKU(t, "MISSING1"),
			quantity: 2,
			want:     catalog.ErrProductNotFound,
		},
		{
			name:     "zero quantity",
			sku:      product.SKU(),
			quantity: 0,
			want:     pricing.ErrInvalidQuantity,
		},
		{
			name:     "negative quantity",
			sku:      product.SKU(),
			quantity: -1,
			want:     pricing.ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quote, err := service.Quote(
				tt.sku,
				tt.quantity,
			)

			if err == nil || quote != (Quote{}) {
				t.Fatalf(
					"quote=%+v, err=%v; want zero quote and error",
					quote,
					err,
				)
			}

			if !errors.Is(err, tt.want) {
				t.Fatalf(
					"errors.Is(%v, %v) = false",
					err,
					tt.want,
				)
			}

			if err == tt.want {
				t.Fatal(
					"expected a wrapped error, not bare sentinel",
				)
			}

			// An outer wrapper must not prevent discovering QuoteError.
			outer := fmt.Errorf(
				"handle request: %w",
				err,
			)

			var quoteErr *QuoteError
			if !errors.As(outer, &quoteErr) {
				t.Fatalf(
					"errors.As failed for %v",
					outer,
				)
			}

			if quoteErr.SKU != tt.sku ||
				quoteErr.Quantity != tt.quantity {
				t.Fatalf(
					"wrong structured context: %+v",
					quoteErr,
				)
			}

			if !errors.Is(outer, tt.want) {
				t.Fatalf(
					"outer wrapper lost original error: %v",
					outer,
				)
			}

			if errors.Unwrap(err) == nil {
				t.Fatal(
					"QuoteError should provide Unwrap",
				)
			}
		})
	}
}

func TestQuoteErrorOverflow(t *testing.T) {
	product := mustTestProduct(
		t,
		"COFFEE01",
		"Кофе",
		math.MaxInt64,
	)

	service, err := NewService(
		fixedFinder{product: product},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Quote(product.SKU(), 2)

	if !errors.Is(err, pricing.ErrTotalOverflow) {
		t.Fatalf(
			"got %v; want overflow",
			err,
		)
	}

	var quoteErr *QuoteError

	if !errors.As(err, &quoteErr) ||
		quoteErr.Quantity != 2 {
		t.Fatalf(
			"missing QuoteError with quantity 2: %v",
			err,
		)
	}
}

func TestWrappedVersusFormattedError(t *testing.T) {
	wrapped := fmt.Errorf(
		"checkout: %w",
		pricing.ErrInvalidQuantity,
	)

	formatted := fmt.Errorf(
		"checkout: %v",
		pricing.ErrInvalidQuantity,
	)

	if !errors.Is(wrapped, pricing.ErrInvalidQuantity) {
		t.Fatal(
			"wrapping must preserve error identity",
		)
	}

	if errors.Is(formatted, pricing.ErrInvalidQuantity) {
		t.Fatal(
			"formatting without wrapping must not preserve the error chain",
		)
	}
}

func TestErrorsJoinTree(t *testing.T) {
	joined := errors.Join(
		catalog.ErrProductNotFound,
		pricing.ErrInvalidQuantity,
	)

	if !errors.Is(joined, catalog.ErrProductNotFound) ||
		!errors.Is(joined, pricing.ErrInvalidQuantity) {
		t.Fatal(
			"errors.Join must allow Is to match both causes",
		)
	}
}

func mustSKU(
	t *testing.T,
	raw string,
) catalog.SKU {
	t.Helper()

	sku, err := catalog.ParseSKU(raw)
	if err != nil {
		t.Fatal(err)
	}

	return sku
}
