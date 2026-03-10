package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"smart-inventory-backend/internal/models"
	"smart-inventory-backend/internal/repo"
	"smart-inventory-backend/internal/service"
)

type Handler struct {
	svc    *service.Service
	val    *validator.Validate
	apiKey string
}

func New(svc *service.Service, apiKey string) *Handler {
	return &Handler{svc: svc, val: validator.New(), apiKey: apiKey}
}

func (h *Handler) Router() http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestID())
	r.Use(requestLogger())
	r.Use(cors())
	r.Use(h.apiKeyAuth())

	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.OPTIONS("/*path", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	r.GET("/inventory", h.listInventory)
	r.GET("/inventory/adjustments", h.listAdjustments)
	r.POST("/inventory/adjust", h.adjustInventory)

	r.GET("/customers", h.listCustomers)
	r.POST("/customers", h.createCustomer)

	r.GET("/categories", h.listCategories)
	r.POST("/categories", h.createCategory)

	r.GET("/products", h.listProducts)
	r.POST("/products", h.createProduct)

	r.GET("/stock-ins", h.listStockIns)
	r.POST("/stock-ins", h.createStockIn)
	r.POST("/stock-ins/:code/status", h.updateStockInStatus)
	r.GET("/stock-ins/code/:code", h.getStockInByCode)
	r.DELETE("/stock-ins/:code", h.deleteStockIn)

	r.GET("/stock-outs", h.listStockOuts)
	r.POST("/stock-outs", h.createStockOut)
	r.POST("/stock-outs/:code/allocate", h.allocateStockOut)
	r.POST("/stock-outs/:code/status", h.updateStockOutStatus)
	r.GET("/stock-outs/code/:code", h.getStockOutByCode)
	r.DELETE("/stock-outs/:code", h.deleteStockOut)

	r.GET("/reports/stock-ins", h.reportStockIn)
	r.GET("/reports/stock-outs", h.reportStockOut)

	return r
}

func SetGinWriters(w io.Writer) {
	if w == nil {
		return
	}
	gin.DefaultWriter = w
	gin.DefaultErrorWriter = w
}

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = newRequestID()
		}
		c.Set("request_id", rid)
		c.Writer.Header().Set("X-Request-Id", rid)
		c.Next()
	}
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		ip := c.ClientIP()
		ua := c.Request.UserAgent()
		rid, _ := c.Get("request_id")
		errMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()
		if errMsg == "" {
			errMsg = "-"
		}
		gin.DefaultWriter.Write([]byte(
			time.Now().Format(time.RFC3339) + " " +
				method + " " + path + " " +
				"status=" + strconv.Itoa(status) + " " +
				"latency=" + latency.String() + " " +
				"ip=" + ip + " " +
				"rid=" + toString(rid) + " " +
				"err=" + errMsg + " " +
				"ua=\"" + ua + "\"\n",
		))
	}
}

func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b)
}

func toString(v any) string {
	if v == nil {
		return "-"
	}
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return "-"
}

func (h *Handler) apiKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		if h.apiKey == "" {
			c.Next()
			return
		}
		key := c.GetHeader("X-API-Key")
		if key == "" || key != h.apiKey {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Error:   &APIError{Code: "unauthorized", Message: "API key tidak valid"},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Request-Id, X-Requested-With, Authorization, Accept")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Request-Id")
		c.Writer.Header().Set("Access-Control-Max-Age", "600")
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (h *Handler) listInventory(c *gin.Context) {
	filter := repo.InventoryFilter{
		Name:     c.Query("name"),
		SKU:      c.Query("sku"),
		Customer: c.Query("customer"),
	}
	items, err := h.svc.ListInventory(c.Request.Context(), filter)
	if err != nil {
		writeError(c, err)
		return
	}
	resp := make([]InventoryResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, InventoryResponse{
			SKU:          it.SKU,
			Name:         it.Name,
			Customer:     it.Customer,
			Category:     it.Category,
			PhysicalQty:  it.PhysicalQty,
			AvailableQty: it.AvailableQty,
			UpdatedAt:    it.LastUpdatedAt,
		})
	}
	writeSuccess(c, http.StatusOK, resp)
}

func (h *Handler) adjustInventory(c *gin.Context) {
	var req AdjustInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid_body", "invalid body")
		return
	}
	if err := h.val.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}
	if err := h.svc.AdjustStock(c.Request.Context(), req.SKU, req.NewPhysical, req.Reason); err != nil {
		writeError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) listAdjustments(c *gin.Context) {
	sku := c.Query("sku")
	if sku == "" {
		writeBadRequest(c, "invalid_sku", "sku wajib diisi")
		return
	}
	items, err := h.svc.ListAdjustments(c.Request.Context(), sku)
	if err != nil {
		writeError(c, err)
		return
	}
	resp := make([]StockAdjustmentDTO, 0, len(items))
	for _, it := range items {
		resp = append(resp, StockAdjustmentDTO{
			SKU:       it.SKU,
			Delta:     it.Delta,
			Reason:    it.Reason,
			Source:    it.Source,
			RefCode:   it.RefCode,
			Bucket:    it.Bucket,
			CreatedAt: it.CreatedAt,
		})
	}
	writeSuccess(c, http.StatusOK, resp)
}

func (h *Handler) createStockIn(c *gin.Context) {
	var req CreateStockInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid_body", "invalid body")
		return
	}
	if err := h.val.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}
	items := make([]models.LineItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, models.LineItem{SKU: it.SKU, Qty: it.Qty})
	}
	code, uuid, err := h.svc.CreateStockIn(c.Request.Context(), req.Code, items)
	if err != nil {
		writeError(c, err)
		return
	}
	writeSuccess(c, http.StatusCreated, gin.H{"code": code, "uuid": uuid})
}

func (h *Handler) updateStockInStatus(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		writeBadRequest(c, "invalid_code", "nomor dokumen wajib diisi")
		return
	}
	var body UpdateStatusRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		writeBadRequest(c, "invalid_body", "invalid body")
		return
	}
	if err := h.val.Struct(body); err != nil {
		writeValidationError(c, err)
		return
	}
	if err := h.svc.UpdateStockInStatus(c.Request.Context(), code, body.Status); err != nil {
		writeError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) createStockOut(c *gin.Context) {
	var req CreateStockOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid_body", "invalid body")
		return
	}
	if err := h.val.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}
	items := make([]models.LineItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, models.LineItem{SKU: it.SKU, Qty: it.Qty})
	}
	code, uuid, err := h.svc.CreateStockOut(c.Request.Context(), req.Code, items)
	if err != nil {
		writeError(c, err)
		return
	}
	writeSuccess(c, http.StatusCreated, gin.H{"code": code, "uuid": uuid})
}

func (h *Handler) deleteStockIn(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		writeBadRequest(c, "invalid_code", "nomor dokumen wajib diisi")
		return
	}
	if err := h.svc.DeleteStockIn(c.Request.Context(), code); err != nil {
		writeError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) deleteStockOut(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		writeBadRequest(c, "invalid_code", "nomor dokumen wajib diisi")
		return
	}
	if err := h.svc.DeleteStockOut(c.Request.Context(), code); err != nil {
		writeError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) allocateStockOut(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		writeBadRequest(c, "invalid_code", "nomor dokumen wajib diisi")
		return
	}
	if err := h.svc.AllocateStockOut(c.Request.Context(), code); err != nil {
		writeError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) updateStockOutStatus(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		writeBadRequest(c, "invalid_code", "nomor dokumen wajib diisi")
		return
	}
	var body UpdateStatusRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		writeBadRequest(c, "invalid_body", "invalid body")
		return
	}
	if err := h.val.Struct(body); err != nil {
		writeValidationError(c, err)
		return
	}
	if err := h.svc.UpdateStockOutStatus(c.Request.Context(), code, body.Status); err != nil {
		writeError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) listCustomers(c *gin.Context) {
	items, err := h.svc.ListCustomers(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	resp := make([]CustomerDTO, 0, len(items))
	for _, it := range items {
		resp = append(resp, CustomerDTO{ID: it.ID, Name: it.Name})
	}
	writeSuccess(c, http.StatusOK, resp)
}

func (h *Handler) createCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid_body", "invalid body")
		return
	}
	if err := h.val.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}
	id, err := h.svc.CreateCustomer(c.Request.Context(), req.Name)
	if err != nil {
		writeError(c, err)
		return
	}
	writeSuccess(c, http.StatusCreated, gin.H{"id": id})
}

func (h *Handler) listCategories(c *gin.Context) {
	items, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	resp := make([]CategoryDTO, 0, len(items))
	for _, it := range items {
		resp = append(resp, CategoryDTO{ID: it.ID, Name: it.Name})
	}
	writeSuccess(c, http.StatusOK, resp)
}

func (h *Handler) createCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid_body", "invalid body")
		return
	}
	if err := h.val.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}
	id, err := h.svc.CreateCategory(c.Request.Context(), req.Name)
	if err != nil {
		writeError(c, err)
		return
	}
	writeSuccess(c, http.StatusCreated, gin.H{"id": id})
}

func (h *Handler) listProducts(c *gin.Context) {
	items, err := h.svc.ListProducts(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	resp := make([]ProductDTO, 0, len(items))
	for _, it := range items {
		resp = append(resp, ProductDTO{
			ID:       it.ID,
			SKU:      it.SKU,
			Name:     it.Name,
			Customer: it.Customer,
			Category: it.Category,
		})
	}
	writeSuccess(c, http.StatusOK, resp)
}

func (h *Handler) createProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid_body", "invalid body")
		return
	}
	if err := h.val.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}
	id, err := h.svc.CreateProduct(c.Request.Context(), req.SKU, req.Name, req.CustomerID, req.CategoryID)
	if err != nil {
		writeError(c, err)
		return
	}
	writeSuccess(c, http.StatusCreated, gin.H{"id": id})
}

func (h *Handler) listStockIns(c *gin.Context) {
	statuses := parseStatusQuery(c.Query("status"))
	items, err := h.svc.ListStockInByStatus(c.Request.Context(), statuses)
	if err != nil {
		writeError(c, err)
		return
	}
	resp := make([]TransactionResponse, 0, len(items))
	for _, it := range items {
		lines := make([]StockLineDTO, 0, len(it.Items))
		for _, li := range it.Items {
			lines = append(lines, StockLineDTO{SKU: li.SKU, Qty: li.Qty})
		}
		resp = append(resp, TransactionResponse{
			UUID:      it.UUID,
			Code:      it.Code,
			Status:    it.Status,
			Items:     lines,
			CreatedAt: it.CreatedAt,
			UpdatedAt: it.UpdatedAt,
		})
	}
	writeSuccess(c, http.StatusOK, resp)
}
func (h *Handler) getStockInByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		writeBadRequest(c, "invalid_code", "nomor dokumen wajib diisi")
		return
	}
	tx, err := h.svc.GetStockInByCode(c.Request.Context(), code)
	if err != nil {
		writeError(c, err)
		return
	}
	lines := make([]StockLineDTO, 0, len(tx.Items))
	for _, it := range tx.Items {
		lines = append(lines, StockLineDTO{SKU: it.SKU, Qty: it.Qty})
	}
	writeSuccess(c, http.StatusOK, TransactionResponse{
		UUID:      tx.UUID,
		Code:      tx.Code,
		Status:    tx.Status,
		Items:     lines,
		CreatedAt: tx.CreatedAt,
		UpdatedAt: tx.UpdatedAt,
	})
}

func (h *Handler) listStockOuts(c *gin.Context) {
	statuses := parseStatusQuery(c.Query("status"))
	items, err := h.svc.ListStockOutByStatus(c.Request.Context(), statuses)
	if err != nil {
		writeError(c, err)
		return
	}
	resp := make([]TransactionResponse, 0, len(items))
	for _, it := range items {
		lines := make([]StockLineDTO, 0, len(it.Items))
		for _, li := range it.Items {
			lines = append(lines, StockLineDTO{SKU: li.SKU, Qty: li.Qty})
		}
		resp = append(resp, TransactionResponse{
			UUID:      it.UUID,
			Code:      it.Code,
			Status:    it.Status,
			Items:     lines,
			CreatedAt: it.CreatedAt,
			UpdatedAt: it.UpdatedAt,
		})
	}
	writeSuccess(c, http.StatusOK, resp)
}

func (h *Handler) getStockOutByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		writeBadRequest(c, "invalid_code", "nomor dokumen wajib diisi")
		return
	}
	tx, err := h.svc.GetStockOutByCode(c.Request.Context(), code)
	if err != nil {
		writeError(c, err)
		return
	}
	lines := make([]StockLineDTO, 0, len(tx.Items))
	for _, it := range tx.Items {
		lines = append(lines, StockLineDTO{SKU: it.SKU, Qty: it.Qty})
	}
	writeSuccess(c, http.StatusOK, TransactionResponse{
		UUID:      tx.UUID,
		Code:      tx.Code,
		Status:    tx.Status,
		Items:     lines,
		CreatedAt: tx.CreatedAt,
		UpdatedAt: tx.UpdatedAt,
	})
}

func parseStatusQuery(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToUpper(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	switch err {
	case service.ErrInvalidStatus, service.ErrNotAllowed:
		status = http.StatusBadRequest
		code = "bad_request"
	case service.ErrInsufficient:
		status = http.StatusConflict
		code = "insufficient_stock"
	}
	c.JSON(status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: err.Error(),
		},
	})
}

func writeBadRequest(c *gin.Context, code, msg string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: msg,
		},
	})
}

func writeSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, APIResponse{
		Success: true,
		Data:    data,
	})
}

func writeValidationError(c *gin.Context, err error) {
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		writeBadRequest(c, "validation_error", "validation failed")
		return
	}
	details := make([]gin.H, 0, len(ve))
	for _, fe := range ve {
		details = append(details, gin.H{
			"field": fe.Field(),
			"tag":   fe.Tag(),
			"value": fe.Value(),
		})
	}
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "validation_error",
			Message: "validation failed",
		},
		Data: details,
	})
}
func (h *Handler) reportStockIn(c *gin.Context) {
	from, to, err := parseDateRange(c)
	if err != nil {
		writeBadRequest(c, "invalid_date", err.Error())
		return
	}
	items, err := h.svc.ReportsStockIn(c.Request.Context(), from, to)
	if err != nil {
		writeError(c, err)
		return
	}
	resp := make([]TransactionResponse, 0, len(items))
	for _, tx := range items {
		lines := make([]StockLineDTO, 0, len(tx.Items))
		for _, it := range tx.Items {
			lines = append(lines, StockLineDTO{SKU: it.SKU, Qty: it.Qty})
		}
		resp = append(resp, TransactionResponse{
			UUID:      tx.UUID,
			Code:      tx.Code,
			Status:    tx.Status,
			Items:     lines,
			CreatedAt: tx.CreatedAt,
			UpdatedAt: tx.UpdatedAt,
		})
	}
	writeSuccess(c, http.StatusOK, resp)
}

func (h *Handler) reportStockOut(c *gin.Context) {
	from, to, err := parseDateRange(c)
	if err != nil {
		writeBadRequest(c, "invalid_date", err.Error())
		return
	}
	items, err := h.svc.ReportsStockOut(c.Request.Context(), from, to)
	if err != nil {
		writeError(c, err)
		return
	}
	resp := make([]TransactionResponse, 0, len(items))
	for _, tx := range items {
		lines := make([]StockLineDTO, 0, len(tx.Items))
		for _, it := range tx.Items {
			lines = append(lines, StockLineDTO{SKU: it.SKU, Qty: it.Qty})
		}
		resp = append(resp, TransactionResponse{
			UUID:      tx.UUID,
			Code:      tx.Code,
			Status:    tx.Status,
			Items:     lines,
			CreatedAt: tx.CreatedAt,
			UpdatedAt: tx.UpdatedAt,
		})
	}
	writeSuccess(c, http.StatusOK, resp)
}

func parseDateRange(c *gin.Context) (*time.Time, *time.Time, error) {
	layout := "2006-01-02"
	var fromPtr *time.Time
	var toPtr *time.Time
	if v := c.Query("date_from"); v != "" {
		t, err := time.Parse(layout, v)
		if err != nil {
			return nil, nil, errors.New("format date_from harus YYYY-MM-DD")
		}
		fromPtr = &t
	}
	if v := c.Query("date_to"); v != "" {
		t, err := time.Parse(layout, v)
		if err != nil {
			return nil, nil, errors.New("format date_to harus YYYY-MM-DD")
		}
		t = t.Add(24 * time.Hour)
		toPtr = &t
	}
	return fromPtr, toPtr, nil
}
