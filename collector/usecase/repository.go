package usecase

import (
	"context"
	"fmt"
	"github.com/kijexs/github-repository-indexer/collector/domain"
)

type RepositoryUseCase struct {
	provider domain.RepositoryProvider
}

func NewRepositoryUseCase(provider domain.RepositoryProvider) *RepositoryUseCase {
	usecase := RepositoryUseCase{provider: provider}
	return &usecase
}

func (usecase *RepositoryUseCase) GetRepository(ctx context.Context, req domain.RepositoryRequest) (*domain.RepositoryInfo, error) {
	if req.Owner == "" || req.Repo == "" {
		return nil, domain.ErrInvalidInput
	}
	repoInfo, err := usecase.provider.GetRepository(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get repository: %w", err)
	}
	return &repoInfo, nil
}
