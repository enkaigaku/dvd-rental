package repository

import (
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/enkaigaku/dvd-rental/internal/film/repository/sqlcgen"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("record not found")

// --- pgtype.Text helpers ---

func textToString(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func stringToText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// --- pgtype.Numeric helpers ---

func numericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0"
	}
	// Use big.Float for precise conversion.
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

// --- pgtype.Int4 / Int2 helpers ---

func int4ToInt32(n pgtype.Int4) int32 {
	if n.Valid {
		return n.Int32
	}
	return 0
}

func int32ToInt4(v int32) pgtype.Int4 {
	if v == 0 {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: v, Valid: true}
}

func int2ToInt16(n pgtype.Int2) int16 {
	if n.Valid {
		return n.Int16
	}
	return 0
}

func int16ToInt2(v int16) pgtype.Int2 {
	if v == 0 {
		return pgtype.Int2{Valid: false}
	}
	return pgtype.Int2{Int16: v, Valid: true}
}

// --- year domain helpers ---
// The public.year domain maps to interface{} in sqlc.
// At runtime pgx returns it as int32.

func yearToInt32(v interface{}) int32 {
	if v == nil {
		return 0
	}
	switch y := v.(type) {
	case int32:
		return y
	case int64:
		return int32(y)
	case float64:
		return int32(y)
	default:
		return 0
	}
}

// --- mpaa_rating helpers ---

func ratingToString(r sqlcgen.NullMpaaRating) string {
	if r.Valid {
		return string(r.MpaaRating)
	}
	return ""
}

func stringToRating(s string) sqlcgen.NullMpaaRating {
	if s == "" {
		return sqlcgen.NullMpaaRating{Valid: false}
	}
	return sqlcgen.NullMpaaRating{MpaaRating: sqlcgen.MpaaRating(s), Valid: true}
}

// --- language name helper ---
// language.name is character(20) which is space-padded.

func trimLanguageName(s string) string {
	return strings.TrimRight(s, " ")
}

// --- film row conversion ---
// All film query row types have the same columns, so we use a generic helper.

type filmRow interface {
	getFilmFields() (
		filmID int32, title string, description pgtype.Text,
		releaseYear interface{}, languageID int32, originalLanguageID pgtype.Int4,
		rentalDuration int16, rentalRate pgtype.Numeric, length pgtype.Int2,
		replacementCost pgtype.Numeric, rating sqlcgen.NullMpaaRating,
		specialFeatures []string, lastUpdate pgtype.Timestamptz,
	)
}

// filmFields holds the common fields extracted from any film row type.
type filmFields struct {
	FilmID             int32
	Title              string
	Description        pgtype.Text
	ReleaseYear        interface{}
	LanguageID         int32
	OriginalLanguageID pgtype.Int4
	RentalDuration     int16
	RentalRate         pgtype.Numeric
	Length             pgtype.Int2
	ReplacementCost    pgtype.Numeric
	Rating             sqlcgen.NullMpaaRating
	SpecialFeatures    []string
	LastUpdate         pgtype.Timestamptz
}

func convertFilmFields(f filmFields) filmConvertedFields {
	return filmConvertedFields{
		FilmID:             f.FilmID,
		Title:              f.Title,
		Description:        textToString(f.Description),
		ReleaseYear:        yearToInt32(f.ReleaseYear),
		LanguageID:         f.LanguageID,
		OriginalLanguageID: int4ToInt32(f.OriginalLanguageID),
		RentalDuration:     f.RentalDuration,
		RentalRate:         numericToString(f.RentalRate),
		Length:             int2ToInt16(f.Length),
		ReplacementCost:    numericToString(f.ReplacementCost),
		Rating:             ratingToString(f.Rating),
		SpecialFeatures:    f.SpecialFeatures,
		LastUpdate:         f.LastUpdate.Time,
	}
}

type filmConvertedFields struct {
	FilmID             int32
	Title              string
	Description        string
	ReleaseYear        int32
	LanguageID         int32
	OriginalLanguageID int32
	RentalDuration     int16
	RentalRate         string
	Length             int16
	ReplacementCost    string
	Rating             string
	SpecialFeatures    []string
	LastUpdate         time.Time
}
