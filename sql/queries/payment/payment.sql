-- name: GetPayment :one
SELECT payment_id, customer_id, staff_id, rental_id, amount, payment_date
FROM payment
WHERE payment_id = $1;

-- name: ListPayments :many
SELECT payment_id, customer_id, staff_id, rental_id, amount, payment_date
FROM payment
ORDER BY payment_date DESC, payment_id DESC
LIMIT $1 OFFSET $2;

-- name: CountPayments :one
SELECT count(*) FROM payment;

-- name: ListPaymentsByCustomer :many
SELECT payment_id, customer_id, staff_id, rental_id, amount, payment_date
FROM payment
WHERE customer_id = $1
ORDER BY payment_date DESC, payment_id DESC
LIMIT $2 OFFSET $3;

-- name: CountPaymentsByCustomer :one
SELECT count(*) FROM payment WHERE customer_id = $1;

-- name: ListPaymentsByStaff :many
SELECT payment_id, customer_id, staff_id, rental_id, amount, payment_date
FROM payment
WHERE staff_id = $1
ORDER BY payment_date DESC, payment_id DESC
LIMIT $2 OFFSET $3;

-- name: CountPaymentsByStaff :one
SELECT count(*) FROM payment WHERE staff_id = $1;

-- name: ListPaymentsByRental :many
SELECT payment_id, customer_id, staff_id, rental_id, amount, payment_date
FROM payment
WHERE rental_id = $1
ORDER BY payment_date DESC, payment_id DESC
LIMIT $2 OFFSET $3;

-- name: CountPaymentsByRental :one
SELECT count(*) FROM payment WHERE rental_id = $1;

-- name: ListPaymentsByDateRange :many
SELECT payment_id, customer_id, staff_id, rental_id, amount, payment_date
FROM payment
WHERE payment_date >= $1 AND payment_date < $2
ORDER BY payment_date DESC, payment_id DESC
LIMIT $3 OFFSET $4;

-- name: CountPaymentsByDateRange :one
SELECT count(*) FROM payment
WHERE payment_date >= $1 AND payment_date < $2;

-- name: CreatePayment :one
INSERT INTO payment (customer_id, staff_id, rental_id, amount, payment_date)
VALUES ($1, $2, $3, $4, now())
RETURNING payment_id, customer_id, staff_id, rental_id, amount, payment_date;

-- name: DeletePayment :exec
DELETE FROM payment WHERE payment_id = $1;

-- name: GetCustomerName :one
SELECT first_name || ' ' || last_name AS full_name
FROM customer
WHERE customer_id = $1;

-- name: GetStaffName :one
SELECT first_name || ' ' || last_name AS full_name
FROM staff
WHERE staff_id = $1;

-- name: GetRentalDate :one
SELECT rental_date
FROM rental
WHERE rental_id = $1;
