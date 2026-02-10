package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/enkaigaku/dvd-rental/internal/film/model"
	"github.com/enkaigaku/dvd-rental/gen/sqlc/film"
)

// LanguageRepository defines the read-only data access interface for languages.
type LanguageRepository interface {
	GetLanguage(ctx context.Context, languageID int32) (model.Language, error)
	ListLanguages(ctx context.Context) ([]model.Language, error)
	CountLanguages(ctx context.Context) (int64, error)
}

type languageRepository struct {
	q *filmsqlc.Queries
}

// NewLanguageRepository creates a new LanguageRepository backed by PostgreSQL.
func NewLanguageRepository(pool *pgxpool.Pool) LanguageRepository {
	return &languageRepository{q: filmsqlc.New(pool)}
}

func (r *languageRepository) GetLanguage(ctx context.Context, languageID int32) (model.Language, error) {
	row, err := r.q.GetLanguage(ctx, languageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Language{}, ErrNotFound
		}
		return model.Language{}, fmt.Errorf("get language: %w", err)
	}
	return toLanguageModel(row), nil
}

func (r *languageRepository) ListLanguages(ctx context.Context) ([]model.Language, error) {
	rows, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("list languages: %w", err)
	}
	return toLanguageModels(rows), nil
}

func (r *languageRepository) CountLanguages(ctx context.Context) (int64, error) {
	count, err := r.q.CountLanguages(ctx)
	if err != nil {
		return 0, fmt.Errorf("count languages: %w", err)
	}
	return count, nil
}

// --- row to model conversions ---
// language.name is character(20), space-padded — must trim.

func toLanguageModel(r filmsqlc.Language) model.Language {
	return model.Language{
		LanguageID: r.LanguageID,
		Name:       trimLanguageName(r.Name),
		LastUpdate: r.LastUpdate.Time,
	}
}

func toLanguageModels(rows []filmsqlc.Language) []model.Language {
	languages := make([]model.Language, len(rows))
	for i, r := range rows {
		languages[i] = toLanguageModel(r)
	}
	return languages
}
