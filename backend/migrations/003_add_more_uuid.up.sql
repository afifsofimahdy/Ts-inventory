CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE customers ADD COLUMN IF NOT EXISTS uuid UUID NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE inventory ADD COLUMN IF NOT EXISTS uuid UUID NOT NULL DEFAULT gen_random_uuid();

CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_uuid ON customers(uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_uuid ON inventory(uuid);
