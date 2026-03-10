DROP INDEX IF EXISTS idx_inventory_uuid;
DROP INDEX IF EXISTS idx_customers_uuid;

ALTER TABLE inventory DROP COLUMN IF EXISTS uuid;
ALTER TABLE customers DROP COLUMN IF EXISTS uuid;
