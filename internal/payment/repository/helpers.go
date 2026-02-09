package repository

import (
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func timestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

func timeToTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: !t.IsZero()}
}

func numericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0"
	}
	val := new(big.Float).SetInt(n.Int)
	if n.Exp < 0 {
		divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil))
		val.Quo(val, divisor)
	} else if n.Exp > 0 {
		multiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil))
		val.Mul(val, multiplier)
	}
	return val.Text('f', int(max32(-n.Exp, 0)))
}

func stringToNumeric(s string) pgtype.Numeric {
	var n pgtype.Numeric
	if s == "" {
		return pgtype.Numeric{Valid: false}
	}
	if err := n.Scan(s); err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return n
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
