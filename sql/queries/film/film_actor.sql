-- name: AddActorToFilm :exec
INSERT INTO film_actor (actor_id, film_id)
VALUES ($1, $2);

-- name: RemoveActorFromFilm :exec
DELETE FROM film_actor
WHERE actor_id = $1 AND film_id = $2;
