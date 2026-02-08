package service

import "errors"

var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidArgument indicates the request contains invalid parameters.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrAlreadyExists indicates a uniqueness constraint violation.
	ErrAlreadyExists = errors.New("already exists")
	// ErrForeignKey indicates the operation violates a foreign key constraint.
	ErrForeignKey = errors.New("referenced by another entity")
)
