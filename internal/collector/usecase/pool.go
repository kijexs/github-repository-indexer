package usecase

import (
	"context"
	"sync"

	"github.com/kijexs/github-repository-indexer/internal/collector/domain"
)

func loadRepositoriesParallel(ctx context.Context, requests []domain.RepositoryRequest, workerCount int, fetch RepositoryFetcher) ([]domain.Result, error) {
	result := []domain.Result{}
	if workerCount <= 0 {
		workerCount = 4
	}
	if len(requests) == 0 {
		return result, ctx.Err()
	}
	var wg sync.WaitGroup

	jobs := make(chan domain.RepositoryRequest)
	results := make(chan domain.Result)

	wg.Go(func() {
		defer close(jobs)
		for _, request := range requests {
			select {
			case <-ctx.Done():
				return
			case jobs <- request:
			}

		}
	})

	for i := 0; i < workerCount; i++ {
		wg.Go(func() {
			worker(ctx, fetch, jobs, results)
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		result = append(result, res)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
