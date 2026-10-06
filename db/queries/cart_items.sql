-- name: ListLines :many
SELECT * FROM cart_items WHERE cart_id = $1 ORDER BY id;

-- name: GetLine :one
SELECT * FROM cart_items WHERE id = $1 AND cart_id = $2;

-- name: UpsertLine :one
INSERT INTO cart_items (cart_id, catalog_item_id, name, price_cents, currency, qty)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (cart_id, catalog_item_id) DO UPDATE SET
    qty = cart_items.qty + EXCLUDED.qty,
    name = EXCLUDED.name, price_cents = EXCLUDED.price_cents,
    currency = EXCLUDED.currency, updated_at = NOW()
RETURNING *;

-- name: SetLineQty :one
UPDATE cart_items SET qty = $1, name = $2, price_cents = $3, currency = $4,
    updated_at = NOW()
WHERE id = $5 AND cart_id = $6 RETURNING *;

-- name: DeleteLine :exec
DELETE FROM cart_items WHERE id = $1 AND cart_id = $2;
