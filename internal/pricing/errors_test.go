package pricing

import (
	"errors"
	"math"
	"testing"
)

func TestCalculateLineTotalErrorKinds(t *testing.T) {
	tests := []struct {
		name     string
		price    int64
		quantity int64
		want     error
	}{
		{
			name:     "negative price",
			price:    -1,
			quantity: 2,
			want:     ErrInvalidUnitPrice,
		},
		{
			name:     "zero quantity",
			price:    100,
			quantity: 0,
			want:     ErrInvalidQuantity,
		},
		{
			name:     "negative quantity",
			price:    100,
			quantity: -1,
			want:     ErrInvalidQuantity,
		},
		{
			name:     "overflow",
			price:    math.MaxInt64,
			quantity: 2,
			want:     ErrTotalOverflow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateLineTotal(
				tt.price,
				tt.quantity,
			)

			if !errors.Is(err, tt.want) {
				t.Fatalf(
					"got %v; want error kind %v",
					err,
					tt.want,
				)
			}
		})
	}
}

func TestCalculateLineTotalSuccessHasNoError(
	t *testing.T,
) {
	got, err := CalculateLineTotal(12990, 3)

	if err != nil || got != 38970 {
		t.Fatalf(
			"got=%d, err=%v; want 38970, nil",
			got,
			err,
		)
	}
}
