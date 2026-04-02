CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    code VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(256) NOT NULL
);

-- Add category_id without NOT NULL constraint initially (nullable)
-- This allows the column to exist before categories are populated.
-- NOT NULL constraint is hardened in migration 008 after data integrity is ensured.
-- This phased approach ensures migrations remain idempotent and support partially-populated schemas.
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS category_id INTEGER REFERENCES categories(id);
