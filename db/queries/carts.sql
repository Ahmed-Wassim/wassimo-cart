-- name: GetCartByUser :one
SELECT * FROM carts WHERE user_id = $1;

-- name: CreateCart :one
INSERT INTO carts (user_id, branch_id) VALUES ($1, $2) RETURNING *;

-- name: GetCartByID :one
SELECT * FROM carts WHERE id = $1;

-- name: DeleteCart :exec
DELETE FROM carts WHERE id = $1;
