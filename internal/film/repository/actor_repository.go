package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/enkaigaku/dvd-rental/internal/film/model"
	"github.com/enkaigaku/dvd-rental/internal/film/repository/sqlcgen"
)

// ActorRepository defines the data access interface for actors.
type ActorRepository interface {
	GetActor(ctx context.Context, actorID int32) (model.Actor, error)
	ListActors(ctx context.Context, limit, offset int32) ([]model.Actor, error)
	CountActors(ctx context.Context) (int64, error)
	ListActorsByFilm(ctx context.Context, filmID int32) ([]model.Actor, error)
	CreateActor(ctx context.Context, firstName, lastName string) (model.Actor, error)
	UpdateActor(ctx context.Context, actorID int32, firstName, lastName string) (model.Actor, error)
	DeleteActor(ctx context.Context, actorID int32) error
}

type actorRepository struct {
	q *sqlcgen.Queries
}

// NewActorRepository creates a new ActorRepository backed by PostgreSQL.
func NewActorRepository(pool *pgxpool.Pool) ActorRepository {
	return &actorRepository{q: sqlcgen.New(pool)}
}

func (r *actorRepository) GetActor(ctx context.Context, actorID int32) (model.Actor, error) {
	row, err := r.q.GetActor(ctx, actorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Actor{}, ErrNotFound
		}
		return model.Actor{}, fmt.Errorf("get actor: %w", err)
	}
	return toActorModel(row), nil
}

func (r *actorRepository) ListActors(ctx context.Context, limit, offset int32) ([]model.Actor, error) {
	rows, err := r.q.ListActors(ctx, sqlcgen.ListActorsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list actors: %w", err)
	}
	return toActorModels(rows), nil
}

func (r *actorRepository) CountActors(ctx context.Context) (int64, error) {
	count, err := r.q.CountActors(ctx)
	if err != nil {
		return 0, fmt.Errorf("count actors: %w", err)
	}
	return count, nil
}

func (r *actorRepository) ListActorsByFilm(ctx context.Context, filmID int32) ([]model.Actor, error) {
	rows, err := r.q.ListActorsByFilm(ctx, filmID)
	if err != nil {
		return nil, fmt.Errorf("list actors by film: %w", err)
	}
	return toActorModels(rows), nil
}

func (r *actorRepository) CreateActor(ctx context.Context, firstName, lastName string) (model.Actor, error) {
	row, err := r.q.CreateActor(ctx, sqlcgen.CreateActorParams{
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		return model.Actor{}, fmt.Errorf("create actor: %w", err)
	}
	return toActorModel(row), nil
}

func (r *actorRepository) UpdateActor(ctx context.Context, actorID int32, firstName, lastName string) (model.Actor, error) {
	row, err := r.q.UpdateActor(ctx, sqlcgen.UpdateActorParams{
		ActorID:   actorID,
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Actor{}, ErrNotFound
		}
		return model.Actor{}, fmt.Errorf("update actor: %w", err)
	}
	return toActorModel(row), nil
}

func (r *actorRepository) DeleteActor(ctx context.Context, actorID int32) error {
	if err := r.q.DeleteActor(ctx, actorID); err != nil {
		return fmt.Errorf("delete actor: %w", err)
	}
	return nil
}

// --- row to model conversions ---

func toActorModel(r sqlcgen.Actor) model.Actor {
	return model.Actor{
		ActorID:    r.ActorID,
		FirstName:  r.FirstName,
		LastName:   r.LastName,
		LastUpdate: r.LastUpdate.Time,
	}
}

func toActorModels(rows []sqlcgen.Actor) []model.Actor {
	actors := make([]model.Actor, len(rows))
	for i, r := range rows {
		actors[i] = toActorModel(r)
	}
	return actors
}
