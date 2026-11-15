package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/checkout"
)

func TestRunCLIDefault(t *testing.T) {
	var stdout, stderr bytes.Buffer
	status := runCLI(nil, &stdout, &stderr)
	if status != 0 || stderr.Len() != 0 {
		t.Fatalf("status=%d stderr=%q", status, stderr.String())
	}
	for _, text := range []string{
		"Order total: 38970 kopecks",
		"Distinct names: [Кофе Чай]",
		"Original quote for 2: 30000 kopecks",
	} {
		if !strings.Contains(stdout.String(), text) {
			t.Fatalf("missing %q in stdout:\n%s", text, stdout.String())
		}
	}
}

func TestRunCLIParsesFlagsAndExports(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	var stdout, stderr bytes.Buffer
	status := runCLI([]string{
		"-sku", "BREAD001", "-name", "Хлеб", "-price", "500",
		"-quantity", "4", "-export-csv", path,
	}, &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status=%d stderr=%q", status, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Order total: 2000 kopecks") {
		t.Fatalf("wrong price calculation: %q", stdout.String())
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "BREAD001,Хлеб,500") {
		t.Fatalf("unexpected CSV: %q", string(contents))
	}
}

func TestRunCLIClassifiesApplicationErrors(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		expected string
	}{
		{name: "quantity", args: []string{"-quantity", "0"}, expected: "invalid input: quote SKU COFFEE01"},
		{name: "invalid SKU", args: []string{"-sku", "bad"}, expected: "invalid input: product SKU"},
		{name: "duplicate", args: []string{"-sku", "TEA00001"}, expected: "conflict:"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := runCLI(tc.args, &out, &errOut)
			if code != 1 || !strings.Contains(errOut.String(), tc.expected) || out.Len() != 0 {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, out.String(), errOut.String())
			}
		})
	}
}

func TestRunCLIInvalidArguments(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		contains string
	}{
		{name: "unknown flag", args: []string{"-unknown"}, contains: "flag provided but not defined"},
		{name: "bad int", args: []string{"-quantity", "NaN"}, contains: "invalid value"},
		{name: "positional", args: []string{"unexpected"}, contains: "unexpected positional arguments"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := runCLI(tc.args, &out, &errOut)
			if code != 2 || !strings.Contains(errOut.String(), tc.contains) || out.Len() != 0 {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, out.String(), errOut.String())
			}
		})
	}
}

func TestRunCLIHelp(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := runCLI([]string{"-h"}, &out, &errOut); code != 0 || !strings.Contains(errOut.String(), "-export-csv") {
		t.Fatalf("help code=%d stdout=%q stderr=%q", code, out.String(), errOut.String())
	}
}

func TestPrintErrorWrapsNotFoundAndQuoteContext(t *testing.T) {
	sku, err := catalog.ParseSKU("MISSING1")
	if err != nil {
		t.Fatal(err)
	}
	err = &checkout.QuoteError{SKU: sku, Quantity: 2, Err: catalog.ErrProductNotFound}
	err = errors.Join(err)
	var out bytes.Buffer
	printError(&out, err)
	if !strings.Contains(out.String(), "failed to quote SKU MISSING1, quantity 2") ||
		!strings.Contains(out.String(), "not found:") {
		t.Fatalf("unexpected error output: %q", out.String())
	}
}
