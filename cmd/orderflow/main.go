package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/catalogcsv"
	"github.com/Fortrend96/orderflow/internal/checkout"
	"github.com/Fortrend96/orderflow/internal/pricing"
)

func run() error {
	skuFlag := flag.String("sku", "COFFEE01", "Product SKU")
	nameFlag := flag.String("name", "Кофе", "Product name")
	priceFlag := flag.Int64("price", 12990, "Unit price in kopecks")
	quantityFlag := flag.Int64("quantity", 3, "Number of items")
	exportFlag := flag.String("export-csv", "", "Optional output CSV path")
	flag.Parse()

	store, err := catalog.NewCatalog(2)
	if err != nil {
		return err
	}

	first, err := catalog.NewProduct(*skuFlag, *nameFlag, *priceFlag)
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

	sku, err := catalog.ParseSKU(*skuFlag)
	if err != nil {
		return err
	}

	// Composition root: pass a concrete catalog to a service requiring only ProductFinder.
	orderService, err := checkout.NewService(store)
	if err != nil {
		return err
	}

	quote, err := orderService.Quote(sku, *quantityFlag)
	if err != nil {
		return err
	}

	fmt.Printf("Found: %s | %s\n", quote.SKU.String(), quote.ProductName)
	fmt.Printf("Products: %d\n", store.Count())
	for _, product := range store.List() {
		fmt.Printf(
			"- %s | %s | %d kopecks\n",
			product.SKU().String(), product.Name(), product.UnitPriceKopecks(),
		)
	}
	fmt.Printf("Order total: %d kopecks\n", quote.TotalKopecks)

	if *exportFlag != "" {
		if err := catalogcsv.ExportFile(*exportFlag, store.List()); err != nil {
			return fmt.Errorf("export catalog to %q: %w", *exportFlag, err)
		}
		fmt.Printf("Catalog exported to: %s\n", *exportFlag)
	}

	if err := store.UpdatePrice(sku, 15000); err != nil {
		return err
	}
	updated, ok := store.Find(sku)
	if !ok {
		return fmt.Errorf("updated product not found")
	}
	fmt.Printf("Updated price: %d kopecks\n", updated.UnitPriceKopecks())
	fmt.Printf("Original price: %d kopecks\n", first.UnitPriceKopecks())

	clone := store.Clone()
	if err := clone.UpdatePrice(sku, 22000); err != nil {
		return err
	}
	teaSKU, err := catalog.ParseSKU("TEA00001")
	if err != nil {
		return err
	}
	if !clone.Remove(teaSKU) {
		return fmt.Errorf("could not remove tea from clone")
	}

	originalProduct, found := store.Find(sku)
	if !found {
		return fmt.Errorf("original catalog lost product")
	}
	cloneProduct, found := clone.Find(sku)
	if !found {
		return fmt.Errorf("cloned catalog lost product")
	}
	fmt.Printf("Original catalog count: %d\n", store.Count())
	fmt.Printf("Cloned catalog count: %d\n", clone.Count())
	fmt.Printf("Original catalog current price: %d kopecks\n", originalProduct.UnitPriceKopecks())
	fmt.Printf("Clone price: %d kopecks\n", cloneProduct.UnitPriceKopecks())

	// The same service type works with a different injected catalog.
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
	fmt.Printf("Original quote for 2: %d kopecks\n", originalQuote.TotalKopecks)
	fmt.Printf("Clone quote for 2: %d kopecks\n", cloneQuote.TotalKopecks)
	return nil
}

func main() {
	if err := run(); err != nil {
		var quoteErr *checkout.QuoteError
		if errors.As(err, &quoteErr) {
			fmt.Fprintf(os.Stderr, "failed to quote SKU %s, quantity %d\n", quoteErr.SKU.String(), quoteErr.Quantity)
		}

		switch {
		case errors.Is(err, catalog.ErrProductNotFound):
			fmt.Fprintln(os.Stderr, "not found:", err)
		case errors.Is(err, catalog.ErrInvalidSKU),
			errors.Is(err, catalog.ErrInvalidName),
			errors.Is(err, catalog.ErrInvalidPrice),
			errors.Is(err, pricing.ErrInvalidQuantity),
			errors.Is(err, pricing.ErrInvalidUnitPrice),
			errors.Is(err, pricing.ErrTotalOverflow):
			fmt.Fprintln(os.Stderr, "invalid input:", err)
		default:
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		os.Exit(1)
	}
}
