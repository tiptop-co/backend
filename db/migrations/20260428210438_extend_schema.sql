-- +goose Up
-- +goose StatementBegin

ALTER TABLE venues ADD COLUMN manager_id text;
ALTER TABLE waiter_requests ADD COLUMN venue_id text;

CREATE INDEX IF NOT EXISTS idx_orders_table_id           ON orders        (table_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id      ON order_items   (order_id);
CREATE INDEX IF NOT EXISTS idx_transactions_order_id     ON transactions  (order_id);
CREATE INDEX IF NOT EXISTS idx_transaction_items_tx      ON transaction_items (transaction_id);
CREATE INDEX IF NOT EXISTS idx_dishes_category_id        ON dishes        (category_id);
CREATE INDEX IF NOT EXISTS idx_dishes_venue_id           ON dishes        (venue_id);
CREATE INDEX IF NOT EXISTS idx_menu_categories_venue_id  ON menu_categories (venue_id);
CREATE INDEX IF NOT EXISTS idx_tables_venue_id           ON tables        (venue_id);
CREATE INDEX IF NOT EXISTS idx_tables_waiter_id          ON tables        (waiter_id);
CREATE INDEX IF NOT EXISTS idx_waiter_requests_table_id  ON waiter_requests (table_id);
CREATE INDEX IF NOT EXISTS idx_waiter_requests_waiter_id ON waiter_requests (waiter_id);
CREATE INDEX IF NOT EXISTS idx_waiter_requests_venue_id  ON waiter_requests (venue_id);
CREATE INDEX IF NOT EXISTS idx_users_venue_id            ON users         (venue_id);
CREATE INDEX IF NOT EXISTS idx_venues_manager_id         ON venues        (manager_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_venues_manager_id;
DROP INDEX IF EXISTS idx_users_venue_id;
DROP INDEX IF EXISTS idx_waiter_requests_venue_id;
DROP INDEX IF EXISTS idx_waiter_requests_waiter_id;
DROP INDEX IF EXISTS idx_waiter_requests_table_id;
DROP INDEX IF EXISTS idx_tables_waiter_id;
DROP INDEX IF EXISTS idx_tables_venue_id;
DROP INDEX IF EXISTS idx_menu_categories_venue_id;
DROP INDEX IF EXISTS idx_dishes_venue_id;
DROP INDEX IF EXISTS idx_dishes_category_id;
DROP INDEX IF EXISTS idx_transaction_items_tx;
DROP INDEX IF EXISTS idx_transactions_order_id;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_orders_table_id;

ALTER TABLE waiter_requests DROP COLUMN IF EXISTS venue_id;
ALTER TABLE venues DROP COLUMN IF EXISTS manager_id;

-- +goose StatementEnd
