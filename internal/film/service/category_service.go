package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tokyoyuan/dvd-rental/internal/film/model"
	"github.com/tokyoyuan/dvd-rental/internal/film/repository"
)

// CategoryService contains business logic for category operations (read-only).
type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

// GetCategory returns a single category by ID.
func (s *CategoryService) GetCategory(ctx context.Context, categoryID int32) (model.Category, error) {
	if categoryID <= 0 {
		return model.Category{}, fmt.Errorf("category_id must be positive: %w", ErrInvalidArgument)
	}

	category, err := s.categoryRepo.GetCategory(ctx, categoryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Category{}, fmt.Errorf("category %d: %w", categoryID, ErrNotFound)
		}
		return model.Category{}, err
	}
	return category, nil
}

// ListCategories returns all categories.
func (s *CategoryService) ListCategories(ctx context.Context) ([]model.Category, int64, error) {
	categories, err := s.categoryRepo.ListCategories(ctx)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.categoryRepo.CountCategories(ctx)
	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}
