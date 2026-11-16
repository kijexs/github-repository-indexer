package adapter

import (
	"context"
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
		switch status.Code(err) {
		case codes.NotFound:
			return domain.RepositoryInfo{}, domain.ErrRepositoryNotFound
		case codes.InvalidArgument:
			return domain.RepositoryInfo{}, domain.ErrInvalidInput
		case codes.Unavailable:
			return domain.RepositoryInfo{}, domain.ErrUpstreamUnavailable
		default:
			return domain.RepositoryInfo{}, domain.ErrInternal
		}
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
