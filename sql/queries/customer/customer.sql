-- name: GetCustomer :one
SELECT customer_id, store_id, first_name, last_name, email,
       address_id, activebool, create_date, last_update, active
FROM customer
WHERE customer_id = $1;

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
