-- name: GetCustomer :one
SELECT customer_id, store_id, first_name, last_name, email,
       address_id, activebool, create_date, last_update, active
FROM customer
WHERE customer_id = $1;

-- name: GetCustomerByEmail :one
SELECT customer_id, store_id, first_name, last_name, email,
       address_id, activebool, create_date, last_update, active,
       password_hash
FROM customer
WHERE email = $1;

-- name: ListCustomers :many
SELECT customer_id, store_id, first_name, last_name, email,
       address_id, activebool, create_date, last_update, active
FROM customer
ORDER BY customer_id
LIMIT $1 OFFSET $2;

-- name: CountCustomers :one
SELECT count(*) FROM customer;

-- name: ListCustomersByStore :many
SELECT customer_id, store_id, first_name, last_name, email,
       address_id, activebool, create_date, last_update, active
FROM customer
WHERE store_id = $1
ORDER BY customer_id
LIMIT $2 OFFSET $3;

-- name: CountCustomersByStore :one
SELECT count(*) FROM customer WHERE store_id = $1;

-- name: CreateCustomer :one
INSERT INTO customer (store_id, first_name, last_name, email, address_id, activebool, active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING customer_id, store_id, first_name, last_name, email,
          address_id, activebool, create_date, last_update, active;

-- name: UpdateCustomer :one
UPDATE customer
SET store_id = $2,
    first_name = $3,
    last_name = $4,
    email = $5,
    address_id = $6,
    activebool = $7,
    active = $8
WHERE customer_id = $1
RETURNING customer_id, store_id, first_name, last_name, email,
          address_id, activebool, create_date, last_update, active;

-- name: DeleteCustomer :exec
DELETE FROM customer WHERE customer_id = $1;

-- name: CountCustomerOverdueRentals :one
SELECT count(*) FROM rental r
JOIN inventory i ON i.inventory_id = r.inventory_id
JOIN film f ON f.film_id = i.film_id
WHERE r.customer_id = $1
  AND r.return_date IS NULL
  AND r.rental_date + f.rental_duration * interval '24 hours' < now();

-- name: CountCustomerActiveRentals :one
SELECT count(*) FROM rental
WHERE customer_id = $1 AND return_date IS NULL;

-- name: GetCustomerBalance :one
SELECT (
  COALESCE((SELECT SUM(rc.amount) FROM rental_charge rc WHERE rc.customer_id = $1), 0.00)
  - COALESCE((SELECT SUM(amount) FROM payment WHERE customer_id = $1), 0.00)
)::text AS balance;

-- name: CountCustomerTotalRentals :one
SELECT count(*) FROM rental WHERE customer_id = $1;

-- name: GetCustomerTotalSpent :one
SELECT COALESCE(SUM(amount), 0.00)::text AS total FROM payment WHERE customer_id = $1;

-- name: GetCustomerFavoriteCategory :one
SELECT c.name
FROM rental r
JOIN inventory i ON i.inventory_id = r.inventory_id
JOIN film_category fc ON fc.film_id = i.film_id
JOIN category c ON c.category_id = fc.category_id
WHERE r.customer_id = $1
GROUP BY c.category_id, c.name
ORDER BY count(*) DESC, c.category_id ASC
LIMIT 1;
