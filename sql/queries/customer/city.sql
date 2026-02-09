-- name: GetCity :one
SELECT city_id, city, country_id, last_update
FROM city
WHERE city_id = $1;

-- name: ListCities :many
SELECT city_id, city, country_id, last_update
FROM city
ORDER BY city_id
LIMIT $1 OFFSET $2;

-- name: CountCities :one
SELECT count(*) FROM city;
