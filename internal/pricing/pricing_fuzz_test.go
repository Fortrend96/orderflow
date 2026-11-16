package pricing

import (
	"errors"
	"math"
	"math/big"
	"testing"
)

// FuzzCalculateLineTotal uses math/big as an independent overflow oracle.
func FuzzCalculateLineTotal(f *testing.F) {
	for _, seed := range [][2]int64{
		{0, 1}, {100, 3}, {100, 0}, {-1, 2}, {100, -1},
		{math.MaxInt64, 1}, {math.MaxInt64, 2}, {math.MaxInt64 / 2, 2},
	} {
		f.Add(seed[0], seed[1])
	}

	f.Fuzz(func(t *testing.T, price, quantity int64) {
		got, err := CalculateLineTotal(price, quantity)
		switch {
		case price < 0:
			if got != 0 || !errors.Is(err, ErrInvalidUnitPrice) {
				t.Fatalf("price=%d quantity=%d: got %d, %v", price, quantity, got, err)
			}
		case quantity <= 0:
			if got != 0 || !errors.Is(err, ErrInvalidQuantity) {
				t.Fatalf("price=%d quantity=%d: got %d, %v", price, quantity, got, err)
			}
		default:
			exact := new(big.Int).Mul(big.NewInt(price), big.NewInt(quantity))
			if !exact.IsInt64() {
				if got != 0 || !errors.Is(err, ErrTotalOverflow) {
					t.Fatalf("expected overflow for %d * %d, got %d, %v", price, quantity, got, err)
				}
			} else if err != nil || got != exact.Int64() {
				t.Fatalf("price=%d quantity=%d: got %d, %v; exact=%d", price, quantity, got, err, exact.Int64())
			}
		}
	})
}
