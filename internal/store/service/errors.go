// Package service implements business logic for the store service.
package service

import "errors"

var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidArgument indicates a request parameter is invalid.
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrAlreadyExists indicates a resource with the same unique key already exists.
	ErrAlreadyExists = errors.New("already exists")
)
