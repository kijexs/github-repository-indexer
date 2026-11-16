package usecase

import (
	"context"
	"github.com/kijexs/github-repository-indexer/internal/collector/domain"
)

type RepositoryFetcher func(ctx context.Context, req domain.RepositoryRequest) (domain.RepositoryInfo, error)

func worker(ctx context.Context, fetch RepositoryFetcher, jobs <-chan domain.RepositoryRequest, results chan<- domain.Result) {
	for request := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			repoInfo, err := fetch(ctx, request)

			result := domain.Result{
				Request: request,
				Data:    repoInfo,
				Err:     err,
			}

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}
