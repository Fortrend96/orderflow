package demo

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/checkout"
)

func TestWriteReportUsesValuesAndSortedGroups(t *testing.T) {
	coffee, err := catalog.NewProduct(
		"COFFEE01",
		"Кофе",
		12990,
	)
	if err != nil {
		t.Fatal(err)
	}

	tea, err := catalog.NewProduct(
		"TEA00001",
		"Чай",
		5500,
	)
	if err != nil {
		t.Fatal(err)
	}

	data := reportData{
		quote: checkout.Quote{
			SKU:              coffee.SKU(),
			ProductName:      coffee.Name(),
			Quantity:         3,
			UnitPriceKopecks: 12990,
			TotalKopecks:     38970,
		},
		products: []catalog.Product{
			coffee,
			tea,
		},
		groups: map[int64][]catalog.Product{
			12990: {coffee},
			5500:  {tea},
		},
		names: []string{
			"Кофе",
			"Чай",
		},
		originalPrice:        12990,
		updatedPrice:         15000,
		originalCount:        2,
		cloneCount:           1,
		originalCurrentPrice: 15000,
		clonedPrice:          22000,
		originalQuoteForTwo:  30000,
		clonedQuoteForTwo:    44000,
	}

	var out bytes.Buffer

	if err := writeReport(&out, data); err != nil {
		t.Fatal(err)
	}

	text := out.String()

	first := strings.Index(text, "Price group 5500")
	second := strings.Index(text, "Price group 12990")

	if first < 0 || second < 0 || first >= second {
		t.Fatalf(
			"price groups were not sorted: %q",
			text,
		)
	}

	for _, line := range []string{
		"Order total: 38970 kopecks",
		"Original price: 12990 kopecks",
		"Updated price: 15000 kopecks",
		"Clone price: 22000 kopecks",
	} {
		if !strings.Contains(text, line) {
			t.Fatalf(
				"missing %q in report",
				line,
			)
		}
	}
}

type partialErrorWriter struct {
	cause error
}

func (w partialErrorWriter) Write(data []byte) (int, error) {
	if len(data) > 5 {
		return 5, w.cause
	}

	return 0, w.cause
}

func TestWriteReportPreservesOutputError(t *testing.T) {
	original := errors.New("output disconnected")

	err := writeReport(
		partialErrorWriter{
			cause: original,
		},
		reportData{},
	)

	if !errors.Is(err, original) {
		t.Fatalf(
			"writeReport: %v; want wrapped output error",
			err,
		)
	}
}
