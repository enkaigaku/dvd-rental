-- name: GetCategory :one
SELECT category_id, name, last_update
FROM category
WHERE category_id = $1;

-- name: ListCategories :many
SELECT category_id, name, last_update
FROM category
ORDER BY name;

-- name: CountCategories :one
SELECT count(*) FROM category;

-- name: ListCategoriesByFilm :many
-- Returns all categories for a given film (no pagination needed, typically 1-2 categories).
SELECT c.category_id, c.name, c.last_update
FROM category c
JOIN film_category fc ON c.category_id = fc.category_id
WHERE fc.film_id = $1
ORDER BY c.name;
