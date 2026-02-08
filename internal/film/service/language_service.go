package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tokyoyuan/dvd-rental/internal/film/model"
	"github.com/tokyoyuan/dvd-rental/internal/film/repository"
)

// LanguageService contains business logic for language operations (read-only).
type LanguageService struct {
	languageRepo repository.LanguageRepository
}

// NewLanguageService creates a new LanguageService.
func NewLanguageService(languageRepo repository.LanguageRepository) *LanguageService {
	return &LanguageService{languageRepo: languageRepo}
}

// GetLanguage returns a single language by ID.
func (s *LanguageService) GetLanguage(ctx context.Context, languageID int32) (model.Language, error) {
	if languageID <= 0 {
		return model.Language{}, fmt.Errorf("language_id must be positive: %w", ErrInvalidArgument)
	}

	language, err := s.languageRepo.GetLanguage(ctx, languageID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Language{}, fmt.Errorf("language %d: %w", languageID, ErrNotFound)
		}
		return model.Language{}, err
	}
	return language, nil
}

// ListLanguages returns all languages.
func (s *LanguageService) ListLanguages(ctx context.Context) ([]model.Language, int64, error) {
	languages, err := s.languageRepo.ListLanguages(ctx)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.languageRepo.CountLanguages(ctx)
	if err != nil {
		return nil, 0, err
	}

	return languages, total, nil
}
