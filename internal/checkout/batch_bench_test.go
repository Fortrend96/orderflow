package checkout

import (
	"fmt"
	"testing"
	"time"

	"github.com/Fortrend96/orderflow/internal/catalog"
)

var batchSink []Quote

type delayedFinder struct {
	base ProductFinder
}

func (f delayedFinder) Find(
	sku catalog.SKU,
) (catalog.Product, bool) {
	// Модель ожидания I/O, а не реальный сетевой запрос.
	time.Sleep(time.Millisecond)

	return f.base.Find(sku)
}

func BenchmarkQuoteBatchCPU(b *testing.B) {
	s, seed := batchFixture(b)

	lines := make([]LineRequest, 1000)

	for i := range lines {
		lines[i] = seed[i%len(seed)]
	}

	benchmarkBatch(b, s, lines)
}

func BenchmarkQuoteBatchWait(b *testing.B) {
	s, seed := batchFixture(b)

	wrapped, err := NewService(
		delayedFinder{base: s.finder},
	)

	if err != nil {
		b.Fatal(err)
	}

	lines := make([]LineRequest, 16)

	for i := range lines {
		lines[i] = seed[i%len(seed)]
	}

	benchmarkBatch(b, wrapped, lines)
}

func benchmarkBatch(
	b *testing.B,
	s *Service,
	lines []LineRequest,
) {
	b.Helper()

	for _, workers := range []int{1, 2, 4, 8} {
		b.Run(
			fmt.Sprintf("workers=%d", workers),
			func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					quotes, err := s.QuoteBatchConcurrent(
						lines,
						workers,
					)

					if err != nil {
						b.Fatal(err)
					}

					batchSink = quotes
				}
			},
		)
	}
}
