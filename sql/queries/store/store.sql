-- name: GetStore :one
SELECT store_id, manager_staff_id, address_id, last_update
FROM store
WHERE store_id = $1;

-- name: ListStores :many
SELECT store_id, manager_staff_id, address_id, last_update
FROM store
ORDER BY store_id
LIMIT $1 OFFSET $2;

-- name: CountStores :one
SELECT count(*) FROM store;

-- name: CreateStore :one
INSERT INTO store (manager_staff_id, address_id)
VALUES ($1, $2)
RETURNING store_id, manager_staff_id, address_id, last_update;

-- name: UpdateStore :one
UPDATE store
SET manager_staff_id = $2, address_id = $3
WHERE store_id = $1
RETURNING store_id, manager_staff_id, address_id, last_update;

-- name: DeleteStore :exec
DELETE FROM store WHERE store_id = $1;
