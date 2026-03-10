-- Customers
INSERT INTO customers (name) VALUES ('Tirtamas Retail'), ('Corporate') ON CONFLICT DO NOTHING;

-- Categories
INSERT INTO categories (name) VALUES
  ('Smartphone'),
  ('Laptop'),
  ('Tablet'),
  ('Accessories'),
  ('Audio'),
  ('Gaming'),
  ('Networking')
ON CONFLICT DO NOTHING;

-- Items (20 items)
INSERT INTO items (sku, name, customer_id, category_id) VALUES
  ('SM-001', 'Smartphone Alpha 128GB', 1, (SELECT id FROM categories WHERE name='Smartphone')),
  ('SM-002', 'Smartphone Alpha 256GB', 1, (SELECT id FROM categories WHERE name='Smartphone')),
  ('SM-003', 'Smartphone Beta 128GB', 1, (SELECT id FROM categories WHERE name='Smartphone')),
  ('LP-001', 'Laptop Pro 14" i5', 2, (SELECT id FROM categories WHERE name='Laptop')),
  ('LP-002', 'Laptop Pro 14" i7', 2, (SELECT id FROM categories WHERE name='Laptop')),
  ('LP-003', 'Laptop Air 13"', 1, (SELECT id FROM categories WHERE name='Laptop')),
  ('TB-001', 'Tablet Lite 10"', 1, (SELECT id FROM categories WHERE name='Tablet')),
  ('TB-002', 'Tablet Pro 11"', 2, (SELECT id FROM categories WHERE name='Tablet')),
  ('TB-003', 'Tablet Pro 13"', 2, (SELECT id FROM categories WHERE name='Tablet')),
  ('AC-001', 'USB-C Charger 30W', 1, (SELECT id FROM categories WHERE name='Accessories')),
  ('AC-002', 'USB-C Charger 65W', 1, (SELECT id FROM categories WHERE name='Accessories')),
  ('AC-003', 'Powerbank 10.000mAh', 1, (SELECT id FROM categories WHERE name='Accessories')),
  ('AC-004', 'Wireless Mouse', 2, (SELECT id FROM categories WHERE name='Accessories')),
  ('AU-001', 'Wireless Earbuds', 1, (SELECT id FROM categories WHERE name='Audio')),
  ('AU-002', 'Bluetooth Speaker', 1, (SELECT id FROM categories WHERE name='Audio')),
  ('AU-003', 'Noise Cancelling Headset', 2, (SELECT id FROM categories WHERE name='Audio')),
  ('GM-001', 'Gaming Console X', 2, (SELECT id FROM categories WHERE name='Gaming')),
  ('GM-002', 'Game Controller Pro', 2, (SELECT id FROM categories WHERE name='Gaming')),
  ('NW-001', 'WiFi Router AX1800', 1, (SELECT id FROM categories WHERE name='Networking')),
  ('NW-002', 'Mesh WiFi 2-Pack', 2, (SELECT id FROM categories WHERE name='Networking'))
ON CONFLICT DO NOTHING;

-- Inventory
INSERT INTO inventory (item_id, physical_qty, available_qty)
SELECT i.id, 50, 50 FROM items i
ON CONFLICT (item_id) DO NOTHING;
