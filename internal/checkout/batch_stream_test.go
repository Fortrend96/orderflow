package checkout

import (
	"fmt"
	"testing"
	"time"
)

func TestQuoteStreamClosesAfterAllResults(t *testing.T) {
	s, lines := batchFixture(t)

	for _, buffer := range []int{0, 1, 8} {
		t.Run(fmt.Sprint(buffer), func(t *testing.T) {
			t.Parallel()

			for _, input := range [][]LineRequest{nil, lines} {
				checkCompleteStream(t, s, input, buffer)
			}
		})
	}
}

func checkCompleteStream(
	t *testing.T,
	s *Service,
	lines []LineRequest,
	buffer int,
) {
	t.Helper()

	stop := make(chan struct{})
	defer close(stop)

	results := s.streamQuoteResults(lines, 3, buffer, stop)

	seen := make([]bool, len(lines))
	count := 0

	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()

	for {
		select {
		case r, ok := <-results:
			if !ok {
				if count != len(lines) {
					t.Fatalf(
						"count=%d, want %d",
						count,
						len(lines),
					)
				}

				return
			}

			if r.index < 0 ||
				r.index >= len(lines) ||
				seen[r.index] ||
				r.err != nil {
				t.Fatalf("unexpected result: %+v", r)
			}

			line := lines[r.index]

			want, err := s.Quote(
				line.SKU,
				line.Quantity,
			)

			if err != nil || r.quote != want {
				t.Fatalf(
					"quote=%+v want=%+v err=%v",
					r.quote,
					want,
					err,
				)
			}

			seen[r.index] = true
			count++

		case <-deadline.C:
			t.Fatal("results did not close")
		}
	}
}

func TestQuoteStreamStopWithoutConsumer(t *testing.T) {
	p := mustTestProduct(t, "COFFEE01", "Кофе", 100)

	f := &gatedFinder{
		product: p,
		entered: make(chan struct{}, 1),
		release: make(chan struct{}),
	}

	s, err := NewService(f)
	if err != nil {
		t.Fatal(err)
	}

	lines := []LineRequest{
		{p.SKU(), 1},
		{p.SKU(), 2},
		{p.SKU(), 3},
	}

	stop := make(chan struct{})

	results := s.streamQuoteResults(lines, 1, 0, stop)

	stopped, released := false, false

	defer func() {
		if !stopped {
			close(stop)
		}
		if !released {
			close(f.release)
		}
	}()

	select {
	case <-f.entered:

	case <-time.After(5 * time.Second):
		t.Fatal("worker did not enter Find")
	}

	close(stop)
	stopped = true

	// Сигнал остановки не прерывает Find, пока он ожидает release.
	select {
	case <-results:
		t.Fatal("stream finished while Find was blocked")

	default:
	}

	close(f.release)
	released = true

	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()

	for {
		select {
		case _, ok := <-results:
			// После stop допустим последний результат:
			// select не имеет приоритета.
			if !ok {
				if f.active.Load() != 0 ||
					f.calls.Load() != 1 {
					t.Fatalf(
						"active=%d calls=%d",
						f.active.Load(),
						f.calls.Load(),
					)
				}

				return
			}

		case <-deadline.C:
			t.Fatal("stream did not stop")
		}
	}
}

func TestQuoteStreamPreclosedStop(t *testing.T) {
	s, lines := batchFixture(t)

	stop := make(chan struct{})
	close(stop)

	results := s.streamQuoteResults(lines, 3, 2, stop)

	select {
	case _, ok := <-results:
		if ok {
			t.Fatal("unexpected result with preclosed stop")
		}

	case <-time.After(5 * time.Second):
		t.Fatal("stream did not close")
	}
}

func TestQuoteProducerStopsWithoutReceiver(t *testing.T) {
	_, lines := batchFixture(t)

	stop := make(chan struct{})
	jobs := make(chan quoteJob)
	done := make(chan struct{})

	go func() {
		defer close(done)

		sendQuoteJobs(lines, jobs, stop)
	}()

	// Получаем первое задание; дальше получателя нет.
	select {
	case j := <-jobs:
		if j.index != 0 {
			close(stop)
			t.Fatalf("index=%d", j.index)
		}

	case <-time.After(5 * time.Second):
		close(stop)
		t.Fatal("no first job")
	}

	close(stop)

	select {
	case <-done:

	case <-time.After(5 * time.Second):
		t.Fatal("producer did not stop")
	}

	if _, ok := <-jobs; ok {
		t.Fatal("jobs was not closed")
	}
}

func TestQuoteWorkerStopsWithoutReceiver(t *testing.T) {
	p := mustTestProduct(t, "COFFEE01", "Кофе", 100)

	f := &gatedFinder{
		product: p,
		entered: make(chan struct{}, 1),
		release: make(chan struct{}),
	}

	s, err := NewService(f)
	if err != nil {
		t.Fatal(err)
	}

	jobs := make(chan quoteJob, 1)

	jobs <- quoteJob{
		index: 0,
		line: LineRequest{
			p.SKU(),
			1,
		},
	}

	close(jobs)

	results := make(chan quoteResult)
	stop := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)

		s.processQuoteJobs(jobs, results, stop)
	}()

	select {
	case <-f.entered:

	case <-time.After(5 * time.Second):
		close(stop)
		close(f.release)
		t.Fatal("worker did not enter Find")
	}

	close(f.release)
	close(stop)

	// Здесь никто не читает results. Worker обязан завершиться по stop.
	select {
	case <-done:

	case <-time.After(5 * time.Second):
		t.Fatal("worker blocked on results")
	}
}
