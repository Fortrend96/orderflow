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
		first.UnitPriceKopecks,
		*quantityFlag,
	)

	if err != nil {
		return err
	}

	fmt.Printf("Order total: %d kopecks\n", total)

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
