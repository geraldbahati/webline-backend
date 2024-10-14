-- Example using Goose migration to remove session_reference and add guest_id

-- +goose Up
ALTER TABLE shopping_carts
    DROP CONSTRAINT IF EXISTS fk_shopping_carts_session_reference,
    DROP COLUMN IF EXISTS session_reference,
    ADD COLUMN guest_id uuid UNIQUE;

ALTER TABLE shopping_carts
    ADD CONSTRAINT chk_cart_owner
    CHECK (
        (user_id IS NOT NULL AND guest_id IS NULL) OR
        (user_id IS NULL AND guest_id IS NOT NULL)
    );

-- Optionally, create an index on guest_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_shopping_carts_guest_id ON shopping_carts(guest_id);

-- +goose Down
ALTER TABLE shopping_carts
    DROP CONSTRAINT IF EXISTS chk_cart_owner,
    DROP COLUMN IF EXISTS guest_id,
    ADD COLUMN session_reference uuid;

ALTER TABLE shopping_carts
    ADD CONSTRAINT fk_shopping_carts_session_reference FOREIGN KEY (session_reference) REFERENCES sessions(session_id) ON DELETE SET NULL;

DROP INDEX IF EXISTS idx_shopping_carts_guest_id;
