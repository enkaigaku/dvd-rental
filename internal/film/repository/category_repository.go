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

// CategoryRepository defines the read-only data access interface for categories.
type CategoryRepository interface {
	GetCategory(ctx context.Context, categoryID int32) (model.Category, error)
	ListCategories(ctx context.Context) ([]model.Category, error)
	CountCategories(ctx context.Context) (int64, error)
	ListCategoriesByFilm(ctx context.Context, filmID int32) ([]model.Category, error)
}

type categoryRepository struct {
	q *filmsqlc.Queries
}

// NewCategoryRepository creates a new CategoryRepository backed by PostgreSQL.
func NewCategoryRepository(pool *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{q: filmsqlc.New(pool)}
}

func (r *categoryRepository) GetCategory(ctx context.Context, categoryID int32) (model.Category, error) {
	row, err := r.q.GetCategory(ctx, categoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Category{}, ErrNotFound
		}
		return model.Category{}, fmt.Errorf("get category: %w", err)
	}
	return toCategoryModel(row), nil
}

func (r *categoryRepository) ListCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return toCategoryModels(rows), nil
}

func (r *categoryRepository) CountCategories(ctx context.Context) (int64, error) {
	count, err := r.q.CountCategories(ctx)
	if err != nil {
		return 0, fmt.Errorf("count categories: %w", err)
	}
	return count, nil
}

func (r *categoryRepository) ListCategoriesByFilm(ctx context.Context, filmID int32) ([]model.Category, error) {
	rows, err := r.q.ListCategoriesByFilm(ctx, filmID)
	if err != nil {
		return nil, fmt.Errorf("list categories by film: %w", err)
	}
	return toCategoryModels(rows), nil
}

// --- row to model conversions ---

func toCategoryModel(r filmsqlc.Category) model.Category {
	return model.Category{
		CategoryID: r.CategoryID,
		Name:       r.Name,
		LastUpdate: r.LastUpdate.Time,
	}
}

func toCategoryModels(rows []filmsqlc.Category) []model.Category {
	categories := make([]model.Category, len(rows))
	for i, r := range rows {
		categories[i] = toCategoryModel(r)
	}
	return categories
}
