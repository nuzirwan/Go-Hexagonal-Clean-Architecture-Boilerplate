CREATE TABLE IF NOT EXISTS products (
    id         VARCHAR(26) PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    price      DECIMAL(12, 2) NOT NULL,
    currency   VARCHAR(3) NOT NULL,
    stock      INTEGER NOT NULL DEFAULT 0,
    status     VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_id_desc ON products(id DESC);
