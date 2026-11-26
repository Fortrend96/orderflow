package checkout

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/pricing"
)

func TestQuoteBatchChannelsOrder(t *testing.T) {
	s, lines := batchFixture(t)
	before := slices.Clone(lines)
	want, err := s.QuoteBatchConcurrent(lines, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, workers := range []int{1, 2, 3, 20} {
		for _, buffer := range []int{0, 1, 3, 20} {
			t.Run(fmt.Sprintf("workers=%d/buffer=%d", workers, buffer), func(t *testing.T) {
				t.Parallel()
				got, err := s.QuoteBatchChannels(lines, workers, buffer)
				if err != nil || !slices.Equal(got, want) {
					t.Fatalf("got %v, %v; want %v", got, err, want)
				}
				if !slices.Equal(lines, before) {
					t.Fatal("input changed")
				}
			})
		}
	}
}

func TestQuoteBatchChannelsErrors(t *testing.T) {
	s, lines := batchFixture(t)
	for _, tc := range []struct {
		workers, buffer int
		want            error
	}{
		{0, 0, ErrInvalidWorkerCount}, {-1, 0, ErrInvalidWorkerCount}, {1, -1, ErrInvalidChannelBuffer}, {0, -1, ErrInvalidWorkerCount},
	} {
		got, err := s.QuoteBatchChannels(nil, tc.workers, tc.buffer)
		if got != nil || !errors.Is(err, tc.want) {
			t.Fatalf("got %v, %v; want %v", got, err, tc.want)
		}
	}
	got, err := s.QuoteBatchChannels(nil, 2, 0)
	if got == nil || len(got) != 0 || err != nil {
		t.Fatalf("empty: %v, %v", got, err)
	}
	missing, err := catalog.ParseSKU("MISSING1")
	if err != nil {
		t.Fatal(err)
	}
	for _, buffer := range []int{0, 1, 20} {
		got, err := s.QuoteBatchChannels([]LineRequest{{missing, 1}}, 2, buffer)
		if got != nil || !errors.Is(err, catalog.ErrProductNotFound) {
			t.Fatalf("missing: %v, %v", got, err)
		}
		got, err = s.QuoteBatchChannels([]LineRequest{{lines[0].SKU, math.MaxInt64}}, 2, buffer)
		if got != nil || !errors.Is(err, pricing.ErrTotalOverflow) {
			t.Fatalf("overflow: %v, %v", got, err)
		}
		bad := slices.Clone(lines)
		bad[1].Quantity = 0
		bad[4].SKU = missing
		got, err = s.QuoteBatchChannels(bad, 3, buffer)
		var detail *QuoteError
		if got != nil || !errors.Is(err, pricing.ErrInvalidQuantity) || !errors.As(err, &detail) || detail.Quantity != 0 || detail.SKU != bad[1].SKU || !strings.HasPrefix(err.Error(), "line 1:") {
			t.Fatalf("multiple errors: %v, %v", got, err)
		}
	}
}

func TestQuoteBatchChannelsOverlapsAndJoins(t *testing.T) {
	for _, buffer := range []int{0, 2} {
		t.Run(fmt.Sprint(buffer), func(t *testing.T) {
			p := mustTestProduct(t, "COFFEE01", "Кофе", 100)
			f := &gatedFinder{product: p, entered: make(chan struct{}, 5), release: make(chan struct{})}
			s, err := NewService(f)
			if err != nil {
				t.Fatal(err)
			}
			lines := make([]LineRequest, 5)
			for i := range lines {
				lines[i] = LineRequest{p.SKU(), 1}
			}
			lines[0].Quantity = 0
			done := make(chan error, 1)
			go func() { _, err := s.QuoteBatchChannels(lines, 2, buffer); done <- err }()
			released := false
			defer func() {
				if !released {
					close(f.release)
				}
			}()
			for i := 0; i < 2; i++ {
				select {
				case <-f.entered:
				case <-time.After(5 * time.Second):
					t.Fatal("workers did not overlap")
				}
			}
			select {
			case <-done:
				t.Fatal("returned before release")
			default:
			}
			close(f.release)
			released = true
			select {
			case err := <-done:
				if !errors.Is(err, pricing.ErrInvalidQuantity) {
					t.Fatal(err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("batch did not finish")
			}
			if f.peak.Load() != 2 || f.calls.Load() != 5 || f.active.Load() != 0 {
				t.Fatalf("peak=%d calls=%d active=%d", f.peak.Load(), f.calls.Load(), f.active.Load())
			}
		})
	}
}

func ExampleService_QuoteBatchChannels() {
	store, _ := catalog.NewCatalog(1)
	p, _ := catalog.NewProduct("COFFEE01", "Кофе", 100)
	_ = store.Add(p)
	s, _ := NewService(store)
	quotes, err := s.QuoteBatchChannels([]LineRequest{{p.SKU(), 2}, {p.SKU(), 3}}, 2, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, q := range quotes {
		fmt.Println(q.TotalKopecks)
	}
	// Output:
	// 200
	// 300
}
