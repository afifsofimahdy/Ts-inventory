package repo

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"smart-inventory-backend/internal/models"
)

type PgRepo struct {
	write *pgxpool.Pool
	read  *pgxpool.Pool
}

func NewPgRepo(write *pgxpool.Pool, read *pgxpool.Pool) *PgRepo {
	if read == nil {
		read = write
	}
	return &PgRepo{write: write, read: read}
}

func (r *PgRepo) WritePool() *pgxpool.Pool { return r.write }

func (r *PgRepo) List(ctx context.Context, filter InventoryFilter) ([]models.Inventory, error) {
	var args []interface{}
	var where []string

	if filter.Name != "" {
		args = append(args, "%"+filter.Name+"%")
		where = append(where, "LOWER(i.name) LIKE LOWER($"+itoa(len(args))+")")
	}
	if filter.SKU != "" {
		args = append(args, "%"+filter.SKU+"%")
		where = append(where, "LOWER(i.sku) LIKE LOWER($"+itoa(len(args))+")")
	}
	if filter.Customer != "" {
		args = append(args, "%"+filter.Customer+"%")
		where = append(where, "LOWER(c.name) LIKE LOWER($"+itoa(len(args))+")")
	}

	query := `
		SELECT inv.item_id, i.uuid, i.sku, i.name, c.name, COALESCE(cat.name, ''), inv.physical_qty, inv.available_qty, inv.updated_at
		FROM inventory inv
		JOIN items i ON i.id = inv.item_id
		JOIN customers c ON c.id = i.customer_id
		LEFT JOIN categories cat ON cat.id = i.category_id AND cat.deleted_at IS NULL
		WHERE inv.deleted_at IS NULL AND i.deleted_at IS NULL AND c.deleted_at IS NULL
	`
	if len(where) > 0 {
		query += " AND " + strings.Join(where, " AND ")
	}
	query += " ORDER BY i.name"

	rows, err := r.read.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.Inventory
	for rows.Next() {
		var inv models.Inventory
		if err := rows.Scan(&inv.ItemID, &inv.ItemUUID, &inv.SKU, &inv.Name, &inv.Customer, &inv.Category, &inv.PhysicalQty, &inv.AvailableQty, &inv.LastUpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, inv)
	}
	return res, rows.Err()
}

func (r *PgRepo) GetForUpdate(ctx context.Context, tx pgx.Tx, itemID int64) (models.Inventory, error) {
	var inv models.Inventory
	row := tx.QueryRow(ctx, `
		SELECT inv.item_id, i.uuid, i.sku, i.name, c.name, COALESCE(cat.name, ''), inv.physical_qty, inv.available_qty, inv.updated_at
		FROM inventory inv
		JOIN items i ON i.id = inv.item_id
		JOIN customers c ON c.id = i.customer_id
		LEFT JOIN categories cat ON cat.id = i.category_id AND cat.deleted_at IS NULL
		WHERE inv.item_id = $1 AND inv.deleted_at IS NULL AND i.deleted_at IS NULL AND c.deleted_at IS NULL
		FOR UPDATE OF inv`, itemID)
	if err := row.Scan(&inv.ItemID, &inv.ItemUUID, &inv.SKU, &inv.Name, &inv.Customer, &inv.Category, &inv.PhysicalQty, &inv.AvailableQty, &inv.LastUpdatedAt); err != nil {
		return models.Inventory{}, err
	}
	return inv, nil
}

func (r *PgRepo) GetItemIDByUUID(ctx context.Context, uuid string) (int64, error) {
	var id int64
	err := r.read.QueryRow(ctx, `SELECT id FROM items WHERE uuid=$1 AND deleted_at IS NULL`, uuid).Scan(&id)
	return id, err
}

func (r *PgRepo) GetItemIDBySKU(ctx context.Context, sku string) (int64, error) {
	var id int64
	err := r.read.QueryRow(ctx, `SELECT id FROM items WHERE sku=$1 AND deleted_at IS NULL`, sku).Scan(&id)
	return id, err
}

func (r *PgRepo) UpdateQty(ctx context.Context, tx pgx.Tx, itemID int64, physical, available int64) error {
	_, err := tx.Exec(ctx, `UPDATE inventory SET physical_qty=$2, available_qty=$3, updated_at=NOW() WHERE item_id=$1 AND deleted_at IS NULL`, itemID, physical, available)
	return err
}

func (r *PgRepo) ListAdjustments(ctx context.Context, sku string) ([]models.StockAdjustment, error) {
	rows, err := r.read.Query(ctx, `
		SELECT i.sku,
		       sa.delta,
		       COALESCE(sa.reason, ''),
		       COALESCE(sa.source, ''),
		       COALESCE(sa.ref_code, ''),
		       COALESCE(sa.bucket, ''),
		       sa.created_at
		FROM stock_adjustments sa
		JOIN items i ON i.id = sa.item_id
		WHERE i.sku=$1 AND sa.deleted_at IS NULL
		ORDER BY sa.created_at DESC`, sku)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []models.StockAdjustment
	for rows.Next() {
		var s models.StockAdjustment
		if err := rows.Scan(&s.SKU, &s.Delta, &s.Reason, &s.Source, &s.RefCode, &s.Bucket, &s.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, rows.Err()
}

func (r *PgRepo) Create(ctx context.Context, tx pgx.Tx, code string, items []models.LineItem) (int64, string, error) {
	var id int64
	var uuid string
	if err := tx.QueryRow(ctx, `INSERT INTO stock_in(code, status) VALUES ($1, 'CREATED') RETURNING id, uuid`, code).Scan(&id, &uuid); err != nil {
		return 0, "", err
	}
	for _, it := range items {
		if _, err := tx.Exec(ctx, `INSERT INTO stock_in_items(stock_in_id, item_id, qty) VALUES ($1,$2,$3)`, id, it.ItemID, it.Qty); err != nil {
			return 0, "", err
		}
	}
	return id, uuid, nil
}

func (r *PgRepo) Get(ctx context.Context, tx pgx.Tx, id int64) (models.StockIn, error) {
	var s models.StockIn
	row := tx.QueryRow(ctx, `SELECT id, uuid, code, status, created_at, updated_at FROM stock_in WHERE id=$1 AND deleted_at IS NULL`, id)
	if err := row.Scan(&s.ID, &s.UUID, &s.Code, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return models.StockIn{}, err
	}
	rows, err := tx.Query(ctx, `SELECT i.uuid, i.sku, sii.item_id, sii.qty FROM stock_in_items sii JOIN items i ON i.id = sii.item_id WHERE sii.stock_in_id=$1 AND sii.deleted_at IS NULL`, id)
	if err != nil {
		return models.StockIn{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var li models.LineItem
		if err := rows.Scan(&li.ItemUUID, &li.SKU, &li.ItemID, &li.Qty); err != nil {
			return models.StockIn{}, err
		}
		s.Items = append(s.Items, li)
	}
	return s, rows.Err()
}

func (r *PgRepo) GetByUUID(ctx context.Context, tx pgx.Tx, uuid string) (models.StockIn, error) {
	var s models.StockIn
	row := tx.QueryRow(ctx, `SELECT id, uuid, code, status, created_at, updated_at FROM stock_in WHERE uuid=$1 AND deleted_at IS NULL`, uuid)
	if err := row.Scan(&s.ID, &s.UUID, &s.Code, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return models.StockIn{}, err
	}
	rows, err := tx.Query(ctx, `SELECT i.uuid, i.sku, sii.item_id, sii.qty FROM stock_in_items sii JOIN items i ON i.id = sii.item_id WHERE sii.stock_in_id=$1 AND sii.deleted_at IS NULL`, s.ID)
	if err != nil {
		return models.StockIn{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var li models.LineItem
		if err := rows.Scan(&li.ItemUUID, &li.SKU, &li.ItemID, &li.Qty); err != nil {
			return models.StockIn{}, err
		}
		s.Items = append(s.Items, li)
	}
	return s, rows.Err()
}

func (r *PgRepo) GetByCode(ctx context.Context, tx pgx.Tx, code string) (models.StockIn, error) {
	var s models.StockIn
	row := tx.QueryRow(ctx, `SELECT id, uuid, code, status, created_at, updated_at FROM stock_in WHERE code=$1 AND deleted_at IS NULL`, code)
	if err := row.Scan(&s.ID, &s.UUID, &s.Code, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return models.StockIn{}, err
	}
	rows, err := tx.Query(ctx, `SELECT i.uuid, i.sku, sii.item_id, sii.qty FROM stock_in_items sii JOIN items i ON i.id = sii.item_id WHERE sii.stock_in_id=$1 AND sii.deleted_at IS NULL`, s.ID)
	if err != nil {
		return models.StockIn{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var li models.LineItem
		if err := rows.Scan(&li.ItemUUID, &li.SKU, &li.ItemID, &li.Qty); err != nil {
			return models.StockIn{}, err
		}
		s.Items = append(s.Items, li)
	}
	return s, rows.Err()
}

func (r *PgRepo) ListStockInByStatus(ctx context.Context, statuses []string) ([]models.StockIn, error) {
	if len(statuses) == 0 {
		return []models.StockIn{}, nil
	}
	query, args := buildStatusQuery(`
		SELECT id, uuid, code, status, created_at, updated_at
		FROM stock_in
		WHERE deleted_at IS NULL AND status IN ({{placeholders}})
		ORDER BY created_at DESC`, statuses)
	rows, err := r.read.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.StockIn
	for rows.Next() {
		var s models.StockIn
		if err := rows.Scan(&s.ID, &s.UUID, &s.Code, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		items, err := r.listStockInItems(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		s.Items = items
		res = append(res, s)
	}
	return res, rows.Err()
}

func (r *PgRepo) UpdateStatus(ctx context.Context, tx pgx.Tx, id int64, status string) error {
	_, err := tx.Exec(ctx, `UPDATE stock_in SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	return err
}

func (r *PgRepo) AddLog(ctx context.Context, tx pgx.Tx, id int64, fromStatus, toStatus string) error {
	_, err := tx.Exec(ctx, `INSERT INTO stock_in_logs(stock_in_id, from_status, to_status) VALUES ($1,$2,$3)`, id, fromStatus, toStatus)
	return err
}

func (r *PgRepo) DeleteStockIn(ctx context.Context, tx pgx.Tx, id int64) error {
	if _, err := tx.Exec(ctx, `UPDATE stock_in_items SET deleted_at=NOW() WHERE stock_in_id=$1 AND deleted_at IS NULL`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE stock_in_logs SET deleted_at=NOW() WHERE stock_in_id=$1 AND deleted_at IS NULL`, id); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `UPDATE stock_in SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *PgRepo) listStockInItems(ctx context.Context, id int64) ([]models.LineItem, error) {
	rows, err := r.read.Query(ctx, `
		SELECT i.uuid, i.sku, sii.item_id, sii.qty
		FROM stock_in_items sii
		JOIN items i ON i.id = sii.item_id
		WHERE sii.stock_in_id=$1 AND sii.deleted_at IS NULL`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.LineItem
	for rows.Next() {
		var li models.LineItem
		if err := rows.Scan(&li.ItemUUID, &li.SKU, &li.ItemID, &li.Qty); err != nil {
			return nil, err
		}
		items = append(items, li)
	}
	return items, rows.Err()
}

func (r *PgRepo) CreateOut(ctx context.Context, tx pgx.Tx, code string, items []models.LineItem) (int64, string, error) {
	var id int64
	var uuid string
	if err := tx.QueryRow(ctx, `INSERT INTO stock_out(code, status) VALUES ($1, 'DRAFT') RETURNING id, uuid`, code).Scan(&id, &uuid); err != nil {
		return 0, "", err
	}
	for _, it := range items {
		if _, err := tx.Exec(ctx, `INSERT INTO stock_out_items(stock_out_id, item_id, qty) VALUES ($1,$2,$3)`, id, it.ItemID, it.Qty); err != nil {
			return 0, "", err
		}
	}
	return id, uuid, nil
}

func (r *PgRepo) GetOut(ctx context.Context, tx pgx.Tx, id int64) (models.StockOut, error) {
	var s models.StockOut
	row := tx.QueryRow(ctx, `SELECT id, uuid, code, status, created_at, updated_at FROM stock_out WHERE id=$1 AND deleted_at IS NULL`, id)
	if err := row.Scan(&s.ID, &s.UUID, &s.Code, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return models.StockOut{}, err
	}
	rows, err := tx.Query(ctx, `SELECT i.uuid, i.sku, soi.item_id, soi.qty FROM stock_out_items soi JOIN items i ON i.id = soi.item_id WHERE soi.stock_out_id=$1 AND soi.deleted_at IS NULL`, id)
	if err != nil {
		return models.StockOut{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var li models.LineItem
		if err := rows.Scan(&li.ItemUUID, &li.SKU, &li.ItemID, &li.Qty); err != nil {
			return models.StockOut{}, err
		}
		s.Items = append(s.Items, li)
	}
	return s, rows.Err()
}

func (r *PgRepo) GetOutByUUID(ctx context.Context, tx pgx.Tx, uuid string) (models.StockOut, error) {
	var s models.StockOut
	row := tx.QueryRow(ctx, `SELECT id, uuid, code, status, created_at, updated_at FROM stock_out WHERE uuid=$1 AND deleted_at IS NULL`, uuid)
	if err := row.Scan(&s.ID, &s.UUID, &s.Code, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return models.StockOut{}, err
	}
	rows, err := tx.Query(ctx, `SELECT i.uuid, i.sku, soi.item_id, soi.qty FROM stock_out_items soi JOIN items i ON i.id = soi.item_id WHERE soi.stock_out_id=$1 AND soi.deleted_at IS NULL`, s.ID)
	if err != nil {
		return models.StockOut{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var li models.LineItem
		if err := rows.Scan(&li.ItemUUID, &li.SKU, &li.ItemID, &li.Qty); err != nil {
			return models.StockOut{}, err
		}
		s.Items = append(s.Items, li)
	}
	return s, rows.Err()
}

func (r *PgRepo) GetOutByCode(ctx context.Context, tx pgx.Tx, code string) (models.StockOut, error) {
	var s models.StockOut
	row := tx.QueryRow(ctx, `SELECT id, uuid, code, status, created_at, updated_at FROM stock_out WHERE code=$1 AND deleted_at IS NULL`, code)
	if err := row.Scan(&s.ID, &s.UUID, &s.Code, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return models.StockOut{}, err
	}
	rows, err := tx.Query(ctx, `SELECT i.uuid, i.sku, soi.item_id, soi.qty FROM stock_out_items soi JOIN items i ON i.id = soi.item_id WHERE soi.stock_out_id=$1 AND soi.deleted_at IS NULL`, s.ID)
	if err != nil {
		return models.StockOut{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var li models.LineItem
		if err := rows.Scan(&li.ItemUUID, &li.SKU, &li.ItemID, &li.Qty); err != nil {
			return models.StockOut{}, err
		}
		s.Items = append(s.Items, li)
	}
	return s, rows.Err()
}

func (r *PgRepo) ListStockOutByStatus(ctx context.Context, statuses []string) ([]models.StockOut, error) {
	if len(statuses) == 0 {
		return []models.StockOut{}, nil
	}
	query, args := buildStatusQuery(`
		SELECT id, uuid, code, status, created_at, updated_at
		FROM stock_out
		WHERE deleted_at IS NULL AND status IN ({{placeholders}})
		ORDER BY created_at DESC`, statuses)
	rows, err := r.read.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.StockOut
	for rows.Next() {
		var s models.StockOut
		if err := rows.Scan(&s.ID, &s.UUID, &s.Code, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		items, err := r.listStockOutItems(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		s.Items = items
		res = append(res, s)
	}
	return res, rows.Err()
}

func (r *PgRepo) UpdateOutStatus(ctx context.Context, tx pgx.Tx, id int64, status string) error {
	_, err := tx.Exec(ctx, `UPDATE stock_out SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	return err
}

func (r *PgRepo) AddOutLog(ctx context.Context, tx pgx.Tx, id int64, fromStatus, toStatus string) error {
	_, err := tx.Exec(ctx, `INSERT INTO stock_out_logs(stock_out_id, from_status, to_status) VALUES ($1,$2,$3)`, id, fromStatus, toStatus)
	return err
}

func (r *PgRepo) DeleteStockOut(ctx context.Context, tx pgx.Tx, id int64) error {
	if _, err := tx.Exec(ctx, `UPDATE stock_out_items SET deleted_at=NOW() WHERE stock_out_id=$1 AND deleted_at IS NULL`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE stock_out_logs SET deleted_at=NOW() WHERE stock_out_id=$1 AND deleted_at IS NULL`, id); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `UPDATE stock_out SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *PgRepo) listStockOutItems(ctx context.Context, id int64) ([]models.LineItem, error) {
	rows, err := r.read.Query(ctx, `
		SELECT i.uuid, i.sku, soi.item_id, soi.qty
		FROM stock_out_items soi
		JOIN items i ON i.id = soi.item_id
		WHERE soi.stock_out_id=$1 AND soi.deleted_at IS NULL`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.LineItem
	for rows.Next() {
		var li models.LineItem
		if err := rows.Scan(&li.ItemUUID, &li.SKU, &li.ItemID, &li.Qty); err != nil {
			return nil, err
		}
		items = append(items, li)
	}
	return items, rows.Err()
}

func buildStatusQuery(base string, statuses []string) (string, []interface{}) {
	args := make([]interface{}, 0, len(statuses))
	ph := make([]string, 0, len(statuses))
	for i, s := range statuses {
		args = append(args, s)
		ph = append(ph, "$"+itoa(i+1))
	}
	query := strings.Replace(base, "{{placeholders}}", strings.Join(ph, ","), 1)
	return query, args
}

func (r *PgRepo) ListCustomers(ctx context.Context) ([]models.Customer, error) {
	rows, err := r.read.Query(ctx, `SELECT id, name FROM customers WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []models.Customer
	for rows.Next() {
		var c models.Customer
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	return res, rows.Err()
}

func (r *PgRepo) ListStockInDoneRange(ctx context.Context, from *time.Time, to *time.Time) ([]models.StockIn, error) {
	query := `SELECT id, uuid, code, status, created_at, updated_at
		FROM stock_in
		WHERE status='DONE' AND deleted_at IS NULL`
	args := []interface{}{}
	if from != nil {
		args = append(args, *from)
		query += " AND created_at >= $" + itoa(len(args))
	}
	if to != nil {
		args = append(args, *to)
		query += " AND created_at < $" + itoa(len(args))
	}
	query += " ORDER BY updated_at DESC"

	rows, err := r.read.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []models.StockIn
	for rows.Next() {
		var s models.StockIn
		if err := rows.Scan(&s.ID, &s.UUID, &s.Code, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		itemRows, err := r.read.Query(ctx, `SELECT i.uuid, i.sku, sii.item_id, sii.qty FROM stock_in_items sii JOIN items i ON i.id = sii.item_id WHERE sii.stock_in_id=$1 AND sii.deleted_at IS NULL`, s.ID)
		if err != nil {
			return nil, err
		}
		for itemRows.Next() {
			var li models.LineItem
			if err := itemRows.Scan(&li.ItemUUID, &li.SKU, &li.ItemID, &li.Qty); err != nil {
				itemRows.Close()
				return nil, err
			}
			s.Items = append(s.Items, li)
		}
		itemRows.Close()
		res = append(res, s)
	}
	return res, rows.Err()
}

func (r *PgRepo) ListStockOutDoneRange(ctx context.Context, from *time.Time, to *time.Time) ([]models.StockOut, error) {
	query := `SELECT id, uuid, code, status, created_at, updated_at
		FROM stock_out
		WHERE status='DONE' AND deleted_at IS NULL`
	args := []interface{}{}
	if from != nil {
		args = append(args, *from)
		query += " AND created_at >= $" + itoa(len(args))
	}
	if to != nil {
		args = append(args, *to)
		query += " AND created_at < $" + itoa(len(args))
	}
	query += " ORDER BY updated_at DESC"

	rows, err := r.read.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []models.StockOut
	for rows.Next() {
		var s models.StockOut
		if err := rows.Scan(&s.ID, &s.UUID, &s.Code, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		itemRows, err := r.read.Query(ctx, `SELECT i.uuid, i.sku, soi.item_id, soi.qty FROM stock_out_items soi JOIN items i ON i.id = soi.item_id WHERE soi.stock_out_id=$1 AND soi.deleted_at IS NULL`, s.ID)
		if err != nil {
			return nil, err
		}
		for itemRows.Next() {
			var li models.LineItem
			if err := itemRows.Scan(&li.ItemUUID, &li.SKU, &li.ItemID, &li.Qty); err != nil {
				itemRows.Close()
				return nil, err
			}
			s.Items = append(s.Items, li)
		}
		itemRows.Close()
		res = append(res, s)
	}
	return res, rows.Err()
}

func (r *PgRepo) CreateCustomer(ctx context.Context, name string) (int64, error) {
	var id int64
	if err := r.write.QueryRow(ctx, `INSERT INTO customers(name) VALUES ($1) RETURNING id`, name).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *PgRepo) ListCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := r.read.Query(ctx, `SELECT id, name FROM categories WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	return res, rows.Err()
}

func (r *PgRepo) CreateCategory(ctx context.Context, name string) (int64, error) {
	var id int64
	if err := r.write.QueryRow(ctx, `INSERT INTO categories(name) VALUES ($1) RETURNING id`, name).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *PgRepo) ListProducts(ctx context.Context) ([]models.Product, error) {
	rows, err := r.read.Query(ctx, `
		SELECT i.id, i.sku, i.name, c.name, COALESCE(cat.name,'')
		FROM items i
		JOIN customers c ON c.id = i.customer_id
		LEFT JOIN categories cat ON cat.id = i.category_id AND cat.deleted_at IS NULL
		WHERE i.deleted_at IS NULL AND c.deleted_at IS NULL
		ORDER BY i.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Customer, &p.Category); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, rows.Err()
}

func (r *PgRepo) CreateProduct(ctx context.Context, sku, name string, customerID, categoryID int64) (int64, error) {
	var id int64
	if err := r.write.QueryRow(ctx, `INSERT INTO items(sku, name, customer_id, category_id) VALUES ($1,$2,$3,$4) RETURNING id`, sku, name, customerID, categoryID).Scan(&id); err != nil {
		return 0, err
	}
	_, err := r.write.Exec(ctx, `INSERT INTO inventory(item_id, physical_qty, available_qty) VALUES ($1,0,0) ON CONFLICT (item_id) DO NOTHING`, id)
	return id, err
}

func itoa(i int) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = digits[i%10]
		i /= 10
	}
	return string(b[pos:])
}
