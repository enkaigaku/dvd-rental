package service

import "errors"

var (
	// ErrNotFound indicates the requested entity does not exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidArgument indicates a client-supplied value is invalid.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrAlreadyExists indicates a uniqueness constraint was violated.
	ErrAlreadyExists = errors.New("already exists")
	// ErrForeignKey indicates the entity is referenced by another entity.
	ErrForeignKey = errors.New("referenced by another entity")
)
