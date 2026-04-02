-- Strengthen products.category_id constraints (NOT NULL and indexed)

-- Add index on category_id for query performance
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);

-- Backfill strategy: In a seed-based workflow, products are inserted with category data available,
-- so no NULL values should exist. However, for defensive programming, we check for NULLs.
-- If NULLs are found, they represent a data quality issue that requires business input:
-- - Which category should unassigned products belong to?
-- - Should products without categories be deleted?
-- Rather than silently assigning to an arbitrary category, we document the constraint
-- and rely on the data pipeline (seed) to maintain integrity.

-- For environments that require backfilling (e.g., migrating existing databases),
-- uncomment one of these options based on business requirements:

-- Option 1: Fail if any NULL values exist (safest for production; enforces data quality)
-- Uncomment if you want to detect unhandled data issues:
-- DO $$ BEGIN
--   IF EXISTS(SELECT 1 FROM products WHERE category_id IS NULL) THEN
--     RAISE EXCEPTION 'Data integrity violation: Products with NULL category_id detected. A business decision is needed to assign these products to a category.';
--   END IF;
-- END $$;

-- Option 2: Assign to a specific category (only if explicitly decided)
-- Uncomment only if you have received business approval for the default assignment:
-- UPDATE products 
-- SET category_id = (SELECT id FROM categories WHERE code = 'CLOTHING' LIMIT 1)
-- WHERE category_id IS NULL AND EXISTS (SELECT 1 FROM categories WHERE code = 'CLOTHING');

-- Set category_id to NOT NULL (this finalizes the schema at the end of migration sequence)
ALTER TABLE products
    ALTER COLUMN category_id SET NOT NULL;
