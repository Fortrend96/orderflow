package main

import (
	"flag"
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/pricing"
)

func main() {
	skuFlag := flag.String(
		"sku",
		"COFFEE01",
		"Product SKU: 8 uppercase ASCII letters or digits",
	)

	nameFlag := flag.String(
		"name",
		"Кофе",
		"Product name",
	)

	priceFlag := flag.Int64(
		"price",
		0,
		"Unit price in kopecks",
	)

	quantityFlag := flag.Int64(
		"quantity",
		1,
		"Number of items",
	)

	flag.Parse()

	product, err := catalog.NewProduct(
		*skuFlag,
		*nameFlag,
		*priceFlag,
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	total, err := pricing.CalculateLineTotal(
		product.UnitPriceKopecks,
		*quantityFlag,
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Printf("SKU: %s\n", product.SKU.String())
	fmt.Printf("Name: %s\n", product.Name)
	fmt.Printf("Name bytes: %d\n", len(product.Name))
	fmt.Printf(
		"Name runes: %d\n",
		utf8.RuneCountInString(product.Name),
	)
	fmt.Printf("Total: %d kopecks\n", total)
}
