package models

import "time"

type Item struct {
	ID         int64  `json:"id"`
	UUID       string `json:"uuid"`
	SKU        string `json:"sku"`
	Name       string `json:"name"`
	CustomerID int64  `json:"customer_id"`
	Customer   string `json:"customer"`
}

type Customer struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID       int64  `json:"id"`
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Customer string `json:"customer"`
	Category string `json:"category"`
}

type Inventory struct {
	ItemID        int64     `json:"item_id"`
	ItemUUID      string    `json:"item_uuid"`
	SKU           string    `json:"sku"`
	Name          string    `json:"name"`
	Customer      string    `json:"customer"`
	Category      string    `json:"category"`
	PhysicalQty   int64     `json:"physical_qty"`
	AvailableQty  int64     `json:"available_qty"`
	LastUpdatedAt time.Time `json:"updated_at"`
}

type StockAdjustment struct {
	SKU       string    `json:"sku"`
	Delta     int64     `json:"delta"`
	Reason    string    `json:"reason"`
	Source    string    `json:"source"`
	RefCode   string    `json:"ref_code"`
	Bucket    string    `json:"bucket"`
	CreatedAt time.Time `json:"created_at"`
}

type StockIn struct {
	ID        int64      `json:"id"`
	UUID      string     `json:"uuid"`
	Code      string     `json:"code"`
	Status    string     `json:"status"`
	Items     []LineItem `json:"items"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type StockOut struct {
	ID        int64      `json:"id"`
	UUID      string     `json:"uuid"`
	Code      string     `json:"code"`
	Status    string     `json:"status"`
	Items     []LineItem `json:"items"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type LineItem struct {
	ItemID   int64  `json:"item_id"`
	ItemUUID string `json:"item_uuid"`
	SKU      string `json:"sku"`
	Qty      int64  `json:"qty"`
}

type Report struct {
	TransactionID int64      `json:"transaction_id"`
	Code          string     `json:"code"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	Items         []LineItem `json:"items"`
	DoneAt        time.Time  `json:"done_at"`
}
