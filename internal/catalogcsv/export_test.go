package catalogcsv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Fortrend96/orderflow/internal/catalog"
)

func testProducts(t *testing.T) []catalog.Product {
	t.Helper()
	coffee, err := catalog.NewProduct("COFFEE01", "Кофе, зерновой", 12990)
	if err != nil {
		t.Fatal(err)
	}
	tea, err := catalog.NewProduct("TEA00001", "Чай", 5500)
	if err != nil {
		t.Fatal(err)
	}
	return []catalog.Product{coffee, tea}
}

func TestWriteCSV_UnicodeAndEscaping(t *testing.T) {
	var out bytes.Buffer
	if err := WriteCSV(&out, testProducts(t)); err != nil {
		t.Fatal(err)
	}
	rawCSV := out.String()
	records, err := csv.NewReader(strings.NewReader(rawCSV)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"sku", "name", "unit_price_kopecks"},
		{"COFFEE01", "Кофе, зерновой", "12990"},
		{"TEA00001", "Чай", "5500"},
	}
	if len(records) != len(want) {
		t.Fatalf("records: got %d, want %d", len(records), len(want))
	}
	for i := range records {
		for j := range records[i] {
			if records[i][j] != want[i][j] {
				t.Fatalf("record[%d][%d]=%q, want %q", i, j, records[i][j], want[i][j])
			}
		}
	}
	if !strings.Contains(rawCSV, `"Кофе, зерновой"`) {
		t.Fatal("comma in name must be quoted in CSV")
	}
}

func TestWriteCSV_EmptyCatalogStillHasHeader(t *testing.T) {
	var out bytes.Buffer
	if err := WriteCSV(&out, nil); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "sku,name,unit_price_kopecks\n" {
		t.Fatalf("got %q", got)
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write(_ []byte) (int, error) { return 0, w.err }

func TestWriteCSV_FlushErrorIsReturned(t *testing.T) {
	sentinel := errors.New("disk unavailable")
	err := WriteCSV(failingWriter{err: sentinel}, testProducts(t))
	if !errors.Is(err, sentinel) {
		t.Fatalf("error %v does not contain sentinel", err)
	}
}

func assertNoTemporaryFiles(t *testing.T, dir string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, ".catalog.csv.tmp-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) > 0 {
		t.Fatalf("temporary files remain: %v", matches)
	}
}

func TestExportFile_SuccessAndReplacement(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.csv")
	if err := os.WriteFile(path, []byte("old data"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ExportFile(path, testProducts(t)); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "COFFEE01") || strings.Contains(string(data), "old data") {
		t.Fatalf("unexpected exported data: %q", data)
	}
	assertNoTemporaryFiles(t, dir)
}

func TestExportFile_InvalidDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "catalog.csv")
	if err := ExportFile(path, testProducts(t)); err == nil {
		t.Fatal("expected error")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("destination should not exist, got %v", err)
	}
}

func TestExportFile_EmptyPath(t *testing.T) {
	if err := ExportFile("", nil); err == nil {
		t.Fatal("expected empty-path error")
	}
}

func TestWriteFile_CallbackErrorPreservesPreviousFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.csv")
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("write failed")
	err := writeFile(path, func(w io.Writer) error {
		if _, err := w.Write([]byte("partial")); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped callback error, got %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Fatalf("destination changed on failure: %q", data)
	}
	assertNoTemporaryFiles(t, dir)
}

func TestWriteFile_RenameFailureRemovesTemp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.csv")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	err := ExportFile(path, testProducts(t))
	if err == nil {
		t.Fatal("expected rename error for destination directory")
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		t.Fatalf("destination directory lost, err %v", err)
	}
	assertNoTemporaryFiles(t, dir)
}

func TestWriteFile_DeferCleanupRunsDuringPanic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.csv")
	const panicValue = "bad callback"

	func() {
		defer func() {
			if got := recover(); got != panicValue {
				t.Errorf("recover=%v, want %q", got, panicValue)
			}
		}()
		_ = writeFile(path, func(w io.Writer) error {
			_, _ = w.Write([]byte("partial"))
			panic(panicValue)
		})
	}()

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected destination file, err=%v", err)
	}
	assertNoTemporaryFiles(t, dir)
}

func TestDeferOrderAndArguments(t *testing.T) {
	var calls []int
	func() {
		for i := 1; i <= 3; i++ {
			defer func(n int) { calls = append(calls, n) }(i)
		}
	}()
	want := []int{3, 2, 1}
	for i, v := range want {
		if calls[i] != v {
			t.Fatalf("calls[%d]=%d, want %d", i, calls[i], v)
		}
	}
}
