package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/checkout"
	"github.com/Fortrend96/orderflow/internal/demo"
	"github.com/Fortrend96/orderflow/internal/pricing"
)

// runCLI is the testable boundary between command-line I/O and application code.
// Exit codes: 0 success/help, 1 operation error, 2 invalid arguments.
func runCLI(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("orderflow", flag.ContinueOnError)
	flags.SetOutput(stderr)

	sku := flags.String("sku", "COFFEE01", "Product SKU")
	name := flags.String("name", "Кофе", "Product name")
	price := flags.Int64("price", 12990, "Unit price in kopecks")
	quantity := flags.Int64("quantity", 3, "Number of items")
	export := flags.String("export-csv", "", "Optional output CSV path")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", flags.Args())
		return 2
	}

	cfg := demo.Config{
		SKU:              *sku,
		Name:             *name,
		UnitPriceKopecks: *price,
		Quantity:         *quantity,
		ExportCSVPath:    *export,
	}
	if err := demo.Run(cfg, stdout); err != nil {
		printError(stderr, err)
		return 1
	}
	return 0
}

func printError(output io.Writer, err error) {
	var quoteErr *checkout.QuoteError
	if errors.As(err, &quoteErr) {
		fmt.Fprintf(output, "failed to quote SKU %s, quantity %d\n",
			quoteErr.SKU.String(), quoteErr.Quantity)
	}

	switch {
	case errors.Is(err, catalog.ErrProductNotFound):
		fmt.Fprintln(output, "not found:", err)
	case errors.Is(err, catalog.ErrProductExists):
		fmt.Fprintln(output, "conflict:", err)
	case errors.Is(err, catalog.ErrInvalidSKU),
		errors.Is(err, catalog.ErrInvalidName),
		errors.Is(err, catalog.ErrInvalidPrice),
		errors.Is(err, pricing.ErrInvalidQuantity),
		errors.Is(err, pricing.ErrInvalidUnitPrice),
		errors.Is(err, pricing.ErrTotalOverflow):
		fmt.Fprintln(output, "invalid input:", err)
	default:
		fmt.Fprintln(output, "error:", err)
	}
}

func main() {
	// Do not call os.Exit on success: it skips deferred functions in main.
	if status := runCLI(os.Args[1:], os.Stdout, os.Stderr); status != 0 {
		os.Exit(status)
	}
}
