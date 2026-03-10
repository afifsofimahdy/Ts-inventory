DROP INDEX IF EXISTS idx_items_category_id;
ALTER TABLE items DROP COLUMN IF EXISTS category_id;
DROP TABLE IF EXISTS categories;
