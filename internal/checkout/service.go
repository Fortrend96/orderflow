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
		return Quote{}, fmt.Errorf(
			"%w: %s",
			catalog.ErrProductNotFound,
			sku.String(),
		)
	}

	total, err := pricing.CalculateLineTotal(
		product.UnitPriceKopecks(),
		quantity,
	)

	if err != nil {
		return Quote{}, fmt.Errorf(
			"calculate total for SKU %s: %w",
			sku.String(),
			err,
		)
	}

	return Quote{
		SKU:              product.SKU(),
		ProductName:      product.Name(),
		Quantity:         quantity,
		UnitPriceKopecks: product.UnitPriceKopecks(),
		TotalKopecks:     total,
	}, nil

}
