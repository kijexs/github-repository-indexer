package adapter

import (
	"context"
	"errors"
	"github.com/kijexs/github-repository-indexer/internal/gateway/domain"
	"github.com/kijexs/github-repository-indexer/proto/repopb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CollectorAdapter struct {
	client repopb.RepositoryServiceClient
}

func NewCollectorAdapter(client repopb.RepositoryServiceClient) *CollectorAdapter {
	return &CollectorAdapter{client: client}
}

func (a *CollectorAdapter) GetRepository(ctx context.Context, request domain.RepositoryRequest) (domain.RepositoryInfo, error) {
	repoRequest := repopb.GetRepositoryRequest{
		Owner: request.Owner,
		Repo:  request.Repo,
	}
	repoReply, err := a.client.GetRepository(ctx, &repoRequest)
	if err != nil {
		return domain.RepositoryInfo{}, mapCollectorError(err)
	}
	repoInfo := domain.RepositoryInfo{
		Name:        repoReply.Name,
		Description: repoReply.Description,
		StarsCount:  int(repoReply.StarsCount),
		ForksCount:  int(repoReply.ForksCount),
		DateCreated: repoReply.DateCreated.AsTime(),
	}
	return repoInfo, nil
}

func mapCollectorError(err error) error {
	switch status.Code(err) {
	case codes.Canceled:
		return context.Canceled
	case codes.DeadlineExceeded:
		return context.DeadlineExceeded
	case codes.NotFound:
		return domain.ErrRepositoryNotFound
	case codes.InvalidArgument:
		return domain.ErrInvalidInput
	case codes.Unavailable:
		return domain.ErrUpstreamUnavailable
	default:
		return domain.ErrInternal
	}
}

func (a *CollectorAdapter) LoadRepositories(ctx context.Context, requests []domain.RepositoryRequest) ([]domain.Result, error) {
	req := &repopb.LoadRepositoriesRequest{}
	for _, request := range requests {
		req.Repositories = append(req.Repositories, &repopb.GetRepositoryRequest{Owner: request.Owner, Repo: request.Repo})
	}
	reply, err := a.client.LoadRepositories(ctx, req)
	if err != nil {
		return nil, mapCollectorError(err)
	}
	results := make([]domain.Result, 0, len(reply.Results))
	for _, item := range reply.Results {
		result := domain.Result{Request: domain.RepositoryRequest{Owner: item.Owner, Repo: item.Repo}}
		if item.Error != "" {
			result.Err = errors.New(item.Error)
		} else if item.Data != nil {
			result.Data = domain.RepositoryInfo{Name: item.Data.Name, Description: item.Data.Description, StarsCount: int(item.Data.StarsCount), ForksCount: int(item.Data.ForksCount), DateCreated: item.Data.DateCreated.AsTime()}
		} else {
			result.Err = domain.ErrInternal
		}
		results = append(results, result)
	}
	return results, nil
}
