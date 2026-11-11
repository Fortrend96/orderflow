package checkout

import (
	"errors"
	"fmt"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/pricing"
)

// ProductFinder is the minimum behavior needed by Service.
// It is implemented implicitly by *catalog.Catalog.
type ProductFinder interface {
	Find(sku catalog.SKU) (catalog.Product, bool)
}

// Quote contains a calculated order-line total in kopecks.
type Quote struct {
	SKU              catalog.SKU
	ProductName      string
	Quantity         int64
	UnitPriceKopecks int64
	TotalKopecks     int64
}

// QuoteError describes which quote failed while retaining its underlying cause.
type QuoteError struct {
	SKU      catalog.SKU
	Quantity int64
	Err      error
}

func (e *QuoteError) Error() string {
	return fmt.Sprintf(
		"quote SKU %s, quantity %d: %v",
		e.SKU.String(),
		e.Quantity,
		e.Err,
	)
}

func (e *QuoteError) Unwrap() error {
	return e.Err
}

// Service calculates order-line quotes using an injected product finder.
type Service struct {
	finder ProductFinder
}

func NewService(finder ProductFinder) (*Service, error) {
	if finder == nil {
		return nil, errors.New("product finder is nil")
	}

	return &Service{finder: finder}, nil
}

func (s *Service) Quote(
	sku catalog.SKU,
	quantity int64,
) (Quote, error) {
	product, found := s.finder.Find(sku)

	if !found {
		return Quote{}, &QuoteError{
			SKU:      sku,
			Quantity: quantity,
			Err: fmt.Errorf(
				"find product: %w",
				catalog.ErrProductNotFound,
			),
		}
	}

	total, err := pricing.CalculateLineTotal(
		product.UnitPriceKopecks(),
		quantity,
	)

	if err != nil {
		return Quote{}, &QuoteError{
			SKU:      sku,
			Quantity: quantity,
			Err: fmt.Errorf(
				"calculate total: %w",
				err,
			),
		}
	}

	return Quote{
		SKU:              product.SKU(),
		ProductName:      product.Name(),
		Quantity:         quantity,
		UnitPriceKopecks: product.UnitPriceKopecks(),
		TotalKopecks:     total,
	}, nil

}
