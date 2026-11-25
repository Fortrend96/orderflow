// Package demo runs the OrderFlow console demonstration.
// Business calculations live in checkout; this package only orchestrates them.
package demo

import (
	"errors"
	"fmt"
	"io"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/catalogcsv"
	"github.com/Fortrend96/orderflow/internal/checkout"
)

// Config contains explicit inputs for a single demonstration run.
// CLI defaults belong to cmd/orderflow, not to this package.
type Config struct {
	SKU              string
	Name             string
	UnitPriceKopecks int64
	Quantity         int64
	ExportCSVPath    string
}

// Run demonstrates catalog, checkout and CSV export without parsing flags.
// It writes the report only after the operations have succeeded.
func Run(cfg Config, output io.Writer) error {
	if output == nil {
		return errors.New("report writer is nil")
	}

	store, err := catalog.NewCatalog(2)
	if err != nil {
		return err
	}

	first, err := catalog.NewProduct(
		cfg.SKU,
		cfg.Name,
		cfg.UnitPriceKopecks,
	)
	if err != nil {
		return err
	}

	if err := store.Add(first); err != nil {
		return err
	}

	second, err := catalog.NewProduct(
		"TEA00001",
		"Чай",
		5500,
	)
	if err != nil {
		return err
	}

	if err := store.Add(second); err != nil {
		return err
	}

	sku := first.SKU()

	orderService, err := checkout.NewService(store)
	if err != nil {
		return err
	}

	quote, err := orderService.Quote(sku, cfg.Quantity)
	if err != nil {
		return err
	}

	// Export occurs before the demonstration changes the price.
	if cfg.ExportCSVPath != "" {
		if err := catalogcsv.ExportFile(
			cfg.ExportCSVPath,
			store.List(),
		); err != nil {
			return fmt.Errorf(
				"export catalog to %q: %w",
				cfg.ExportCSVPath,
				err,
			)
		}
	}

	report := reportData{
		quote:         quote,
		products:      store.List(),
		groups:        store.PriceGroups(),
		names:         store.DistinctNames(),
		exportPath:    cfg.ExportCSVPath,
		originalPrice: first.UnitPriceKopecks(),
	}

	if err := store.UpdatePrice(sku, 15000); err != nil {
		return err
	}

	updated, found := store.Find(sku)
	if !found {
		return fmt.Errorf(
			"updated product not found: %s",
			sku,
		)
	}

	clone := store.Clone()
	if err := clone.UpdatePrice(sku, 22000); err != nil {
		return err
	}

	teaSKU, err := catalog.ParseSKU("TEA00001")
	if err != nil {
		return err
	}

	if !clone.Remove(teaSKU) {
		return errors.New("could not remove tea from clone")
	}

	originalProduct, found := store.Find(sku)
	if !found {
		return errors.New("original catalog lost product")
	}

	cloneProduct, found := clone.Find(sku)
	if !found {
		return errors.New("cloned catalog lost product")
	}

	cloneService, err := checkout.NewService(clone)
	if err != nil {
		return err
	}

	originalQuote, err := orderService.Quote(sku, 2)
	if err != nil {
		return err
	}

	cloneQuote, err := cloneService.Quote(sku, 2)
	if err != nil {
		return err
	}

	report.updatedPrice = updated.UnitPriceKopecks()
	report.originalCount = store.Count()
	report.cloneCount = clone.Count()
	report.originalCurrentPrice = originalProduct.UnitPriceKopecks()
	report.clonedPrice = cloneProduct.UnitPriceKopecks()
	report.originalQuoteForTwo = originalQuote.TotalKopecks
	report.clonedQuoteForTwo = cloneQuote.TotalKopecks

	return writeReport(output, report)
}
