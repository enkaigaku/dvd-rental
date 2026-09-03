package repository

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// textToString converts a pgtype.Text to a Go string. NULL maps to "".
func textToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// stringToText converts a Go string to pgtype.Text. "" maps to NULL.
func stringToText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

// dateToTime converts a pgtype.Date to time.Time.
func dateToTime(d pgtype.Date) time.Time {
	if !d.Valid {
		return time.Time{}
	}
	return d.Time
}

// timestamptzToTime converts a pgtype.Timestamptz to time.Time.
func timestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

// boolToActive converts a boolean to the legacy active integer field.
// true → 1, false → 0.
func boolToActive(b bool) pgtype.Int4 {
	v := int32(0)
	if b {
		v = 1
	}
	return pgtype.Int4{Int32: v, Valid: true}
}
