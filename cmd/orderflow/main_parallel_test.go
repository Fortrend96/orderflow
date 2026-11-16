package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Separate FlagSet, output buffers, catalog, and temp directory per subtest
// let these CLI scenarios run concurrently without shared mutable state.
func TestRunCLIParallelScenarios(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		label string
		want  string
	}{
		{name: "coffee", args: []string{"-sku", "COFFEE01", "-name", "Кофе", "-price", "200", "-quantity", "2"}, label: "coffee", want: "Order total: 400 kopecks"},
		{name: "bread", args: []string{"-sku", "BREAD001", "-name", "Хлеб", "-price", "125", "-quantity", "4"}, label: "bread", want: "Order total: 500 kopecks"},
		{name: "tea", args: []string{"-sku", "TEA00002", "-name", "Чай", "-price", "333", "-quantity", "3"}, label: "tea", want: "Order total: 999 kopecks"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), tc.label+".csv")
			args := append(append([]string{}, tc.args...), "-export-csv", path)
			var stdout, stderr bytes.Buffer
			status := runCLI(args, &stdout, &stderr)
			if status != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), tc.want) {
				t.Fatalf("status=%d, out=%q, err=%q", status, stdout.String(), stderr.String())
			}
			// TempDir itself registers a cleanup. Our later cleanup runs first.
			t.Cleanup(func() {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("CSV missing during cleanup: %v", err)
					return
				}
				if !strings.Contains(string(data), "sku,name,unit_price_kopecks") {
					t.Errorf("CSV header missing in %q", path)
				}
			})
		})
	}
}
