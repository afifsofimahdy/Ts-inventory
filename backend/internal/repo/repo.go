package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"smart-inventory-backend/internal/models"
)

type InventoryRepo interface {
	List(ctx context.Context, filter InventoryFilter) ([]models.Inventory, error)
	GetForUpdate(ctx context.Context, tx pgx.Tx, itemID int64) (models.Inventory, error)
	GetItemIDByUUID(ctx context.Context, uuid string) (int64, error)
	GetItemIDBySKU(ctx context.Context, sku string) (int64, error)
	UpdateQty(ctx context.Context, tx pgx.Tx, itemID int64, physical, available int64) error
	ListAdjustments(ctx context.Context, sku string) ([]models.StockAdjustment, error)
}

type MasterRepo interface {
	ListCustomers(ctx context.Context) ([]models.Customer, error)
	CreateCustomer(ctx context.Context, name string) (int64, error)
	ListCategories(ctx context.Context) ([]models.Category, error)
	CreateCategory(ctx context.Context, name string) (int64, error)
	ListProducts(ctx context.Context) ([]models.Product, error)
	CreateProduct(ctx context.Context, sku, name string, customerID, categoryID int64) (int64, error)
}

type StockInRepo interface {
	Create(ctx context.Context, tx pgx.Tx, code string, items []models.LineItem) (int64, string, error)
	Get(ctx context.Context, tx pgx.Tx, id int64) (models.StockIn, error)
	GetByUUID(ctx context.Context, tx pgx.Tx, uuid string) (models.StockIn, error)
	GetByCode(ctx context.Context, tx pgx.Tx, code string) (models.StockIn, error)
	ListStockInByStatus(ctx context.Context, statuses []string) ([]models.StockIn, error)
	DeleteStockIn(ctx context.Context, tx pgx.Tx, id int64) error
	UpdateStatus(ctx context.Context, tx pgx.Tx, id int64, status string) error
	AddLog(ctx context.Context, tx pgx.Tx, id int64, fromStatus, toStatus string) error
}

type StockOutRepo interface {
	CreateOut(ctx context.Context, tx pgx.Tx, code string, items []models.LineItem) (int64, string, error)
	GetOut(ctx context.Context, tx pgx.Tx, id int64) (models.StockOut, error)
	GetOutByUUID(ctx context.Context, tx pgx.Tx, uuid string) (models.StockOut, error)
	GetOutByCode(ctx context.Context, tx pgx.Tx, code string) (models.StockOut, error)
	ListStockOutByStatus(ctx context.Context, statuses []string) ([]models.StockOut, error)
	DeleteStockOut(ctx context.Context, tx pgx.Tx, id int64) error
	UpdateOutStatus(ctx context.Context, tx pgx.Tx, id int64, status string) error
	AddOutLog(ctx context.Context, tx pgx.Tx, id int64, fromStatus, toStatus string) error
}

type ReportRepo interface {
	ListStockInDoneRange(ctx context.Context, from *time.Time, to *time.Time) ([]models.StockIn, error)
	ListStockOutDoneRange(ctx context.Context, from *time.Time, to *time.Time) ([]models.StockOut, error)
}

type InventoryFilter struct {
	Name     string
	SKU      string
	Customer string
}
