package demo

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/checkout"
)

// reportData holds a snapshot of a completed demo run. It contains values,
// not a pointer to the mutable Catalog that is updated during the scenario.
type reportData struct {
	quote                checkout.Quote
	products             []catalog.Product
	groups               map[int64][]catalog.Product
	names                []string
	exportPath           string
	originalPrice        int64
	updatedPrice         int64
	originalCount        int
	cloneCount           int
	originalCurrentPrice int64
	clonedPrice          int64
	originalQuoteForTwo  int64
	clonedQuoteForTwo    int64
}

// writeReport formats the result after the demo has finished its operations.
// It performs a single write to the supplied writer and preserves its error.
func writeReport(output io.Writer, data reportData) error {
	var report strings.Builder

	fmt.Fprintf(
		&report,
		"Found: %s | %s\n",
		data.quote.SKU.String(),
		data.quote.ProductName,
	)

	fmt.Fprintf(
		&report,
		"Products: %d\n",
		len(data.products),
	)

	for _, product := range data.products {
		fmt.Fprintf(
			&report,
			"- %s | %s | %d kopecks\n",
			product.SKU().String(),
			product.Name(),
			product.UnitPriceKopecks(),
		)
	}

	fmt.Fprintf(
		&report,
		"Order total: %d kopecks\n",
		data.quote.TotalKopecks,
	)

	for _, price := range slices.Sorted(maps.Keys(data.groups)) {
		fmt.Fprintf(
			&report,
			"Price group %d kopecks: %d product(s)\n",
			price,
			len(data.groups[price]),
		)
	}

	fmt.Fprintf(
		&report,
		"Distinct names: %v\n",
		data.names,
	)

	if data.exportPath != "" {
		fmt.Fprintf(
			&report,
			"Catalog exported to: %s\n",
			data.exportPath,
		)
	}

	fmt.Fprintf(
		&report,
		"Updated price: %d kopecks\n",
		data.updatedPrice,
	)

	fmt.Fprintf(
		&report,
		"Original price: %d kopecks\n",
		data.originalPrice,
	)

	fmt.Fprintf(
		&report,
		"Original catalog count: %d\n",
		data.originalCount,
	)

	fmt.Fprintf(
		&report,
		"Cloned catalog count: %d\n",
		data.cloneCount,
	)

	fmt.Fprintf(
		&report,
		"Original catalog current price: %d kopecks\n",
		data.originalCurrentPrice,
	)

	fmt.Fprintf(
		&report,
		"Clone price: %d kopecks\n",
		data.clonedPrice,
	)

	fmt.Fprintf(
		&report,
		"Original quote for 2: %d kopecks\n",
		data.originalQuoteForTwo,
	)

	fmt.Fprintf(
		&report,
		"Clone quote for 2: %d kopecks\n",
		data.clonedQuoteForTwo,
	)

	if _, err := io.WriteString(output, report.String()); err != nil {
		return fmt.Errorf("write console report: %w", err)
	}

	return nil
}
