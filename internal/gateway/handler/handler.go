package handler

import (
	"encoding/json"
	"errors"
	"github.com/kijexs/github-repository-indexer/internal/gateway/domain"
	"github.com/kijexs/github-repository-indexer/internal/gateway/usecase"
	"net/http"
	"time"
)

type RepositoryHandler struct {
	usecase *usecase.GatewayUseCase
}

func NewRepositoryHandler(usecase *usecase.GatewayUseCase) *RepositoryHandler {
	handler := RepositoryHandler{usecase: usecase}
	return &handler
}

type JsonRepositoryResponse struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StarsCount  int       `json:"stargazers_count"`
	ForksCount  int       `json:"forks_count"`
	DateCreated time.Time `json:"created_at"`
}

// GetRepository godoc
// @Summary Get repository information
// @Description Returns information about a GitHub repository
// @Tags repositories
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Success 200 {object} JsonRepositoryResponse
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Failure 503 {string} string
// @Router /repositories/{owner}/{repo} [get]
func (h *RepositoryHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	req := domain.RepositoryRequest{Owner: owner, Repo: repo}

	repoInfo, err := h.usecase.GetRepository(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			http.Error(w, "invalid input", http.StatusBadRequest)

		case errors.Is(err, domain.ErrRepositoryNotFound):
			http.Error(w, "repository not found", http.StatusNotFound)

		case errors.Is(err, domain.ErrUpstreamUnavailable):
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)

		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	jsonRepoInfo := JsonRepositoryResponse{
		Name:        repoInfo.Name,
		Description: repoInfo.Description,
		StarsCount:  repoInfo.StarsCount,
		ForksCount:  repoInfo.ForksCount,
		DateCreated: repoInfo.DateCreated,
	}

	repoJSON, err := json.Marshal(jsonRepoInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(repoJSON)
}
