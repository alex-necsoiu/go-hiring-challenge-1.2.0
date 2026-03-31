-- Add name column to products table
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS name VARCHAR(256) NOT NULL DEFAULT 'Product';
