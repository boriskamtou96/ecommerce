-- A plain UNIQUE constraint counts soft deleted rows, but GORM hides them
-- behind "deleted_at IS NULL". A row that the application believes is gone
-- therefore still blocks the insert of its replacement (SQLSTATE 23505).
--
-- Replacing each constraint with a unique index restricted to live rows
-- keeps uniqueness where it matters and lets a soft deleted row be
-- superseded: re-adding a product to a cart after checkout, reusing the SKU
-- of a deleted product, registering again with the email of a deleted
-- account.

ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_cart_id_product_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_cart_items_cart_id_product_id
    ON cart_items (cart_id, product_id)
    WHERE deleted_at IS NULL;

ALTER TABLE carts DROP CONSTRAINT IF EXISTS carts_user_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_carts_user_id
    ON carts (user_id)
    WHERE deleted_at IS NULL;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_users_email
    ON users (email)
    WHERE deleted_at IS NULL;

ALTER TABLE products DROP CONSTRAINT IF EXISTS products_sku_key;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_products_sku
    ON products (sku)
    WHERE deleted_at IS NULL;

ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS refresh_tokens_token_key;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_refresh_tokens_token
    ON refresh_tokens (token)
    WHERE deleted_at IS NULL;
