// Package catalogcsv exports validated catalog products to CSV files.
package catalogcsv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Fortrend96/orderflow/internal/catalog"
)

// WriteCSV writes a header and catalog products to w.
// It does not close w: ownership remains with the caller.
func WriteCSV(w io.Writer, products []catalog.Product) error {
	writer := csv.NewWriter(w)

	if err := writer.Write([]string{"sku", "name", "unit_price_kopecks"}); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	for _, product := range products {
		record := []string{
			product.SKU().String(),
			product.Name(),
			strconv.FormatInt(product.UnitPriceKopecks(), 10),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write CSV product %s: %w", product.SKU(), err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush CSV: %w", err)
	}
	return nil
}

// ExportFile writes CSV to a temporary file in the destination directory,
// then renames it into place. Failure leaves any existing destination unchanged
// before rename. Rename replacement/atomicity guarantees are OS-specific.
func ExportFile(path string, products []catalog.Product) error {
	return writeFile(path, func(w io.Writer) error {
		return WriteCSV(w, products)
	})
}

// writeFile owns the temporary file. write only borrows the io.Writer.
// The callback is injected so tests can trigger write errors and panics.
func writeFile(path string, write func(io.Writer) error) (err error) {
	if path == "" {
		return errors.New("CSV path must not be empty")
	}

	dir := filepath.Dir(path)
	pattern := "." + filepath.Base(path) + ".tmp-*"
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return fmt.Errorf("create temporary CSV file: %w", err)
	}

	tempPath := file.Name()
	closed := false
	published := false

	// Always release the file, including during panic unwinding.
	// A failed export also removes the temporary file.
	defer func() {
		if !closed {
			if closeErr := file.Close(); closeErr != nil {
				err = errors.Join(err, fmt.Errorf("close temporary CSV file: %w", closeErr))
			}
		}
		if !published {
			if removeErr := os.Remove(tempPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				err = errors.Join(err, fmt.Errorf("remove temporary CSV file: %w", removeErr))
			}
		}
	}()

	if err := write(file); err != nil {
		return fmt.Errorf("write CSV: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync temporary CSV file: %w", err)
	}

	// Close explicitly before rename, so the close error is not discarded.
	closeErr := file.Close()
	closed = true
	if closeErr != nil {
		return fmt.Errorf("close temporary CSV file: %w", closeErr)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("rename temporary CSV file: %w", err)
	}
	published = true
	return nil
}
