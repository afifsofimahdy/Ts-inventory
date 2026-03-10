CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE items ADD COLUMN IF NOT EXISTS uuid UUID NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE stock_in ADD COLUMN IF NOT EXISTS uuid UUID NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE stock_out ADD COLUMN IF NOT EXISTS uuid UUID NOT NULL DEFAULT gen_random_uuid();

CREATE UNIQUE INDEX IF NOT EXISTS idx_items_uuid ON items(uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_in_uuid ON stock_in(uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_out_uuid ON stock_out(uuid);
