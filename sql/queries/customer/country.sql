-- name: GetCountry :one
SELECT country_id, country, last_update
FROM country
WHERE country_id = $1;

-- name: ListCountries :many
SELECT country_id, country, last_update
FROM country
ORDER BY country_id
LIMIT $1 OFFSET $2;

-- name: CountCountries :one
SELECT count(*) FROM country;
