CREATE TABLE IF NOT EXISTS roles (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users (
    id       SERIAL PRIMARY KEY,
    email    TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    role_id  INT NOT NULL REFERENCES roles (id) ON DELETE RESTRICT,
    photo    TEXT
);

CREATE TABLE IF NOT EXISTS categories (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_by INT REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by INT REFERENCES users (id),
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS products (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    price      BIGINT NOT NULL DEFAULT 0,
    category   INT NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    in_stock   BOOLEAN NOT NULL DEFAULT FALSE,
    quantity   INT NOT NULL DEFAULT 0,
    img        TEXT,
    created_by INT REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by INT REFERENCES users (id),
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS purchases (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users (id),
    product_id INT NOT NULL REFERENCES products (id),
    quantity   INT NOT NULL,
    unit_price BIGINT NOT NULL,
    discount   BIGINT NOT NULL DEFAULT 0,
    total      BIGINT NOT NULL,
    wallet_ref TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE categories ADD COLUMN IF NOT EXISTS created_by INT REFERENCES users (id);
ALTER TABLE categories ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE categories ADD COLUMN IF NOT EXISTS updated_by INT REFERENCES users (id);
ALTER TABLE categories ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

ALTER TABLE products ADD COLUMN IF NOT EXISTS created_by INT REFERENCES users (id);
ALTER TABLE products ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE products ADD COLUMN IF NOT EXISTS updated_by INT REFERENCES users (id);
ALTER TABLE products ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;