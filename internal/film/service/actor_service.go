package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/enkaigaku/dvd-rental/internal/film/model"
	"github.com/enkaigaku/dvd-rental/internal/film/repository"
)

// ActorService contains business logic for actor operations.
type ActorService struct {
	actorRepo repository.ActorRepository
}

// NewActorService creates a new ActorService.
func NewActorService(actorRepo repository.ActorRepository) *ActorService {
	return &ActorService{actorRepo: actorRepo}
}

// GetActor returns a single actor by ID.
func (s *ActorService) GetActor(ctx context.Context, actorID int32) (model.Actor, error) {
	if actorID <= 0 {
		return model.Actor{}, fmt.Errorf("actor_id must be positive: %w", ErrInvalidArgument)
	}

	actor, err := s.actorRepo.GetActor(ctx, actorID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Actor{}, fmt.Errorf("actor %d: %w", actorID, ErrNotFound)
		}
		return model.Actor{}, err
	}
	return actor, nil
}

// ListActors returns a paginated list of actors.
func (s *ActorService) ListActors(ctx context.Context, pageSize, page int32) ([]model.Actor, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	actors, err := s.actorRepo.ListActors(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.actorRepo.CountActors(ctx)
	if err != nil {
		return nil, 0, err
	}

	return actors, total, nil
}

// ListActorsByFilm returns all actors for a given film.
func (s *ActorService) ListActorsByFilm(ctx context.Context, filmID int32) ([]model.Actor, error) {
	if filmID <= 0 {
		return nil, fmt.Errorf("film_id must be positive: %w", ErrInvalidArgument)
	}

	actors, err := s.actorRepo.ListActorsByFilm(ctx, filmID)
	if err != nil {
		return nil, err
	}
	return actors, nil
}

// CreateActor creates a new actor.
func (s *ActorService) CreateActor(ctx context.Context, firstName, lastName string) (model.Actor, error) {
	if firstName == "" {
		return model.Actor{}, fmt.Errorf("first_name must not be empty: %w", ErrInvalidArgument)
	}
	if lastName == "" {
		return model.Actor{}, fmt.Errorf("last_name must not be empty: %w", ErrInvalidArgument)
	}

	actor, err := s.actorRepo.CreateActor(ctx, firstName, lastName)
	if err != nil {
		return model.Actor{}, err
	}
	return actor, nil
}

// UpdateActor updates an existing actor.
func (s *ActorService) UpdateActor(ctx context.Context, actorID int32, firstName, lastName string) (model.Actor, error) {
	if actorID <= 0 {
		return model.Actor{}, fmt.Errorf("actor_id must be positive: %w", ErrInvalidArgument)
	}
	if firstName == "" {
		return model.Actor{}, fmt.Errorf("first_name must not be empty: %w", ErrInvalidArgument)
	}
	if lastName == "" {
		return model.Actor{}, fmt.Errorf("last_name must not be empty: %w", ErrInvalidArgument)
	}

	actor, err := s.actorRepo.UpdateActor(ctx, actorID, firstName, lastName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Actor{}, fmt.Errorf("actor %d: %w", actorID, ErrNotFound)
		}
		return model.Actor{}, err
	}
	return actor, nil
}

// DeleteActor deletes an actor. Fails if film_actor records reference it.
func (s *ActorService) DeleteActor(ctx context.Context, actorID int32) error {
	if actorID <= 0 {
		return fmt.Errorf("actor_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify the actor exists.
	if _, err := s.actorRepo.GetActor(ctx, actorID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("actor %d: %w", actorID, ErrNotFound)
		}
		return err
	}

	if err := s.actorRepo.DeleteActor(ctx, actorID); err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("actor %d is referenced by film records: %w", actorID, ErrForeignKey)
		}
		return err
	}
	return nil
}
