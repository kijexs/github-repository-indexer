package github

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kijexs/github-repository-indexer/collector/domain"
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
		return domain.RepositoryInfo{}, fmt.Errorf(
			"%w: github request failed: %v",
			domain.ErrUpstreamUnavailable,
			err,
		)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return repoInfo, fmt.Errorf("unexpected status: got %v", res.StatusCode)
	}

	switch {
	case res.StatusCode == http.StatusOK:

	case res.StatusCode == http.StatusNotFound:
		return repoInfo, domain.ErrRepositoryNotFound

	case res.StatusCode >= 500 && res.StatusCode <= 599:
		return repoInfo, domain.ErrUpstreamUnavailable

	default:
		return repoInfo, domain.ErrInternal
	}

	err = json.NewDecoder(res.Body).Decode(&githubRepoInfo)
	if err != nil {
		return repoInfo, domain.ErrInternal
	}

	githubResponseToRepositoryInfo(&githubRepoInfo, &repoInfo)
	return repoInfo, nil
}
