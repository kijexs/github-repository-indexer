package domain

import "errors"

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrRepositoryNotFound  = errors.New("repository not found")
	ErrUpstreamUnavailable = errors.New("upstream unavailable")
	ErrInternal            = errors.New("internal error")
)
