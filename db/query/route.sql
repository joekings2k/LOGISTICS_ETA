-- name: CreateRoute :one
INSERT INTO routes (
    id,
    driver_id,
    vehicle_id,
    origin_address,
    origin_lat,
    origin_lng,
    destination_address,
    destination_lat,
    destination_lng,
    estimated_distance_km,
    estimated_duration_min,
    status
)
VALUES (
    $1, $2, $3,
    $4, $5, $6,
    $7, $8, $9,
    $10, $11, $12
)
RETURNING *;


-- name: GetRouteByID :one
SELECT * FROM routes WHERE id = $1;

-- name: GetRoutesByDriverID :many
SELECT * FROM routes
WHERE driver_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateRouteStatus :one
UPDATE routes
SET status = COALESCE(NULLIF($2, ''), status),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: updateRouteActualDuration :one
UPDATE routes
SET actual_duration_min = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *; -- when the route is completed


-- name: DeleteRoute :exec
DELETE FROM routes WHERE id = $1;

-- name: ListRoutes :many
SELECT * FROM routes
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;


