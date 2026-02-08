package service

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL error code for unique_violation.
const pgUniqueViolation = "23505"

// isUniqueViolation checks whether the error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}
