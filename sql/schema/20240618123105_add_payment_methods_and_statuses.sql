-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- +goose StatementBegin
CREATE TABLE payment_methods (
    id SERIAL PRIMARY KEY,
    method VARCHAR(50) UNIQUE NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO payment_methods (method) VALUES ('cash'), ('credit_card'), ('debit_card'), ('paypal'), ('mpesa');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE payment_statuses (
    id SERIAL PRIMARY KEY,
    status VARCHAR(50) UNIQUE NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO payment_statuses (status) VALUES ('pending'), ('paid'), ('failed');
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_payments
ADD COLUMN payment_method_id INT,
ADD COLUMN payment_status_id INT;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE order_payments SET payment_method_id = (SELECT id FROM payment_methods WHERE method = order_payments.method);
UPDATE order_payments SET payment_status_id = (SELECT id FROM payment_statuses WHERE status = order_payments.status);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_payments
DROP COLUMN method,
DROP COLUMN status;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_payments
ADD CONSTRAINT fk_payment_method FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id),
ADD CONSTRAINT fk_payment_status FOREIGN KEY (payment_status_id) REFERENCES payment_statuses(id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_order_payments_payment_method_id ON order_payments(payment_method_id);
CREATE INDEX idx_order_payments_payment_status_id ON order_payments(payment_status_id);
-- +goose StatementEnd

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

-- +goose StatementBegin
ALTER TABLE order_payments
DROP CONSTRAINT fk_payment_method,
DROP CONSTRAINT fk_payment_status;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_payments
ADD COLUMN method VARCHAR(50),
ADD COLUMN status VARCHAR(50);
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE order_payments SET method = (SELECT method FROM payment_methods WHERE id = order_payments.payment_method_id);
UPDATE order_payments SET status = (SELECT status FROM payment_statuses WHERE id = order_payments.payment_status_id);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_payments
DROP COLUMN payment_method_id,
DROP COLUMN payment_status_id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX idx_order_payments_payment_method_id;
DROP INDEX idx_order_payments_payment_status_id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE payment_methods;
DROP TABLE payment_statuses;
-- +goose StatementEnd
