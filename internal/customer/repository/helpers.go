package repository

import (
	"fmt"
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

// numericToFloat tries to convert an interface{} (from a PostgreSQL numeric COALESCE result)
// to float64. Returns (value, true) on success.
func numericToFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int64:
		return float64(n), true
	case string:
		var f float64
		if _, err := fmt.Sscanf(n, "%f", &f); err == nil {
			return f, true
		}
	case pgtype.Numeric:
		if n.Valid {
			f, err := n.Float64Value()
			if err == nil && f.Valid {
				return f.Float64, true
			}
		}
	}
	return 0, false
}
