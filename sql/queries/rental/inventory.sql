-- name: GetInventory :one
SELECT inventory_id, film_id, store_id, last_update
FROM inventory
WHERE inventory_id = $1;

-- name: ListInventory :many
SELECT inventory_id, film_id, store_id, last_update
FROM inventory
ORDER BY inventory_id
LIMIT $1 OFFSET $2;

-- name: CountInventory :one
SELECT count(*) FROM inventory;

-- name: ListInventoryByFilm :many
SELECT inventory_id, film_id, store_id, last_update
FROM inventory
WHERE film_id = $1
ORDER BY inventory_id
LIMIT $2 OFFSET $3;

-- name: CountInventoryByFilm :one
SELECT count(*) FROM inventory WHERE film_id = $1;

-- name: ListInventoryByStore :many
SELECT inventory_id, film_id, store_id, last_update
FROM inventory
WHERE store_id = $1
ORDER BY inventory_id
LIMIT $2 OFFSET $3;

-- name: CountInventoryByStore :one
SELECT count(*) FROM inventory WHERE store_id = $1;

-- name: ListAvailableInventory :many
SELECT i.inventory_id, i.film_id, i.store_id, i.last_update
FROM inventory i
WHERE i.film_id = $1
  AND i.store_id = $2
  AND NOT EXISTS (
      SELECT 1 FROM rental r
      WHERE r.inventory_id = i.inventory_id AND r.return_date IS NULL
  )
ORDER BY i.inventory_id
LIMIT $3 OFFSET $4;

-- name: CountAvailableInventory :one
SELECT count(*)
FROM inventory i
WHERE i.film_id = $1
  AND i.store_id = $2
  AND NOT EXISTS (
      SELECT 1 FROM rental r
      WHERE r.inventory_id = i.inventory_id AND r.return_date IS NULL
  );

-- name: CreateInventory :one
INSERT INTO inventory (film_id, store_id)
VALUES ($1, $2)
RETURNING inventory_id, film_id, store_id, last_update;

-- name: DeleteInventory :exec
DELETE FROM inventory WHERE inventory_id = $1;
