-- +goose Up
-- +goose StatementBegin

CREATE TABLE storages (
    id BIGSERIAL PRIMARY KEY,
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
);

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    vendor VARCHAR(20) UNIQUE NOT NULL, -- уникальный код из условия
    size VARCHAR NOT NULL
);

CREATE INDEX products_vendor_idx ON products(vendor);

CREATE TABLE stored_products (
    id BIGSERIAL PRIMARY KEY,
    storage_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    amount INT NOT NULL DEFAULT 1,

    UNIQUE (storage_id, product_id),

    FOREIGN KEY (storage_id) REFERENCES storages(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- индекс один, потому что pk накладывает unique_index
CREATE INDEX stored_products_product_id_idx ON stored_products(product_id);

CREATE TABLE product_reservations (
    id BIGSERIAL PRIMARY KEY,
    storage_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    amount INT NOT NULL DEFAULT 1,

    UNIQUE (storage_id, product_id),

    FOREIGN KEY (storage_id) REFERENCES storages(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- индекс один, потому что pk накладывает unique_index
CREATE INDEX product_reservations_product_id_idx ON product_reservations(product_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS products_vendor_idx;
DROP INDEX IF EXISTS stored_products_product_id_idx;
DROP INDEX IF EXISTS product_reservations_product_id_idx;

DROP TABLE IF EXISTS stored_products;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS storages;

-- +goose StatementEnd
