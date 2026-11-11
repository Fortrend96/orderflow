package pricing

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrInvalidUnitPrice = errors.New("invalid unit price")
	ErrInvalidQuantity  = errors.New("invalid quantity")
	ErrTotalOverflow    = errors.New("total price overflow")
)

// CalculateLineTotal returns the total price in kopecks for an order line.
func CalculateLineTotal(unitPrice int64, quantity int64) (int64, error) {
	if unitPrice < 0 {
		return 0, fmt.Errorf(
			"%w: got %d",
			ErrInvalidUnitPrice,
			unitPrice,
		)
	}

	if quantity <= 0 {
		return 0, fmt.Errorf(
			"%w: expected positive value, got %d",
			ErrInvalidQuantity,
			quantity,
		)
	}

	if unitPrice > math.MaxInt64/quantity {
		return 0, fmt.Errorf(
			"%w: unit price=%d, quantity=%d",
			ErrTotalOverflow,
			unitPrice,
			quantity,
		)
	}

	return unitPrice * quantity, nil
}

var ErrNegativeAmount = errors.New("order amount cannot be negative")

func CalculateDeliveryFree(orderAmount int64) (int64, error) {
	if orderAmount < 0 {
		return 0, ErrNegativeAmount
	}

	if orderAmount >= 50000 {
		return 0, nil
	}

	return 29900, nil
}
