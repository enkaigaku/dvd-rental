-- name: GetStaff :one
-- Returns full staff record including picture (but not password_hash).
SELECT staff_id, first_name, last_name, address_id, email,
       store_id, active, username, last_update, picture
FROM staff
WHERE staff_id = $1;

-- name: GetStaffByUsername :one
-- Returns staff record with password_hash for authentication (but not picture).
SELECT staff_id, first_name, last_name, address_id, email,
       store_id, active, username, password_hash, last_update
FROM staff
WHERE username = $1;

-- name: ListStaff :many
-- Returns staff list without picture or password_hash.
SELECT staff_id, first_name, last_name, address_id, email,
       store_id, active, username, last_update
FROM staff
ORDER BY staff_id
LIMIT $1 OFFSET $2;

-- name: ListActiveStaff :many
SELECT staff_id, first_name, last_name, address_id, email,
       store_id, active, username, last_update
FROM staff
WHERE active = true
ORDER BY staff_id
LIMIT $1 OFFSET $2;

-- name: ListStaffByStore :many
SELECT staff_id, first_name, last_name, address_id, email,
       store_id, active, username, last_update
FROM staff
WHERE store_id = $1
ORDER BY staff_id
LIMIT $2 OFFSET $3;

-- name: ListActiveStaffByStore :many
SELECT staff_id, first_name, last_name, address_id, email,
       store_id, active, username, last_update
FROM staff
WHERE store_id = $1 AND active = true
ORDER BY staff_id
LIMIT $2 OFFSET $3;

-- name: CountStaff :one
SELECT count(*) FROM staff;

-- name: CountActiveStaff :one
SELECT count(*) FROM staff WHERE active = true;

-- name: CountStaffByStore :one
SELECT count(*) FROM staff WHERE store_id = $1;

-- name: CountActiveStaffByStore :one
SELECT count(*) FROM staff WHERE store_id = $1 AND active = true;

-- name: CreateStaff :one
INSERT INTO staff (first_name, last_name, address_id, email, store_id, username, password_hash)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING staff_id, first_name, last_name, address_id, email,
          store_id, active, username, last_update;

-- name: UpdateStaff :one
UPDATE staff
SET first_name = $2, last_name = $3, address_id = $4,
    email = $5, store_id = $6, username = $7
WHERE staff_id = $1
RETURNING staff_id, first_name, last_name, address_id, email,
          store_id, active, username, last_update;

-- name: DeactivateStaff :one
UPDATE staff
SET active = false
WHERE staff_id = $1
RETURNING staff_id, first_name, last_name, address_id, email,
          store_id, active, username, last_update;

-- name: UpdateStaffPassword :exec
UPDATE staff
SET password_hash = $2
WHERE staff_id = $1;
