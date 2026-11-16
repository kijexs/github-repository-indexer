package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kijexs/github-repository-indexer/internal/collector/domain"
	"net/http"
	"time"
)

type GitHubAdapter struct {
	httpClient *http.Client
}

func NewGitHubAdapter(httpClient *http.Client) *GitHubAdapter {
	gitHubAdapter := GitHubAdapter{httpClient: httpClient}
	return &gitHubAdapter
}

type githubRepositoryResponse struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StarsCount  int       `json:"stargazers_count"`
	ForksCount  int       `json:"forks_count"`
	DateCreated time.Time `json:"created_at"`
}

func githubResponseToRepositoryInfo(githubRepoInfo *githubRepositoryResponse, repoInfo *domain.RepositoryInfo) {
	repoInfo.Name = githubRepoInfo.Name
	repoInfo.Description = githubRepoInfo.Description
	repoInfo.StarsCount = githubRepoInfo.StarsCount
	repoInfo.ForksCount = githubRepoInfo.ForksCount
	repoInfo.DateCreated = githubRepoInfo.DateCreated
}

func (a *GitHubAdapter) GetRepository(ctx context.Context, request domain.RepositoryRequest) (domain.RepositoryInfo, error) {
	var repoInfo domain.RepositoryInfo
	var githubRepoInfo githubRepositoryResponse
	var owner = request.Owner
	var repo = request.Repo
	var req, err = http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.github.com/repos/"+owner+"/"+repo,
		nil,
	)
	if err != nil {
		return repoInfo, fmt.Errorf("forming request: %w", err)
	}
	req.Header.Add("X-GitHub-Api-Version", "2026-03-10")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	res, err := a.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return domain.RepositoryInfo{}, fmt.Errorf("github request: %w", err)
		}
		return domain.RepositoryInfo{}, fmt.Errorf(
			"%w: github request failed: %v",
			domain.ErrUpstreamUnavailable,
			err,
		)
	}
	defer res.Body.Close()

	switch {
	case res.StatusCode == http.StatusOK:

	case res.StatusCode == http.StatusNotFound:
		return repoInfo, domain.ErrRepositoryNotFound

	case res.StatusCode >= 500 || res.StatusCode == http.StatusForbidden || res.StatusCode == http.StatusTooManyRequests:
		return repoInfo, domain.ErrUpstreamUnavailable

	default:
		return repoInfo, fmt.Errorf("%w: github status %d", domain.ErrUpstreamUnavailable, res.StatusCode)
	}

	err = json.NewDecoder(res.Body).Decode(&githubRepoInfo)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return repoInfo, fmt.Errorf("decode github response: %w", err)
		}
		return repoInfo, fmt.Errorf("%w: decode github response: %v", domain.ErrInternal, err)
	}

	githubResponseToRepositoryInfo(&githubRepoInfo, &repoInfo)
	return repoInfo, nil
}
