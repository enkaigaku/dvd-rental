-- name: AddCategoryToFilm :exec
INSERT INTO film_category (film_id, category_id)
VALUES ($1, $2);

-- name: RemoveCategoryFromFilm :exec
DELETE FROM film_category
WHERE film_id = $1 AND category_id = $2;
