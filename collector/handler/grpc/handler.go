package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/kijexs/github-repository-indexer/collector/domain"
	"github.com/kijexs/github-repository-indexer/collector/usecase"
	"github.com/kijexs/github-repository-indexer/proto/repopb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RepositoryHandler struct {
	repopb.UnimplementedRepositoryServiceServer
	useCase *usecase.RepositoryUseCase
}

func NewRepositoryHandler(useCase *usecase.RepositoryUseCase) *RepositoryHandler {
	repoHandler := RepositoryHandler{useCase: useCase}
	return &repoHandler
}

func (h *RepositoryHandler) GetRepository(ctx context.Context, req *repopb.GetRepositoryRequest) (*repopb.GetRepositoryReply, error) {
	if req.Owner == "" || req.Repo == "" {
		return nil, fmt.Errorf("incorrect input params")
	}
	repoRequest := domain.RepositoryRequest{Owner: req.Owner, Repo: req.Repo}

	repoInfo, err := h.useCase.GetRepository(ctx, repoRequest)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return nil, status.Error(
				codes.InvalidArgument,
				"invalid repository parameters",
			)
		case errors.Is(err, domain.ErrRepositoryNotFound):
			return nil, status.Error(
				codes.NotFound,
				"repository not found",
			)
		case errors.Is(err, domain.ErrUpstreamUnavailable):
			return nil, status.Error(
				codes.Unavailable,
				"upstream unavailable",
			)
		default:
			return nil, status.Error(
				codes.Internal,
				"internal error",
			)
		}
	}

	reply := repopb.GetRepositoryReply{
		Name:        repoInfo.Name,
		Description: repoInfo.Description,
		StarsCount:  int64(repoInfo.StarsCount),
		ForksCount:  int64(repoInfo.ForksCount),
		DateCreated: timestamppb.New(repoInfo.DateCreated),
	}

	return &reply, nil
}
