package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/pricing"
)

func run() error {
	skuFlag := flag.String(
		"sku",
		"COFFEE01",
		"Product SKU",
	)

	nameFlag := flag.String(
		"name",
		"Кофе",
		"Product name",
	)

	priceFlag := flag.Int64(
		"price",
		12990,
		"Unit price in kopecks",
	)

	quantityFlag := flag.Int64(
		"quantity",
		3,
		"Number of items",
	)

	flag.Parse()

	store, err := catalog.NewCatalog(2)

	if err != nil {
		return err
	}

	first, err := catalog.NewProduct(
		*skuFlag,
		*nameFlag,
		*priceFlag,
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

	sku, err := catalog.ParseSKU(*skuFlag)

	if err != nil {
		return err
	}

	found, ok := store.Find(sku)

	if !ok {
		return fmt.Errorf(
			"product not found: %s",
			sku.String(),
		)
	}

	fmt.Printf(
		"Found: %s | %s\n",
		found.SKU.String(),
		found.Name,
	)

	fmt.Printf("Products: %d\n", store.Count())

	for _, product := range store.List() {
		fmt.Printf(
			"- %s | %s | %d kopecks\n",
			product.SKU.String(),
			product.Name,
			product.UnitPriceKopecks,
		)
	}

	total, err := pricing.CalculateLineTotal(
		found.UnitPriceKopecks,
		*quantityFlag,
	)

	if err != nil {
		return err
	}

	fmt.Printf("Order total: %d kopecks\n", total)

	if err := store.UpdatePrice(sku, 15000); err != nil {
		return err
	}

	updated, ok := store.Find(sku)

	if !ok {
		return fmt.Errorf("updated product not found")
	}

	fmt.Printf(
		"Updated price: %d kopecks\n",
		updated.UnitPriceKopecks,
	)

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
