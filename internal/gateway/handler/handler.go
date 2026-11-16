package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/kijexs/github-repository-indexer/internal/gateway/domain"
	"github.com/kijexs/github-repository-indexer/internal/gateway/usecase"
	"io"
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
// @Failure 504 {string} string
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
		http.Error(w, err.Error(), mapHTTPError(err))
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

func mapHTTPError(err error) int {
	switch {
	case errors.Is(err, context.Canceled):
		return http.StatusRequestTimeout
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrRepositoryNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrUpstreamUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

type JsonRepositoryRequest struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
}

type JsonLoadRequest struct {
	Repositories []JsonRepositoryRequest `json:"repositories"`
}

type JsonRepositoryResult struct {
	Owner string                  `json:"owner"`
	Repo  string                  `json:"repo"`
	Data  *JsonRepositoryResponse `json:"data,omitempty"`
	Error string                  `json:"error,omitempty"`
}

type JsonLoadResponse struct {
	Results []JsonRepositoryResult `json:"results"`
}

// LoadRepositories godoc
// @Summary Load repositories from GitHub and update saved data
// @Tags repositories
// @Accept json
// @Produce json
// @Param request body JsonLoadRequest true "Repositories to load"
// @Success 200 {object} JsonLoadResponse
// @Failure 400 {string} string
// @Failure 408 {string} string
// @Failure 500 {string} string
// @Failure 503 {string} string
// @Failure 504 {string} string
// @Router /repositories [post]
func (h *RepositoryHandler) LoadRepositories(w http.ResponseWriter, r *http.Request) {
	var req JsonLoadRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		http.Error(w, "expected one JSON object", http.StatusBadRequest)
		return
	}
	if req.Repositories == nil {
		http.Error(w, "repositories must be an array", http.StatusBadRequest)
		return
	}
	requests := make([]domain.RepositoryRequest, 0, len(req.Repositories))
	for _, repository := range req.Repositories {
		requests = append(requests, domain.RepositoryRequest{Owner: repository.Owner, Repo: repository.Repo})
	}
	results, err := h.usecase.LoadRepositories(r.Context(), requests)
	if err != nil {
		http.Error(w, err.Error(), mapHTTPError(err))
		return
	}
	response := JsonLoadResponse{Results: make([]JsonRepositoryResult, 0, len(results))}
	for _, result := range results {
		item := JsonRepositoryResult{Owner: result.Request.Owner, Repo: result.Request.Repo}
		if result.Err != nil {
			item.Error = result.Err.Error()
		} else {
			item.Data = &JsonRepositoryResponse{Name: result.Data.Name, Description: result.Data.Description, StarsCount: result.Data.StarsCount, ForksCount: result.Data.ForksCount, DateCreated: result.Data.DateCreated}
		}
		response.Results = append(response.Results, item)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
