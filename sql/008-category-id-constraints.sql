-- Strengthen products.category_id constraints (NOT NULL and indexed)

-- Add index on category_id for query performance
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);

-- Backfill any missing category_id values (defensive, should not be needed after seed)
-- Assign to first category if none assigned (fallback to category id 1 if it exists)
UPDATE products 
SET category_id = (SELECT id FROM categories LIMIT 1)
WHERE category_id IS NULL AND EXISTS (SELECT 1 FROM categories);

-- Set category_id to NOT NULL
ALTER TABLE products
    ALTER COLUMN category_id SET NOT NULL;
