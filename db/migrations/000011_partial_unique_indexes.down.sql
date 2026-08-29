-- Restoring the plain constraints requires no duplicate to exist among the
-- soft deleted rows either; clean them up first if this rollback fails.

DROP INDEX IF EXISTS uniq_refresh_tokens_token;
ALTER TABLE refresh_tokens ADD CONSTRAINT refresh_tokens_token_key UNIQUE (token);

DROP INDEX IF EXISTS uniq_products_sku;
ALTER TABLE products ADD CONSTRAINT products_sku_key UNIQUE (sku);

DROP INDEX IF EXISTS uniq_users_email;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);

DROP INDEX IF EXISTS uniq_carts_user_id;
ALTER TABLE carts ADD CONSTRAINT carts_user_id_key UNIQUE (user_id);

DROP INDEX IF EXISTS uniq_cart_items_cart_id_product_id;
ALTER TABLE cart_items ADD CONSTRAINT cart_items_cart_id_product_id_key UNIQUE (cart_id, product_id);
