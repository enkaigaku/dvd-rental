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

-- name: GetCustomerBalance :one
SELECT c.customer_id,
       charges.total::text AS total_charges,
       payments.total::text AS total_payments,
       (charges.total - payments.total)::text AS balance,
       charges.rental_count, payments.payment_count
FROM customer c
CROSS JOIN LATERAL (
    SELECT COALESCE(SUM(rc.amount), 0) AS total, COUNT(*)::int AS rental_count
    FROM rental_charge rc WHERE rc.customer_id = c.customer_id
) charges
CROSS JOIN LATERAL (
    SELECT COALESCE(SUM(amount), 0) AS total, COUNT(*)::int AS payment_count
    FROM payment WHERE customer_id = c.customer_id
) payments
WHERE c.customer_id = $1;

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
