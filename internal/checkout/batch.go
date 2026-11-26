package checkout

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Fortrend96/orderflow/internal/catalog"
)

var ErrInvalidWorkerCount = errors.New("Invalid worker count")

// LineRequest описывает одну позицию заказа.
type LineRequest struct {
	SKU      catalog.SKU
	Quantity int64
}

// QuoteBatchConcurrent рассчитывает позиции в исходном порядке.
// Finder должен поддерживать одновременное чтение; вход нельзя менять
// до возврата метода. Все позиции обрабатываются даже при ошибке.
// При ошибке возвращается nil и причина с наименьшим индексом позиции.
func (s *Service) QuoteBatchConcurrent(
	lines []LineRequest,
	workers int,
) ([]Quote, error) {
	if workers <= 0 {
		return nil, fmt.Errorf(
			"%w: %d",
			ErrInvalidWorkerCount,
			workers,
		)
	}

	quotes := make([]Quote, len(lines))

	if len(lines) == 0 {
		return quotes, nil
	}

	workers = min(workers, len(lines))
	errs := make([]error, len(lines))

	// Каждый исполнитель записывает только в собственный диапазон.
	calculate := func(start, end int) {
		for i := start; i < end; i++ {
			line := lines[i]

			quotes[i], errs[i] = s.Quote(line.SKU, line.Quantity)
		}
	}

	if workers == 1 {
		calculate(0, len(lines))
	} else {
		var wg sync.WaitGroup

		wg.Add(workers)

		base, extra := len(lines)/workers, len(lines)%workers

		start := 0

		for worker := 0; worker < workers; worker++ {
			size := base

			if worker < extra {
				size++
			}

			end := start + size

			go func(start, end int) {
				defer wg.Done()

				calculate(start, end)
			}(start, end)

			start = end
		}

		wg.Wait()
	}

	// Выбираем ошибку по порядку входа, а не по времени завершения.
	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf(
				"line %d: %w",
				i,
				err,
			)
		}
	}

	return quotes, nil

}
