-- name: CreateVehicle :one
INSERT INTO vehicles (id, driver_id, license_plate, model, image_url, capacity)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *; -- returns the created vehicle

-- name: GetVehicleByID :one
SELECT * FROM vehicles WHERE id = $1;

-- name: GetVehiclesByDriverID :many
SELECT * FROM vehicles WHERE driver_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: GetVehicleByLicensePlate :one
SELECT * FROM vehicles WHERE license_plate = $1;

-- name: UpdateVehicle :one
UPDATE vehicles
SET 
    model = COALESCE($2, model),
    image_url = COALESCE($3, image_url),
    capacity = COALESCE($4, capacity),
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: DeleteVehicle :exec
DELETE FROM vehicles WHERE id = $1;