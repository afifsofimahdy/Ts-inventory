package handlers

import "time"

type InventoryResponse struct {
	SKU          string    `json:"sku"`
	Name         string    `json:"name"`
	Customer     string    `json:"customer"`
	Category     string    `json:"category"`
	PhysicalQty  int64     `json:"physical_qty"`
	AvailableQty int64     `json:"available_qty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type StockAdjustmentDTO struct {
	SKU       string    `json:"sku"`
	Delta     int64     `json:"delta"`
	Reason    string    `json:"reason"`
	Source    string    `json:"source"`
	RefCode   string    `json:"ref_code"`
	Bucket    string    `json:"bucket"`
	CreatedAt time.Time `json:"created_at"`
}

type AdjustInventoryRequest struct {
	SKU         string `json:"sku" validate:"required"`
	NewPhysical int64  `json:"new_physical" validate:"required,gte=0"`
	Reason      string `json:"reason" validate:"required,min=3,max=255"`
}

type StockLineDTO struct {
	SKU string `json:"sku" validate:"required"`
	Qty int64  `json:"qty" validate:"required,gt=0"`
}

type CreateStockInRequest struct {
	Code  string         `json:"code" validate:"omitempty,min=3,max=50"`
	Items []StockLineDTO `json:"items" validate:"required,min=1,dive"`
}

type CreateStockOutRequest struct {
	Code  string         `json:"code" validate:"omitempty,min=3,max=50"`
	Items []StockLineDTO `json:"items" validate:"required,min=1,dive"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=CREATED IN_PROGRESS DONE CANCELLED DRAFT ALLOCATED"`
}

type TransactionResponse struct {
	UUID      string         `json:"uuid"`
	Code      string         `json:"code"`
	Status    string         `json:"status"`
	Items     []StockLineDTO `json:"items"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type CustomerDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CategoryDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProductDTO struct {
	ID       int64  `json:"id"`
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Customer string `json:"customer"`
	Category string `json:"category"`
}

type CreateCustomerRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type CreateProductRequest struct {
	SKU        string `json:"sku" validate:"required,min=2,max=50"`
	Name       string `json:"name" validate:"required,min=2,max=200"`
	CustomerID int64  `json:"customer_id" validate:"required,gt=0"`
	CategoryID int64  `json:"category_id" validate:"required,gt=0"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIResponse struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}
