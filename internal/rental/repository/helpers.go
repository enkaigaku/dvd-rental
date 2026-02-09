package repository

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// timestamptzToTime converts a pgtype.Timestamptz to time.Time.
// NULL maps to the zero value of time.Time.
func timestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}
