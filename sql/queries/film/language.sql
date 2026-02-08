-- name: GetLanguage :one
SELECT language_id, name, last_update
FROM language
WHERE language_id = $1;

-- name: ListLanguages :many
SELECT language_id, name, last_update
FROM language
ORDER BY name;

-- name: CountLanguages :one
SELECT count(*) FROM language;
