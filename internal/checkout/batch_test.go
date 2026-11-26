package checkout

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Fortrend96/orderflow/internal/catalog"
	"github.com/Fortrend96/orderflow/internal/pricing"
)

func batchFixture(t testing.TB) (*Service, []LineRequest) {
	t.Helper()

	store, err := catalog.NewCatalog(2)
	if err != nil {
		t.Fatal(err)
	}

	lines := make([]LineRequest, 0, 5)

	for _, raw := range []string{"COFFEE01", "TEA00001"} {
		p, err := catalog.NewProduct(raw, "Товар", 100)
		if err != nil {
			t.Fatal(err)
		}

		if err := store.Add(p); err != nil {
			t.Fatal(err)
		}

		lines = append(lines, LineRequest{
			SKU:      p.SKU(),
			Quantity: 2,
		})
	}

	lines = append(
		lines,
		LineRequest{lines[0].SKU, 3},
		LineRequest{lines[1].SKU, 4},
		LineRequest{lines[0].SKU, 5},
	)

	service, err := NewService(store.Clone())
	if err != nil {
		t.Fatal(err)
	}

	return service, lines
}

func TestQuoteBatchConcurrentOrder(t *testing.T) {
	s, lines := batchFixture(t)

	before := slices.Clone(lines)

	want := make([]Quote, len(lines))

	for i, line := range lines {
		q, err := s.Quote(line.SKU, line.Quantity)
		if err != nil {
			t.Fatal(err)
		}

		want[i] = q
	}

	for _, workers := range []int{1, 2, 3, 5, 20} {
		t.Run(fmt.Sprint(workers), func(t *testing.T) {
			t.Parallel()

			got, err := s.QuoteBatchConcurrent(lines, workers)

			if err != nil || !slices.Equal(got, want) {
				t.Fatalf(
					"got %v, %v; want %v",
					got,
					err,
					want,
				)
			}

			if !slices.Equal(lines, before) {
				t.Fatal("input changed")
			}
		})
	}
}

func TestQuoteBatchConcurrentErrors(t *testing.T) {
	s, lines := batchFixture(t)

	for _, workers := range []int{0, -1} {
		if got, err := s.QuoteBatchConcurrent(nil, workers); got != nil ||
			!errors.Is(err, ErrInvalidWorkerCount) {
			t.Fatalf("got %v, %v", got, err)
		}
	}

	if got, err := s.QuoteBatchConcurrent(nil, 2); got == nil ||
		len(got) != 0 || err != nil {
		t.Fatalf("empty: %v, %v", got, err)
	}

	missing, err := catalog.ParseSKU("MISSING1")
	if err != nil {
		t.Fatal(err)
	}

	lines[1].Quantity = 0
	lines[4].SKU = missing

	for _, workers := range []int{1, 2, 20} {
		got, err := s.QuoteBatchConcurrent(lines, workers)

		var detail *QuoteError

		if got != nil ||
			!errors.Is(err, pricing.ErrInvalidQuantity) ||
			!errors.As(err, &detail) ||
			detail.SKU != lines[1].SKU ||
			detail.Quantity != 0 ||
			!strings.HasPrefix(err.Error(), "line 1:") {
			t.Fatalf(
				"workers=%d: got %v, %v",
				workers,
				got,
				err,
			)
		}
	}

	got, err := s.QuoteBatchConcurrent(
		[]LineRequest{{missing, 1}},
		2,
	)

	if got != nil || !errors.Is(err, catalog.ErrProductNotFound) {
		t.Fatalf("missing: %v, %v", got, err)
	}

	got, err = s.QuoteBatchConcurrent(
		[]LineRequest{{lines[0].SKU, math.MaxInt64}},
		2,
	)

	if got != nil || !errors.Is(err, pricing.ErrTotalOverflow) {
		t.Fatalf("overflow: %v, %v", got, err)
	}
}

// Каналы здесь нужны только для управляемого теста,
// без ожиданий через Sleep.
type gatedFinder struct {
	product catalog.Product
	entered chan struct{}
	release chan struct{}
	active  atomic.Int32
	peak    atomic.Int32
	calls   atomic.Int32
}

func (f *gatedFinder) Find(
	sku catalog.SKU,
) (catalog.Product, bool) {
	f.calls.Add(1)

	n := f.active.Add(1)

	defer f.active.Add(-1)

	for old := f.peak.Load(); n > old; old = f.peak.Load() {
		if f.peak.CompareAndSwap(old, n) {
			break
		}
	}

	f.entered <- struct{}{}

	<-f.release

	return f.product, sku == f.product.SKU()
}

func TestQuoteBatchConcurrentOverlapsAndJoins(t *testing.T) {
	p := mustTestProduct(t, "COFFEE01", "Кофе", 100)

	f := &gatedFinder{
		product: p,
		entered: make(chan struct{}, 5),
		release: make(chan struct{}),
	}

	s, err := NewService(f)
	if err != nil {
		t.Fatal(err)
	}

	lines := make([]LineRequest, 5)

	for i := range lines {
		lines[i] = LineRequest{p.SKU(), 1}
	}

	// Даже при ошибке все позиции должны завершиться.
	lines[0].Quantity = 0

	done := make(chan error, 1)

	go func() {
		_, err := s.QuoteBatchConcurrent(lines, 2)

		done <- err
	}()

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
			t.Fatal("two workers did not enter")
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

	if f.peak.Load() != 2 ||
		f.calls.Load() != 5 ||
		f.active.Load() != 0 {
		t.Fatalf(
			"peak=%d calls=%d active=%d",
			f.peak.Load(),
			f.calls.Load(),
			f.active.Load(),
		)
	}
}

func ExampleService_QuoteBatchConcurrent() {
	store, _ := catalog.NewCatalog(1)

	p, _ := catalog.NewProduct(
		"COFFEE01",
		"Кофе",
		100,
	)

	_ = store.Add(p)

	s, _ := NewService(store.Clone())

	quotes, err := s.QuoteBatchConcurrent(
		[]LineRequest{
			{p.SKU(), 2},
			{p.SKU(), 3},
		},
		2,
	)

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
