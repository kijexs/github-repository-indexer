package domain

import (
	"context"
	"time"
)

type RepositoryRequest struct {
	Owner string
	Repo  string
}

type RepositoryInfo struct {
	Name        string
	Description string
	StarsCount  int
	ForksCount  int
	DateCreated time.Time
}

type RepositoryProvider interface {
	GetRepository(ctx context.Context, request RepositoryRequest) (RepositoryInfo, error)
}
