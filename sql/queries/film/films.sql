-- name: GetFilm :one
SELECT film_id, title, description, release_year, language_id,
       original_language_id, rental_duration, rental_rate, length,
       replacement_cost, rating, special_features, last_update
FROM film
WHERE film_id = $1;

-- name: ListFilms :many
SELECT film_id, title, description, release_year, language_id,
       original_language_id, rental_duration, rental_rate, length,
       replacement_cost, rating, special_features, last_update
FROM film
ORDER BY title
LIMIT $1 OFFSET $2;

-- name: CountFilms :one
SELECT count(*) FROM film;

-- name: SearchFilms :many
SELECT film_id, title, description, release_year, language_id,
       original_language_id, rental_duration, rental_rate, length,
       replacement_cost, rating, special_features, last_update
FROM film
WHERE fulltext @@ plainto_tsquery('english', $1)
ORDER BY ts_rank(fulltext, plainto_tsquery('english', $1)) DESC
LIMIT $2 OFFSET $3;

-- name: CountSearchFilms :one
SELECT count(*)
FROM film
WHERE fulltext @@ plainto_tsquery('english', $1);

-- name: ListFilmsByCategory :many
SELECT f.film_id, f.title, f.description, f.release_year, f.language_id,
       f.original_language_id, f.rental_duration, f.rental_rate, f.length,
       f.replacement_cost, f.rating, f.special_features, f.last_update
FROM film f
JOIN film_category fc ON f.film_id = fc.film_id
WHERE fc.category_id = $1
ORDER BY f.title
LIMIT $2 OFFSET $3;

-- name: CountFilmsByCategory :one
SELECT count(*)
FROM film f
JOIN film_category fc ON f.film_id = fc.film_id
WHERE fc.category_id = $1;

-- name: ListFilmsByActor :many
SELECT f.film_id, f.title, f.description, f.release_year, f.language_id,
       f.original_language_id, f.rental_duration, f.rental_rate, f.length,
       f.replacement_cost, f.rating, f.special_features, f.last_update
FROM film f
JOIN film_actor fa ON f.film_id = fa.film_id
WHERE fa.actor_id = $1
ORDER BY f.title
LIMIT $2 OFFSET $3;

-- name: CountFilmsByActor :one
SELECT count(*)
FROM film f
JOIN film_actor fa ON f.film_id = fa.film_id
WHERE fa.actor_id = $1;

-- name: CreateFilm :one
INSERT INTO film (title, description, release_year, language_id,
                  original_language_id, rental_duration, rental_rate,
                  length, replacement_cost, rating, special_features)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING film_id, title, description, release_year, language_id,
          original_language_id, rental_duration, rental_rate, length,
          replacement_cost, rating, special_features, last_update;

-- name: UpdateFilm :one
UPDATE film
SET title = $2, description = $3, release_year = $4, language_id = $5,
    original_language_id = $6, rental_duration = $7, rental_rate = $8,
    length = $9, replacement_cost = $10, rating = $11, special_features = $12
WHERE film_id = $1
RETURNING film_id, title, description, release_year, language_id,
          original_language_id, rental_duration, rental_rate, length,
          replacement_cost, rating, special_features, last_update;

-- name: DeleteFilm :exec
DELETE FROM film WHERE film_id = $1;
