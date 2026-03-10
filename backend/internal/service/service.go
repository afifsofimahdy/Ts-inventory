package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"smart-inventory-backend/internal/models"
	"smart-inventory-backend/internal/repo"
)

var (
	ErrInvalidStatus = errors.New("invalid status transition")
	ErrInsufficient  = errors.New("insufficient stock")
	ErrNotAllowed    = errors.New("operation not allowed")
)

const (
	sourceManual           = "MANUAL"
	sourceStockInDone      = "STOCK_IN_DONE"
	sourceStockOutAllocated = "STOCK_OUT_ALLOCATED"
	sourceStockOutDone     = "STOCK_OUT_DONE"
	sourceStockOutCancelled = "STOCK_OUT_CANCELLED"

	bucketPhysical  = "PHYSICAL"
	bucketAvailable = "AVAILABLE"
)

type Service struct {
	invRepo repo.InventoryRepo
	inRepo  repo.StockInRepo
	outRepo repo.StockOutRepo
	repRepo repo.ReportRepo
	master  repo.MasterRepo
	pool    *pgxpool.Pool
}

func New(pool *pgxpool.Pool, inv repo.InventoryRepo, in repo.StockInRepo, out repo.StockOutRepo, rep repo.ReportRepo, master repo.MasterRepo) *Service {
	return &Service{pool: pool, invRepo: inv, inRepo: in, outRepo: out, repRepo: rep, master: master}
}

func (s *Service) ListInventory(ctx context.Context, filter repo.InventoryFilter) ([]models.Inventory, error) {
	return s.invRepo.List(ctx, filter)
}

func (s *Service) ListAdjustments(ctx context.Context, sku string) ([]models.StockAdjustment, error) {
	return s.invRepo.ListAdjustments(ctx, sku)
}

func (s *Service) AdjustStock(ctx context.Context, sku string, newPhysical int64, reason string) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		itemID, err := s.invRepo.GetItemIDBySKU(ctx, sku)
		if err != nil {
			return err
		}
		inv, err := s.invRepo.GetForUpdate(ctx, tx, itemID)
		if err != nil {
			return err
		}
		reserved := inv.PhysicalQty - inv.AvailableQty
		newAvailable := newPhysical - reserved
		if newAvailable < 0 {
			return ErrNotAllowed
		}
		if err := s.invRepo.UpdateQty(ctx, tx, itemID, newPhysical, newAvailable); err != nil {
			return err
		}
		deltaPhysical := newPhysical - inv.PhysicalQty
		deltaAvailable := newAvailable - inv.AvailableQty
		if deltaPhysical != 0 {
			_, err = tx.Exec(ctx, `INSERT INTO stock_adjustments(item_id, delta, reason, source, bucket) VALUES ($1,$2,$3,$4,$5)`,
				itemID, deltaPhysical, reason, sourceManual, bucketPhysical)
			if err != nil {
				return err
			}
		}
		if deltaAvailable != 0 {
			_, err = tx.Exec(ctx, `INSERT INTO stock_adjustments(item_id, delta, reason, source, bucket) VALUES ($1,$2,$3,$4,$5)`,
				itemID, deltaAvailable, reason, sourceManual, bucketAvailable)
			if err != nil {
				return err
			}
		}
		if err == nil {
			log.Printf("adjust_stock sku=%s delta_physical=%d delta_available=%d reason=%s", sku, deltaPhysical, deltaAvailable, reason)
		}
		return err
	})
}

func (s *Service) CreateStockIn(ctx context.Context, code string, items []models.LineItem) (string, string, error) {
	var uuid string
	var finalCode string
	return finalCode, uuid, s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		for i := range items {
			if items[i].ItemID == 0 {
				itemID, err := s.invRepo.GetItemIDBySKU(ctx, items[i].SKU)
				if err != nil {
					return err
				}
				items[i].ItemID = itemID
			}
		}
		finalCode = code
		if finalCode == "" {
			finalCode = generateDocCode("SIN")
		}
		var id int64
		for i := 0; i < 5; i++ {
			id, uuid, err = s.inRepo.Create(ctx, tx, finalCode, items)
			if err == nil {
				break
			}
			if isUniqueViolation(err) {
				finalCode = generateDocCode("SIN")
				continue
			}
			return err
		}
		if err != nil {
			return err
		}
		log.Printf("stock_in_created code=%s uuid=%s items=%d", finalCode, uuid, len(items))
		return s.inRepo.AddLog(ctx, tx, id, "", "CREATED")
	})
}

func (s *Service) UpdateStockInStatus(ctx context.Context, code string, to string) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		stock, err := s.inRepo.GetByCode(ctx, tx, code)
		if err != nil {
			return err
		}
		from := stock.Status
		if from == "DONE" && to == "CANCELLED" {
			return ErrNotAllowed
		}
		if !validStockInTransition(from, to) {
			return ErrInvalidStatus
		}
		if to == "DONE" {
			for _, it := range stock.Items {
				inv, err := s.invRepo.GetForUpdate(ctx, tx, it.ItemID)
				if err != nil {
					return err
				}
				newPhysical := inv.PhysicalQty + it.Qty
				newAvailable := inv.AvailableQty + it.Qty
				if err := s.invRepo.UpdateQty(ctx, tx, it.ItemID, newPhysical, newAvailable); err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, `INSERT INTO stock_adjustments(item_id, delta, reason, source, ref_code, bucket) VALUES ($1,$2,$3,$4,$5,$6)`,
					it.ItemID, it.Qty, "Barang masuk", sourceStockInDone, code, bucketPhysical); err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, `INSERT INTO stock_adjustments(item_id, delta, reason, source, ref_code, bucket) VALUES ($1,$2,$3,$4,$5,$6)`,
					it.ItemID, it.Qty, "Barang masuk", sourceStockInDone, code, bucketAvailable); err != nil {
					return err
				}
			}
		}
		if err := s.inRepo.UpdateStatus(ctx, tx, stock.ID, to); err != nil {
			return err
		}
		log.Printf("stock_in_status code=%s from=%s to=%s", code, from, to)
		return s.inRepo.AddLog(ctx, tx, stock.ID, from, to)
	})
}

func (s *Service) CreateStockOut(ctx context.Context, code string, items []models.LineItem) (string, string, error) {
	var uuid string
	var finalCode string
	return finalCode, uuid, s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		for i := range items {
			if items[i].ItemID == 0 {
				itemID, err := s.invRepo.GetItemIDBySKU(ctx, items[i].SKU)
				if err != nil {
					return err
				}
				items[i].ItemID = itemID
			}
		}
		finalCode = code
		if finalCode == "" {
			finalCode = generateDocCode("SOUT")
		}
		var id int64
		for i := 0; i < 5; i++ {
			id, uuid, err = s.outRepo.CreateOut(ctx, tx, finalCode, items)
			if err == nil {
				break
			}
			if isUniqueViolation(err) {
				finalCode = generateDocCode("SOUT")
				continue
			}
			return err
		}
		if err != nil {
			return err
		}
		log.Printf("stock_out_created code=%s uuid=%s items=%d", finalCode, uuid, len(items))
		return s.outRepo.AddOutLog(ctx, tx, id, "", "DRAFT")
	})
}

func (s *Service) AllocateStockOut(ctx context.Context, code string) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		stock, err := s.outRepo.GetOutByCode(ctx, tx, code)
		if err != nil {
			return err
		}
		if stock.Status != "DRAFT" {
			return ErrInvalidStatus
		}
		for _, it := range stock.Items {
			inv, err := s.invRepo.GetForUpdate(ctx, tx, it.ItemID)
			if err != nil {
				return err
			}
			if inv.AvailableQty < it.Qty {
				return ErrInsufficient
			}
			newAvailable := inv.AvailableQty - it.Qty
			if err := s.invRepo.UpdateQty(ctx, tx, it.ItemID, inv.PhysicalQty, newAvailable); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `INSERT INTO stock_adjustments(item_id, delta, reason, source, ref_code, bucket) VALUES ($1,$2,$3,$4,$5,$6)`,
				it.ItemID, -it.Qty, "Booking stok", sourceStockOutAllocated, code, bucketAvailable); err != nil {
				return err
			}
		}
		if err := s.outRepo.UpdateOutStatus(ctx, tx, stock.ID, "ALLOCATED"); err != nil {
			return err
		}
		log.Printf("stock_out_allocated code=%s", code)
		return s.outRepo.AddOutLog(ctx, tx, stock.ID, stock.Status, "ALLOCATED")
	})
}

func (s *Service) UpdateStockOutStatus(ctx context.Context, code string, to string) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		stock, err := s.outRepo.GetOutByCode(ctx, tx, code)
		if err != nil {
			return err
		}
		from := stock.Status
		if !validStockOutTransition(from, to) {
			return ErrInvalidStatus
		}
		if to == "CANCELLED" && (from == "IN_PROGRESS" || from == "ALLOCATED") {
			for _, it := range stock.Items {
				inv, err := s.invRepo.GetForUpdate(ctx, tx, it.ItemID)
				if err != nil {
					return err
				}
				newAvailable := inv.AvailableQty + it.Qty
				if err := s.invRepo.UpdateQty(ctx, tx, it.ItemID, inv.PhysicalQty, newAvailable); err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, `INSERT INTO stock_adjustments(item_id, delta, reason, source, ref_code, bucket) VALUES ($1,$2,$3,$4,$5,$6)`,
					it.ItemID, it.Qty, "Batal booking", sourceStockOutCancelled, code, bucketAvailable); err != nil {
					return err
				}
			}
		}
		if to == "DONE" {
			for _, it := range stock.Items {
				inv, err := s.invRepo.GetForUpdate(ctx, tx, it.ItemID)
				if err != nil {
					return err
				}
				newPhysical := inv.PhysicalQty - it.Qty
				if newPhysical < 0 {
					return ErrInsufficient
				}
				if err := s.invRepo.UpdateQty(ctx, tx, it.ItemID, newPhysical, inv.AvailableQty); err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, `INSERT INTO stock_adjustments(item_id, delta, reason, source, ref_code, bucket) VALUES ($1,$2,$3,$4,$5,$6)`,
					it.ItemID, -it.Qty, "Barang keluar", sourceStockOutDone, code, bucketPhysical); err != nil {
					return err
				}
			}
		}
		if err := s.outRepo.UpdateOutStatus(ctx, tx, stock.ID, to); err != nil {
			return err
		}
		log.Printf("stock_out_status code=%s from=%s to=%s", code, from, to)
		return s.outRepo.AddOutLog(ctx, tx, stock.ID, from, to)
	})
}

func (s *Service) ListStockInByStatus(ctx context.Context, statuses []string) ([]models.StockIn, error) {
	return s.inRepo.ListStockInByStatus(ctx, statuses)
}

func (s *Service) ListStockOutByStatus(ctx context.Context, statuses []string) ([]models.StockOut, error) {
	return s.outRepo.ListStockOutByStatus(ctx, statuses)
}

func (s *Service) ReportsStockIn(ctx context.Context, from *time.Time, to *time.Time) ([]models.StockIn, error) {
	return s.repRepo.ListStockInDoneRange(ctx, from, to)
}

func (s *Service) ReportsStockOut(ctx context.Context, from *time.Time, to *time.Time) ([]models.StockOut, error) {
	return s.repRepo.ListStockOutDoneRange(ctx, from, to)
}

func (s *Service) DeleteStockIn(ctx context.Context, code string) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		stock, err := s.inRepo.GetByCode(ctx, tx, code)
		if err != nil {
			return err
		}
		if stock.Status != "CREATED" {
			return ErrNotAllowed
		}
		return s.inRepo.DeleteStockIn(ctx, tx, stock.ID)
	})
}

func (s *Service) DeleteStockOut(ctx context.Context, code string) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		stock, err := s.outRepo.GetOutByCode(ctx, tx, code)
		if err != nil {
			return err
		}
		if stock.Status != "DRAFT" {
			return ErrNotAllowed
		}
		return s.outRepo.DeleteStockOut(ctx, tx, stock.ID)
	})
}

func (s *Service) ListCustomers(ctx context.Context) ([]models.Customer, error) {
	return s.master.ListCustomers(ctx)
}

func (s *Service) CreateCustomer(ctx context.Context, name string) (int64, error) {
	return s.master.CreateCustomer(ctx, name)
}

func (s *Service) ListCategories(ctx context.Context) ([]models.Category, error) {
	return s.master.ListCategories(ctx)
}

func (s *Service) CreateCategory(ctx context.Context, name string) (int64, error) {
	return s.master.CreateCategory(ctx, name)
}

func (s *Service) ListProducts(ctx context.Context) ([]models.Product, error) {
	return s.master.ListProducts(ctx)
}

func (s *Service) CreateProduct(ctx context.Context, sku, name string, customerID, categoryID int64) (int64, error) {
	return s.master.CreateProduct(ctx, sku, name, customerID, categoryID)
}

func (s *Service) GetStockInByCode(ctx context.Context, code string) (models.StockIn, error) {
	var res models.StockIn
	err := s.withTx(ctx, func(tx pgx.Tx) error {
		txRes, err := s.inRepo.GetByCode(ctx, tx, code)
		if err != nil {
			return err
		}
		res = txRes
		return nil
	})
	return res, err
}

func (s *Service) GetStockOutByCode(ctx context.Context, code string) (models.StockOut, error) {
	var res models.StockOut
	err := s.withTx(ctx, func(tx pgx.Tx) error {
		txRes, err := s.outRepo.GetOutByCode(ctx, tx, code)
		if err != nil {
			return err
		}
		res = txRes
		return nil
	})
	return res, err
}

func validStockInTransition(from, to string) bool {
	allowed := map[string][]string{
		"CREATED":     {"IN_PROGRESS", "CANCELLED"},
		"IN_PROGRESS": {"DONE", "CANCELLED"},
		"DONE":        {},
		"CANCELLED":   {},
	}
	for _, v := range allowed[from] {
		if v == to {
			return true
		}
	}
	return false
}

func validStockOutTransition(from, to string) bool {
	allowed := map[string][]string{
		"DRAFT":       {"ALLOCATED", "CANCELLED"},
		"ALLOCATED":   {"IN_PROGRESS", "CANCELLED"},
		"IN_PROGRESS": {"DONE", "CANCELLED"},
		"DONE":        {},
		"CANCELLED":   {},
	}
	for _, v := range allowed[from] {
		if v == to {
			return true
		}
	}
	return false
}

func (s *Service) withTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func generateDocCode(prefix string) string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s-%s-%04d", prefix, time.Now().Format("20060102"), rand.Intn(10000))
}
