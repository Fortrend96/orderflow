// Package demo runs the OrderFlow console demonstration.
// Business calculations live in checkout; this package only orchestrates them.
package demo

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

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

	first, err := catalog.NewProduct(cfg.SKU, cfg.Name, cfg.UnitPriceKopecks)
	if err != nil {
		return err
	}
	if err := store.Add(first); err != nil {
		return err
	}

	second, err := catalog.NewProduct("TEA00001", "Чай", 5500)
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
		if err := catalogcsv.ExportFile(cfg.ExportCSVPath, store.List()); err != nil {
			return fmt.Errorf("export catalog to %q: %w", cfg.ExportCSVPath, err)
		}
	}

	groups := store.PriceGroups()
	sortedPrices := slices.Sorted(maps.Keys(groups))
	initialProducts := store.List()
	initialNames := store.DistinctNames()

	if err := store.UpdatePrice(sku, 15000); err != nil {
		return err
	}
	updated, found := store.Find(sku)
	if !found {
		return fmt.Errorf("updated product not found: %s", sku)
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

	// A builder keeps a failed calculation from emitting a partial report.
	// strings.Builder.Write never returns a non-nil error.
	var report strings.Builder
	fmt.Fprintf(&report, "Found: %s | %s\n", quote.SKU.String(), quote.ProductName)
	fmt.Fprintf(&report, "Products: %d\n", len(initialProducts))
	for _, product := range initialProducts {
		fmt.Fprintf(&report, "- %s | %s | %d kopecks\n",
			product.SKU().String(), product.Name(), product.UnitPriceKopecks())
	}
	fmt.Fprintf(&report, "Order total: %d kopecks\n", quote.TotalKopecks)
	for _, price := range sortedPrices {
		fmt.Fprintf(&report, "Price group %d kopecks: %d product(s)\n", price, len(groups[price]))
	}
	fmt.Fprintf(&report, "Distinct names: %v\n", initialNames)
	if cfg.ExportCSVPath != "" {
		fmt.Fprintf(&report, "Catalog exported to: %s\n", cfg.ExportCSVPath)
	}
	fmt.Fprintf(&report, "Updated price: %d kopecks\n", updated.UnitPriceKopecks())
	fmt.Fprintf(&report, "Original price: %d kopecks\n", first.UnitPriceKopecks())
	fmt.Fprintf(&report, "Original catalog count: %d\n", store.Count())
	fmt.Fprintf(&report, "Cloned catalog count: %d\n", clone.Count())
	fmt.Fprintf(&report, "Original catalog current price: %d kopecks\n", originalProduct.UnitPriceKopecks())
	fmt.Fprintf(&report, "Clone price: %d kopecks\n", cloneProduct.UnitPriceKopecks())
	fmt.Fprintf(&report, "Original quote for 2: %d kopecks\n", originalQuote.TotalKopecks)
	fmt.Fprintf(&report, "Clone quote for 2: %d kopecks\n", cloneQuote.TotalKopecks)

	if _, err := io.WriteString(output, report.String()); err != nil {
		return fmt.Errorf("write console report: %w", err)
	}
	return nil
}
