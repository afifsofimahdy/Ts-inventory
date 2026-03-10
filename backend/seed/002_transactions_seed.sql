-- Sample Stock In (DONE)
INSERT INTO stock_in (code, status) VALUES ('SIN-0001', 'DONE') ON CONFLICT DO NOTHING;
INSERT INTO stock_in_items (stock_in_id, item_id, qty)
SELECT si.id, i.id, 5
FROM stock_in si
JOIN items i ON i.sku IN ('SM-001','LP-001')
WHERE si.code='SIN-0001'
ON CONFLICT DO NOTHING;
INSERT INTO stock_in_logs (stock_in_id, from_status, to_status)
SELECT id, NULL, 'CREATED' FROM stock_in WHERE code='SIN-0001' ON CONFLICT DO NOTHING;
INSERT INTO stock_in_logs (stock_in_id, from_status, to_status)
SELECT id, 'CREATED', 'IN_PROGRESS' FROM stock_in WHERE code='SIN-0001' ON CONFLICT DO NOTHING;
INSERT INTO stock_in_logs (stock_in_id, from_status, to_status)
SELECT id, 'IN_PROGRESS', 'DONE' FROM stock_in WHERE code='SIN-0001' ON CONFLICT DO NOTHING;

-- Apply stock-in effect (increase physical & available)
UPDATE inventory inv
SET physical_qty = inv.physical_qty + sii.qty,
    available_qty = inv.available_qty + sii.qty,
    updated_at = NOW()
FROM stock_in_items sii
JOIN stock_in si ON si.id = sii.stock_in_id
WHERE si.code='SIN-0001' AND inv.item_id = sii.item_id;

-- Sample Stock Out (DONE)
INSERT INTO stock_out (code, status) VALUES ('SOUT-0001', 'DONE') ON CONFLICT DO NOTHING;
INSERT INTO stock_out_items (stock_out_id, item_id, qty)
SELECT so.id, i.id, 3
FROM stock_out so
JOIN items i ON i.sku IN ('SM-001','AC-001')
WHERE so.code='SOUT-0001'
ON CONFLICT DO NOTHING;
INSERT INTO stock_out_logs (stock_out_id, from_status, to_status)
SELECT id, NULL, 'DRAFT' FROM stock_out WHERE code='SOUT-0001' ON CONFLICT DO NOTHING;
INSERT INTO stock_out_logs (stock_out_id, from_status, to_status)
SELECT id, 'DRAFT', 'ALLOCATED' FROM stock_out WHERE code='SOUT-0001' ON CONFLICT DO NOTHING;
INSERT INTO stock_out_logs (stock_out_id, from_status, to_status)
SELECT id, 'ALLOCATED', 'IN_PROGRESS' FROM stock_out WHERE code='SOUT-0001' ON CONFLICT DO NOTHING;
INSERT INTO stock_out_logs (stock_out_id, from_status, to_status)
SELECT id, 'IN_PROGRESS', 'DONE' FROM stock_out WHERE code='SOUT-0001' ON CONFLICT DO NOTHING;

-- Apply stock-out effect (reduce physical; available assumed already reserved)
UPDATE inventory inv
SET physical_qty = inv.physical_qty - soi.qty,
    available_qty = inv.available_qty - soi.qty,
    updated_at = NOW()
FROM stock_out_items soi
JOIN stock_out so ON so.id = soi.stock_out_id
WHERE so.code='SOUT-0001' AND inv.item_id = soi.item_id;
