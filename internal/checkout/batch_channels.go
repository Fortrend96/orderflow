package checkout

import (
	"errors"
	"fmt"
	"sync"
)

var ErrInvalidChannelBuffer = errors.New("invalid channel buffer")

type quoteJob struct {
	index int
	line  LineRequest
}

type quoteResult struct {
	index int
	quote Quote
	err   error
}

// QuoteBatchChannels рассчитывает все позиции через каналы, сохраняя порядок.
// Finder должен поддерживать конкурентные вызовы Find, а вход нельзя менять
// до возврата метода. При ошибке возвращается nil и ошибка с наименьшим индексом.
// buffer задаёт ёмкость обоих каналов, ограниченную количеством позиций.
func (s *Service) QuoteBatchChannels(lines []LineRequest, workers, buffer int) ([]Quote, error) {
	if workers <= 0 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidWorkerCount, workers)
	}
	if buffer < 0 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidChannelBuffer, buffer)
	}
	quotes := make([]Quote, len(lines))
	if len(lines) == 0 {
		return quotes, nil
	}
	workers = min(workers, len(lines))
	buffer = min(buffer, len(lines))

	// nil отключает остановку: публичный метод обрабатывает весь пакет.
	results := s.streamQuoteResults(lines, workers, buffer, nil)

	// Только вызывающая goroutine записывает в итоговый slice.
	firstErrorIndex := len(lines)
	var firstError error

	for result := range results {
		quotes[result.index] = result.quote
		if result.err != nil && result.index < firstErrorIndex {
			firstErrorIndex = result.index
			firstError = result.err
		}
	}

	if firstError != nil {
		return nil, fmt.Errorf("line %d: %w", firstErrorIndex, firstError)
	}

	return quotes, nil
}

// streamQuoteResults запускает поток. workers > 0, buffer >= 0.
// Получатель читает до закрытия либо закрывает stop и ждёт закрытия results.
// Stop не прерывает уже выполняющийся Find.
func (s *Service) streamQuoteResults(lines []LineRequest, workers, buffer int, stop <-chan struct{}) <-chan quoteResult {
	jobs := make(chan quoteJob, buffer)
	results := make(chan quoteResult, buffer)

	var wg sync.WaitGroup
	wg.Add(workers + 1)

	go func() {
		defer wg.Done()
		sendQuoteJobs(lines, jobs, stop)
	}()

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			s.processQuoteJobs(jobs, results, stop)
		}()
	}

	// Единственный владелец закрытия results ждёт всех отправителей.
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func sendQuoteJobs(lines []LineRequest, jobs chan<- quoteJob, stop <-chan struct{}) {
	defer close(jobs)

	for i, line := range lines {
		select {
		case <-stop:
			return

		case jobs <- quoteJob{index: i, line: line}:
		}
	}
}

func (s *Service) processQuoteJobs(jobs <-chan quoteJob, results chan<- quoteResult, stop <-chan struct{}) {
	for {
		// Не начинаем следующую итерацию, если остановка уже запрошена.
		select {
		case <-stop:
			return

		default:
		}

		select {
		case <-stop:
			return

		case job, ok := <-jobs:
			if !ok {
				return
			}

			quote, err := s.Quote(job.line.SKU, job.line.Quantity)

			select {
			case <-stop:
				return

			case results <- quoteResult{index: job.index, quote: quote, err: err}:
			}
		}
	}
}
