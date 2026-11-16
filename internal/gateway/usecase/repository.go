package usecase

import (
	"context"
	"fmt"
	"github.com/kijexs/github-repository-indexer/internal/gateway/domain"
	"regexp"
)

type GatewayUseCase struct {
	provider RepositoryProvider
}

func NewGatewayUseCase(provider RepositoryProvider) *GatewayUseCase {
	usecase := GatewayUseCase{provider: provider}
	return &usecase
}

func (usecase *GatewayUseCase) GetRepository(ctx context.Context, req domain.RepositoryRequest) (*domain.RepositoryInfo, error) {
	if !validRequest(req) {
		return nil, domain.ErrInvalidInput
	}
	repoInfo, err := usecase.provider.GetRepository(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get repository: %w", err)
	}
	return &repoInfo, nil
}

type RepositoryProvider interface {
	GetRepository(context.Context, domain.RepositoryRequest) (domain.RepositoryInfo, error)
	LoadRepositories(context.Context, []domain.RepositoryRequest) ([]domain.Result, error)
}

var repositoryPart = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

func validRequest(req domain.RepositoryRequest) bool {
	return repositoryPart.MatchString(req.Owner) && repositoryPart.MatchString(req.Repo) && req.Owner != "." && req.Owner != ".." && req.Repo != "." && req.Repo != ".."
}

func (usecase *GatewayUseCase) LoadRepositories(ctx context.Context, requests []domain.RepositoryRequest) ([]domain.Result, error) {
	for _, req := range requests {
		if !validRequest(req) {
			return nil, domain.ErrInvalidInput
		}
	}
	return usecase.provider.LoadRepositories(ctx, requests)
}
