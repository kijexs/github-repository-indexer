package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kijexs/github-repository-indexer/internal/collector/domain"
)

type RepositoryRepository struct{ pool *pgxpool.Pool }

func NewRepositoryRepository(pool *pgxpool.Pool) *RepositoryRepository {
	return &RepositoryRepository{pool: pool}
}

func (r *RepositoryRepository) GetRepository(ctx context.Context, req domain.RepositoryRequest) (domain.RepositoryInfo, error) {
	var repo domain.RepositoryInfo
	err := r.pool.QueryRow(ctx, `SELECT name, description, stars, forks, created_at
 FROM repositories WHERE lower(owner) = $1 AND lower(name) = $2`, req.Owner, req.Repo).Scan(
		&repo.Name, &repo.Description, &repo.StarsCount, &repo.ForksCount, &repo.DateCreated)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RepositoryInfo{}, domain.ErrRepositoryNotFound
	}
	if err != nil {
		return domain.RepositoryInfo{}, fmt.Errorf("get repository %s/%s: %w", req.Owner, req.Repo, err)
	}
	return repo, nil
}

func (r *RepositoryRepository) Save(ctx context.Context, req domain.RepositoryRequest, repo domain.RepositoryInfo) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO repositories (owner, name, description, stars, forks, created_at)
 VALUES ($1, $2, $3, $4, $5, $6)
 ON CONFLICT (lower(owner), lower(name)) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description,
 stars = EXCLUDED.stars, forks = EXCLUDED.forks, created_at = EXCLUDED.created_at, fetched_at = clock_timestamp()`,
		req.Owner, repo.Name, repo.Description, repo.StarsCount, repo.ForksCount, repo.DateCreated)
	if err != nil {
		return fmt.Errorf("save repository %s/%s: %w", req.Owner, req.Repo, err)
	}
	return nil
}
