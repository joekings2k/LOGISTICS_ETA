-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING *; -- returns the created user

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET name = $2, email = $3, password_hash = $4, role = $5, updated_at = NOW()
WHERE id = $1
RETURNING *; -- returns the updated user

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateUserPartial :one
UPDATE users
SET
  name = COALESCE(sqlc.narg('name'), name),
  email = COALESCE(sqlc.narg('email'), email),
  password_hash = COALESCE(sqlc.narg('password_hash'), password_hash),
  role = COALESCE(sqlc.narg('role'), role)
WHERE id = sqlc.arg('id')
RETURNING *;