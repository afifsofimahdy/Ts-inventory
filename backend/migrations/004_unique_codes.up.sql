ALTER TABLE stock_in ADD CONSTRAINT stock_in_code_unique UNIQUE (code);
ALTER TABLE stock_out ADD CONSTRAINT stock_out_code_unique UNIQUE (code);
