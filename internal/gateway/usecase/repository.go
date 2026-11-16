package usecase

import (
	"context"
	"fmt"
	"github.com/kijexs/github-repository-indexer/internal/gateway/domain"
)

type GatewayUseCase struct {
	provider domain.RepositoryProvider
}

func NewGatewayUseCase(provider domain.RepositoryProvider) *GatewayUseCase {
	usecase := GatewayUseCase{provider: provider}
	return &usecase
}

func (usecase *GatewayUseCase) GetRepository(ctx context.Context, req domain.RepositoryRequest) (*domain.RepositoryInfo, error) {
	if req.Owner == "" || req.Repo == "" {
		return nil, domain.ErrInvalidInput
	}
	repoInfo, err := usecase.provider.GetRepository(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get repository: %w", err)
	}
	return &repoInfo, nil
}
