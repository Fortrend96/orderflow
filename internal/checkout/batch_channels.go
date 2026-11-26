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
	jobs := make(chan quoteJob, buffer)
	results := make(chan quoteResult, buffer)

	var wg sync.WaitGroup
	wg.Add(workers + 1)
	go func() {
		defer wg.Done()
		sendQuoteJobs(lines, jobs)
	}()
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			s.processQuoteJobs(jobs, results)
		}()
	}

	// Только вызывающая goroutine записывает в итоговый slice.
	firstErrorIndex := len(lines)
	var firstError error
	for i := 0; i < len(lines); i++ {
		result := <-results
		quotes[result.index] = result.quote
		if result.err != nil && result.index < firstErrorIndex {
			firstErrorIndex = result.index
			firstError = result.err
		}
	}
	wg.Wait()
	if firstError != nil {
		return nil, fmt.Errorf("line %d: %w", firstErrorIndex, firstError)
	}
	return quotes, nil
}

// sendQuoteJobs — единственный отправитель и владелец закрытия jobs.
func sendQuoteJobs(lines []LineRequest, jobs chan<- quoteJob) {
	defer close(jobs)
	for i, line := range lines {
		jobs <- quoteJob{index: i, line: line}
	}
}

// processQuoteJobs получает задания и отправляет ровно один результат на задание.
func (s *Service) processQuoteJobs(jobs <-chan quoteJob, results chan<- quoteResult) {
	for job := range jobs {
		quote, err := s.Quote(job.line.SKU, job.line.Quantity)
		results <- quoteResult{index: job.index, quote: quote, err: err}
	}
}
