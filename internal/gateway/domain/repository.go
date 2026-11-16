package domain

import "time"

type RepositoryRequest struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
}

type RepositoryInfo struct {
	Name        string
	Description string
	StarsCount  int
	ForksCount  int
	DateCreated time.Time
}

type Result struct {
	Request RepositoryRequest
	Data    RepositoryInfo
	Err     error
}
