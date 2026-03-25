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

-- name: GetCustomerTotalPayments :one
SELECT COALESCE(SUM(amount), 0)::text AS total
FROM payment
WHERE customer_id = $1;

-- name: GetCustomerTotalCharges :one
SELECT COALESCE(SUM(f.rental_rate), 0)::text AS total
FROM rental r
JOIN inventory i ON i.inventory_id = r.inventory_id
JOIN film f ON f.film_id = i.film_id
WHERE r.customer_id = $1;

-- name: CountRentalsByCustomer :one
SELECT count(*) FROM rental WHERE customer_id = $1;

-- name: GetStoreRevenue :many
SELECT
  i.store_id,
  COUNT(DISTINCT p.payment_id)::int AS payment_count,
  COUNT(DISTINCT p.rental_id)::int AS rental_count,
  COALESCE(SUM(p.amount), 0)::text AS total_revenue
FROM payment p
JOIN rental r ON r.rental_id = p.rental_id
JOIN inventory i ON i.inventory_id = r.inventory_id
WHERE p.payment_date >= $1 AND p.payment_date < $2
GROUP BY i.store_id
ORDER BY i.store_id;
