CREATE TABLE IF NOT EXISTS customers (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS items (
  id SERIAL PRIMARY KEY,
  sku TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  customer_id INT NOT NULL REFERENCES customers(id)
);

CREATE TABLE IF NOT EXISTS inventory (
  item_id INT PRIMARY KEY REFERENCES items(id),
  physical_qty BIGINT NOT NULL DEFAULT 0,
  available_qty BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stock_in (
  id SERIAL PRIMARY KEY,
  code TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stock_in_items (
  id SERIAL PRIMARY KEY,
  stock_in_id INT NOT NULL REFERENCES stock_in(id),
  item_id INT NOT NULL REFERENCES items(id),
  qty BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS stock_in_logs (
  id SERIAL PRIMARY KEY,
  stock_in_id INT NOT NULL REFERENCES stock_in(id),
  from_status TEXT,
  to_status TEXT NOT NULL,
  at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stock_out (
  id SERIAL PRIMARY KEY,
  code TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stock_out_items (
  id SERIAL PRIMARY KEY,
  stock_out_id INT NOT NULL REFERENCES stock_out(id),
  item_id INT NOT NULL REFERENCES items(id),
  qty BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS stock_out_logs (
  id SERIAL PRIMARY KEY,
  stock_out_id INT NOT NULL REFERENCES stock_out(id),
  from_status TEXT,
  to_status TEXT NOT NULL,
  at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stock_adjustments (
  id SERIAL PRIMARY KEY,
  item_id INT NOT NULL REFERENCES items(id),
  delta BIGINT NOT NULL,
  reason TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- seed sample data
INSERT INTO customers (name) VALUES ('Acme'), ('Umbrella') ON CONFLICT DO NOTHING;

INSERT INTO items (sku, name, customer_id) VALUES
  ('SKU-001', 'Widget A', 1),
  ('SKU-002', 'Widget B', 1),
  ('SKU-003', 'Gadget C', 2)
ON CONFLICT DO NOTHING;

INSERT INTO inventory (item_id, physical_qty, available_qty)
SELECT i.id, 100, 100 FROM items i
ON CONFLICT (item_id) DO NOTHING;
