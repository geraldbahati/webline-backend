-- +goose Up
-- Adding index for performance improvement

CREATE INDEX IF NOT EXISTS idx_order_item_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX IF NOT EXISTS idx_order_payments_order_id ON order_payments(order_id);
CREATE INDEX IF NOT EXISTS idx_order_shipments_order_id ON order_shipments(order_id);

-- +goose Down
-- Removing indexes

DROP INDEX IF EXISTS idx_order_item_order_id;
DROP INDEX IF EXISTS idx_order_status_history_order_id;
DROP INDEX IF EXISTS idx_order_payments_order_id;
DROP INDEX IF EXISTS idx_order_shipments_order_id;
