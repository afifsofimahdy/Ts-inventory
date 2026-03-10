CREATE TABLE IF NOT EXISTS categories (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

ALTER TABLE items ADD COLUMN IF NOT EXISTS category_id INT REFERENCES categories(id);

CREATE INDEX IF NOT EXISTS idx_items_category_id ON items(category_id);
