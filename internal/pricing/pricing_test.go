package pricing

import (
	"errors"
	"math"
	"testing"
)

func TestCalculateLineTotal(t *testing.T) {
	tests := []struct {
		name      string
		price     int64
		quantity  int64
		want      int64
		wantError bool
	}{
		{
			name:     "regular order",
			price:    12990,
			quantity: 3,
			want:     38970,
		},
		{
			name:     "zero price",
			price:    0,
			quantity: 2,
			want:     0,
		},
		{
			name:      "negative price",
			price:     -1,
			quantity:  2,
			wantError: true,
		},
		{
			name:      "zero quantity",
			price:     10,
			quantity:  0,
			wantError: true,
		},
		{
			name:      "negative quantity",
			price:     10,
			quantity:  -1,
			wantError: true,
		},
		{
			name:      "overflow",
			price:     math.MaxInt64,
			quantity:  2,
			wantError: true,
		},
		{
			name:     "maximum int64",
			price:    math.MaxInt64,
			quantity: 1,
			want:     math.MaxInt64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateLineTotal(
				tt.price,
				tt.quantity,
			)

			if (err != nil) != tt.wantError {
				t.Fatalf(
					"CalculationLineTotal(%d, %d): err = %v, wantError = %t",
					tt.price,
					tt.quantity,
					err,
					tt.wantError,
				)
			}

			if err != nil {
				return
			}

			if got != tt.want {
				t.Errorf(
					"CalculateLineTotal(%d, %d) = %d, want %d",
					tt.price,
					tt.quantity,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestCalculateDeliveryFee(t *testing.T) {
	tests := []struct {
		name        string
		orderAmount int64
		wantFee     int64
		wantErr     error
	}{
		{
			name:        "Отрицательная сумма",
			orderAmount: -1,
			wantFee:     0,
			wantErr:     ErrNegativeAmount,
		},
		{
			name:        "Нулевая сумма",
			orderAmount: 0,
			wantFee:     29900,
			wantErr:     nil,
		},
		{
			name:        "Значение непосредственно ниже порога",
			orderAmount: 49999, // 499 рублей 99 копеек
			wantFee:     29900,
			wantErr:     nil,
		},
		{
			name:        "Значение точно на пороге",
			orderAmount: 50000, // 500 рублей
			wantFee:     0,
			wantErr:     nil,
		},
		{
			name:        "Значение выше порога",
			orderAmount: 100000, // 1000 рублей
			wantFee:     0,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFee, err := CalculateDeliveryFree(tt.orderAmount)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CalculateDeliveryFee() error = %v, wantErr %v", err, tt.wantErr)
			}

			if gotFee != tt.wantFee {
				t.Errorf("CalculateDeliveryFee() gotFee = %v, wantFee %v", gotFee, tt.wantFee)
			}
		})
	}
}
