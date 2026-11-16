package handler

import (
	"context"
	"errors"
	"github.com/kijexs/github-repository-indexer/internal/collector/domain"
	"github.com/kijexs/github-repository-indexer/internal/collector/usecase"
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
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request must not be nil")
	}
	repoRequest := domain.RepositoryRequest{Owner: req.Owner, Repo: req.Repo}

	repoInfo, err := h.useCase.GetRepository(ctx, repoRequest)
	if err != nil {
		return nil, mapGRPCError(err)
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

func mapGRPCError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request timeout")
	case errors.Is(err, domain.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid repository parameters")
	case errors.Is(err, domain.ErrRepositoryNotFound):
		return status.Error(codes.NotFound, "repository not found")
	case errors.Is(err, domain.ErrUpstreamUnavailable):
		return status.Error(codes.Unavailable, "upstream unavailable")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func (h *RepositoryHandler) LoadRepositories(ctx context.Context, req *repopb.LoadRepositoriesRequest) (*repopb.LoadRepositoriesReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request must not be nil")
	}
	requests := make([]domain.RepositoryRequest, 0, len(req.Repositories))
	for _, repository := range req.Repositories {
		if repository == nil {
			return nil, status.Error(codes.InvalidArgument, "repository must not be nil")
		}
		requests = append(requests, domain.RepositoryRequest{Owner: repository.Owner, Repo: repository.Repo})
	}
	results, err := h.useCase.LoadRepositories(ctx, requests)
	if err != nil {
		return nil, mapGRPCError(err)
	}
	reply := &repopb.LoadRepositoriesReply{}
	for _, result := range results {
		item := &repopb.RepositoryResult{Owner: result.Request.Owner, Repo: result.Request.Repo}
		if result.Err != nil {
			item.Error = status.Convert(mapGRPCError(result.Err)).Message()
		} else {
			item.Data = &repopb.GetRepositoryReply{Name: result.Data.Name, Description: result.Data.Description,
				StarsCount: int64(result.Data.StarsCount), ForksCount: int64(result.Data.ForksCount), DateCreated: timestamppb.New(result.Data.DateCreated)}
		}
		reply.Results = append(reply.Results, item)
	}
	return reply, nil
}
