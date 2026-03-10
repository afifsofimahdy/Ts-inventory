ALTER TABLE stock_adjustments
  ADD COLUMN IF NOT EXISTS source TEXT,
  ADD COLUMN IF NOT EXISTS ref_code TEXT,
  ADD COLUMN IF NOT EXISTS bucket TEXT;

UPDATE stock_adjustments
SET source = COALESCE(source, 'MANUAL'),
    bucket = COALESCE(bucket, 'PHYSICAL');
