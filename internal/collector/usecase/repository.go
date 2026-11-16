package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kijexs/github-repository-indexer/internal/collector/domain"
)

type RepositoryProvider interface {
	GetRepository(context.Context, domain.RepositoryRequest) (domain.RepositoryInfo, error)
}

type RepositoryRepository interface {
	GetRepository(context.Context, domain.RepositoryRequest) (domain.RepositoryInfo, error)
	Save(context.Context, domain.RepositoryRequest, domain.RepositoryInfo) error
}

type RepositoryUseCase struct {
	provider     RepositoryProvider
	repository   RepositoryRepository
	workerCount  int
	batchTimeout time.Duration
}

func NewRepositoryUseCase(provider RepositoryProvider, repository RepositoryRepository, workerCount int, batchTimeout time.Duration) *RepositoryUseCase {
	if workerCount <= 0 {
		workerCount = 4
	}
	if batchTimeout <= 0 {
		batchTimeout = time.Minute
	}
	return &RepositoryUseCase{provider: provider, repository: repository, workerCount: workerCount, batchTimeout: batchTimeout}
}

var repositoryPart = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

func validRequest(req domain.RepositoryRequest) bool {
	return repositoryPart.MatchString(req.Owner) && repositoryPart.MatchString(req.Repo) && req.Owner != "." && req.Owner != ".." && req.Repo != "." && req.Repo != ".."
}

func normalizeRequest(req domain.RepositoryRequest) domain.RepositoryRequest {
	return domain.RepositoryRequest{Owner: strings.ToLower(req.Owner), Repo: strings.ToLower(req.Repo)}
}

func (usecase *RepositoryUseCase) GetRepository(ctx context.Context, req domain.RepositoryRequest) (*domain.RepositoryInfo, error) {
	if !validRequest(req) {
		return nil, domain.ErrInvalidInput
	}
	req = normalizeRequest(req)
	repoInfo, err := usecase.repository.GetRepository(ctx, req)
	if err == nil {
		return &repoInfo, nil
	}
	if !errors.Is(err, domain.ErrRepositoryNotFound) {
		return nil, fmt.Errorf("get saved repository: %w", err)
	}
	repoInfo, err = usecase.fetchRepository(ctx, req)
	if err != nil {
		return nil, err
	}
	return &repoInfo, nil
}

func (usecase *RepositoryUseCase) fetchRepository(ctx context.Context, req domain.RepositoryRequest) (domain.RepositoryInfo, error) {
	repoInfo, err := usecase.provider.GetRepository(ctx, req)
	if err != nil {
		return domain.RepositoryInfo{}, fmt.Errorf("get repository: %w", err)
	}
	if err := usecase.repository.Save(ctx, req, repoInfo); err != nil {
		return domain.RepositoryInfo{}, fmt.Errorf("save repository: %w", err)
	}
	return repoInfo, nil
}

func (usecase *RepositoryUseCase) LoadRepositories(ctx context.Context, requests []domain.RepositoryRequest) ([]domain.Result, error) {
	normalized := make([]domain.RepositoryRequest, len(requests))
	for i, req := range requests {
		if !validRequest(req) {
			return nil, domain.ErrInvalidInput
		}
		normalized[i] = normalizeRequest(req)
	}
	ctx, cancel := context.WithTimeout(ctx, usecase.batchTimeout)
	defer cancel()
	return loadRepositoriesParallel(ctx, normalized, usecase.workerCount, usecase.fetchRepository)
}
