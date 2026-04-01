-- Strengthen products.code constraints (NOT NULL and UNIQUE)

-- First, ensure all existing products have a code (shouldn't be needed if FK constraints work, but safe for migration)
UPDATE products SET code = 'PROD-' || id WHERE code IS NULL;

-- Set code to NOT NULL
ALTER TABLE products
    ALTER COLUMN code SET NOT NULL;

-- Add UNIQUE constraint on code
ALTER TABLE products
    ADD CONSTRAINT products_code_unique UNIQUE (code);
