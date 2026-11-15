package demo_test

import (
	"bytes"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/demo"
	"github.com/Fortrend96/orderflow/internal/pricing"
)

func exampleConfig() demo.Config {
	return demo.Config{
		SKU: "COFFEE01", Name: "Кофе",
		UnitPriceKopecks: 12990, Quantity: 3,
	}
}

func TestRunPreservesPreviousConsoleReport(t *testing.T) {
	var out bytes.Buffer
	if err := demo.Run(exampleConfig(), &out); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"Found: COFFEE01 | Кофе",
		"Products: 2",
		"- COFFEE01 | Кофе | 12990 kopecks",
		"- TEA00001 | Чай | 5500 kopecks",
		"Order total: 38970 kopecks",
		"Price group 5500 kopecks: 1 product(s)",
		"Price group 12990 kopecks: 1 product(s)",
		"Distinct names: [Кофе Чай]",
		"Updated price: 15000 kopecks",
		"Original price: 12990 kopecks",
		"Original catalog count: 2",
		"Cloned catalog count: 1",
		"Original catalog current price: 15000 kopecks",
		"Clone price: 22000 kopecks",
		"Original quote for 2: 30000 kopecks",
		"Clone quote for 2: 44000 kopecks",
	}
	if got := strings.TrimSuffix(out.String(), "\n"); got != strings.Join(want, "\n") {
		t.Fatalf("report mismatch:\n%s", got)
	}
}

func TestRunExportsOriginalPrices(t *testing.T) {
	cfg := exampleConfig()
	cfg.Name = "Кофе, зерновой"
	cfg.ExportCSVPath = filepath.Join(t.TempDir(), "products.csv")

	var out bytes.Buffer
	if err := demo.Run(cfg, &out); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(cfg.ExportCSVPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"sku", "name", "unit_price_kopecks"},
		{"COFFEE01", "Кофе, зерновой", "12990"},
		{"TEA00001", "Чай", "5500"},
	}
	if len(records) != len(want) {
		t.Fatalf("records=%v", records)
	}
	for i := range want {
		for j := range want[i] {
			if records[i][j] != want[i][j] {
				t.Fatalf("record[%d][%d]=%q, want %q", i, j, records[i][j], want[i][j])
			}
		}
	}
	if !strings.Contains(out.String(), "Catalog exported to: "+cfg.ExportCSVPath) {
		t.Fatal("report does not announce successful export")
	}
}

func TestRunErrorsDoNotPrintPartialReport(t *testing.T) {
	cases := []struct {
		name   string
		change func(*demo.Config)
		want   error
	}{
		{name: "bad SKU", change: func(c *demo.Config) { c.SKU = "bad" }, want: catalog.ErrInvalidSKU},
		{name: "bad name", change: func(c *demo.Config) { c.Name = " " }, want: catalog.ErrInvalidName},
		{name: "bad price", change: func(c *demo.Config) { c.UnitPriceKopecks = -1 }, want: catalog.ErrInvalidPrice},
		{name: "bad quantity", change: func(c *demo.Config) { c.Quantity = 0 }, want: pricing.ErrInvalidQuantity},
		{name: "duplicate SKU", change: func(c *demo.Config) { c.SKU = "TEA00001" }, want: catalog.ErrProductExists},
		{name: "export path missing", change: func(c *demo.Config) {
			c.ExportCSVPath = filepath.Join(t.TempDir(), "missing", "report.csv")
		}, want: os.ErrNotExist},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := exampleConfig()
			tc.change(&cfg)
			var out bytes.Buffer
			err := demo.Run(cfg, &out)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got err %v, want %v", err, tc.want)
			}
			if out.Len() != 0 {
				t.Fatalf("partial report on error: %q", out.String())
			}
		})
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestRunPropagatesOutputError(t *testing.T) {
	sentinel := errors.New("output disconnected")
	if err := demo.Run(exampleConfig(), failingWriter{err: sentinel}); !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want wrapped output error", err)
	}
}

func TestRunRejectsNilWriter(t *testing.T) {
	if err := demo.Run(exampleConfig(), nil); err == nil {
		t.Fatal("expected nil writer error")
	}
}
