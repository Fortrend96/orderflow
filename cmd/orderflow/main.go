package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Fortrend96/orderflow/internal/pricing"
)

func main() {
	price := flag.Int64(
		"price",
		0,
		"Unit price in kopecks",
	)

	quantity := flag.Int64(
		"quantity",
		1,
		"Number of items",
	)

	flag.Parse()

	total, err := pricing.CalculateLineTotal(
		*price,
		*quantity,
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	rubles := total / 100
	kopecks := total % 100

	// fmt.Printf("Total: %d kopecks\n", total)
	fmt.Printf("Total: %d.%d RUB", rubles, kopecks)
}
