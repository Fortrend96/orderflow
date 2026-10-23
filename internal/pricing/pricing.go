package pricing

import (
	"errors"
	"fmt"
	"math"
)

// CalculateLineTotal returns the total price in kopecks for an order line.
func CalculateLineTotal(unitPrice int64, quantity int64) (int64, error) {
	if unitPrice < 0 {
		return 0, fmt.Errorf(
			"unit price must be non-negative: %d",
			unitPrice,
		)
	}

	if quantity <= 0 {
		return 0, fmt.Errorf(
			"quantity must be positive: %d",
			quantity,
		)
	}

	if unitPrice > math.MaxInt64/quantity {
		return 0, fmt.Errorf("total price overflows int64")
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
