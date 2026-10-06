-- +goose Up
CREATE TABLE IF NOT EXISTS cart_items (
    id BIGSERIAL PRIMARY KEY,
    cart_id BIGINT NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    catalog_item_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    price_cents INTEGER NOT NULL CHECK (price_cents >= 0),
    currency TEXT NOT NULL DEFAULT 'EGP',
    qty INTEGER NOT NULL CHECK (qty > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (cart_id, catalog_item_id)
);
CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);

-- +goose Down
DROP INDEX IF EXISTS idx_cart_items_cart_id;
DROP TABLE IF EXISTS cart_items;
