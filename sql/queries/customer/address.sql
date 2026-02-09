-- name: GetAddress :one
SELECT address_id, address, address2, district, city_id, postal_code, phone, last_update
FROM address
WHERE address_id = $1;

-- name: ListAddresses :many
SELECT address_id, address, address2, district, city_id, postal_code, phone, last_update
FROM address
ORDER BY address_id
LIMIT $1 OFFSET $2;

-- name: CountAddresses :one
SELECT count(*) FROM address;

-- name: CreateAddress :one
INSERT INTO address (address, address2, district, city_id, postal_code, phone)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING address_id, address, address2, district, city_id, postal_code, phone, last_update;

-- name: UpdateAddress :one
UPDATE address
SET address = $2,
    address2 = $3,
    district = $4,
    city_id = $5,
    postal_code = $6,
    phone = $7
WHERE address_id = $1
RETURNING address_id, address, address2, district, city_id, postal_code, phone, last_update;

-- name: DeleteAddress :exec
DELETE FROM address WHERE address_id = $1;
