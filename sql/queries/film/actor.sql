-- name: GetActor :one
SELECT actor_id, first_name, last_name, last_update
FROM actor
WHERE actor_id = $1;

-- name: ListActors :many
SELECT actor_id, first_name, last_name, last_update
FROM actor
ORDER BY last_name, first_name
LIMIT $1 OFFSET $2;

-- name: CountActors :one
SELECT count(*) FROM actor;

-- name: ListActorsByFilm :many
-- Returns all actors for a given film (no pagination needed, typically small set).
SELECT a.actor_id, a.first_name, a.last_name, a.last_update
FROM actor a
JOIN film_actor fa ON a.actor_id = fa.actor_id
WHERE fa.film_id = $1
ORDER BY a.last_name, a.first_name;

-- name: CreateActor :one
INSERT INTO actor (first_name, last_name)
VALUES ($1, $2)
RETURNING actor_id, first_name, last_name, last_update;

-- name: UpdateActor :one
UPDATE actor
SET first_name = $2, last_name = $3
WHERE actor_id = $1
RETURNING actor_id, first_name, last_name, last_update;

-- name: DeleteActor :exec
DELETE FROM actor WHERE actor_id = $1;
