-- name: GetRental :one
SELECT rental_id, rental_date, inventory_id, customer_id, return_date, staff_id, last_update
FROM rental
WHERE rental_id = $1;

-- name: ListRentals :many
SELECT rental_id, rental_date, inventory_id, customer_id, return_date, staff_id, last_update
FROM rental
ORDER BY rental_id DESC
LIMIT $1 OFFSET $2;

-- name: CountRentals :one
SELECT count(*) FROM rental;

-- name: ListRentalsByCustomer :many
SELECT rental_id, rental_date, inventory_id, customer_id, return_date, staff_id, last_update
FROM rental
WHERE customer_id = $1
ORDER BY rental_id DESC
LIMIT $2 OFFSET $3;

-- name: CountRentalsByCustomer :one
SELECT count(*) FROM rental WHERE customer_id = $1;

-- name: ListRentalsByInventory :many
SELECT rental_id, rental_date, inventory_id, customer_id, return_date, staff_id, last_update
FROM rental
WHERE inventory_id = $1
ORDER BY rental_id DESC
LIMIT $2 OFFSET $3;

-- name: CountRentalsByInventory :one
SELECT count(*) FROM rental WHERE inventory_id = $1;

-- name: ListOverdueRentals :many
SELECT rental_id, rental_date, inventory_id, customer_id, return_date, staff_id, last_update
FROM rental
WHERE return_date IS NULL
ORDER BY rental_date ASC
LIMIT $1 OFFSET $2;

-- name: CountOverdueRentals :one
SELECT count(*) FROM rental WHERE return_date IS NULL;

-- name: CreateRental :one
INSERT INTO rental (rental_date, inventory_id, customer_id, staff_id)
VALUES (now(), $1, $2, $3)
RETURNING rental_id, rental_date, inventory_id, customer_id, return_date, staff_id, last_update;

-- name: ReturnRental :one
UPDATE rental
SET return_date = now()
WHERE rental_id = $1 AND return_date IS NULL
RETURNING rental_id, rental_date, inventory_id, customer_id, return_date, staff_id, last_update;

-- name: DeleteRental :exec
DELETE FROM rental WHERE rental_id = $1;

-- name: GetCustomerName :one
SELECT first_name || ' ' || last_name AS full_name
FROM customer
WHERE customer_id = $1;

-- name: GetFilmTitleByInventory :one
SELECT f.title, i.store_id
FROM inventory i
JOIN film f ON f.film_id = i.film_id
WHERE i.inventory_id = $1;

-- name: IsInventoryAvailable :one
SELECT NOT EXISTS (
    SELECT 1 FROM rental
    WHERE inventory_id = $1 AND return_date IS NULL
) AS available;

-- name: CountActiveRentalsByCustomer :one
SELECT count(*) FROM rental
WHERE customer_id = $1 AND return_date IS NULL;

-- name: CountOverdueRentalsByCustomer :one
SELECT count(*) FROM rental r
JOIN inventory i ON i.inventory_id = r.inventory_id
JOIN film f ON f.film_id = i.film_id
WHERE r.customer_id = $1
  AND r.return_date IS NULL
  AND r.rental_date + (f.rental_duration || ' days')::interval < now() - interval '7 days';

-- name: GetFilmRentalTermsByInventory :one
SELECT f.rental_duration, f.rental_rate, f.replacement_cost, f.title, i.store_id
FROM inventory i
JOIN film f ON f.film_id = i.film_id
WHERE i.inventory_id = $1;

-- name: CreateLateFeePayment :one
INSERT INTO payment (customer_id, staff_id, rental_id, amount, payment_date)
VALUES ($1, $2, $3, $4, now())
RETURNING payment_id, customer_id, staff_id, rental_id, amount, payment_date;
