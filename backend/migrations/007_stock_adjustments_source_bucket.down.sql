ALTER TABLE stock_adjustments
  DROP COLUMN IF EXISTS source,
  DROP COLUMN IF EXISTS ref_code,
  DROP COLUMN IF EXISTS bucket;
