package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhyssh/kos-finance-backend/internal/middleware"
	"github.com/rhyssh/kos-finance-backend/internal/storage"
	"github.com/rhyssh/kos-finance-backend/internal/supabase"
	"github.com/supabase-community/postgrest-go"
)

// alias supabase, bukan supa

func RegisterPaymentRoutes(r *gin.RouterGroup) {
	payment := r.Group("/payments")
	payment.Use(middleware.AuthMiddleware()) // login wajib untuk semua di /payments

	// Semua authenticated boleh upload & lihat bukti & list payments
	payment.POST("", createPayment)
	payment.GET("/:id/bukti", getBukti)
	payment.GET("/all", listAllPayments)     // nanti kita buat ini untuk list semua (pending/verified/rejected)
	payment.GET("/pending", listPendingPayments)
	payment.GET("/list-pengeluaran", listExpenses)

	// Hanya bendahara yang boleh verifikasi
	verifyGroup := payment.Group("")
	verifyGroup.Use(middleware.RoleMiddleware("bendahara"))
	verifyGroup.PATCH("/:id/verify", verifyPayment)

	// Nanti kalau sudah buat expenses, bisa tambah grup serupa
	expenses := r.Group("/expenses")
	expenses.Use(middleware.RoleMiddleware("bendahara"))
	expenses.POST("", createExpense)
}

func createPayment(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	amountStr := c.PostForm("amount")
	monthYear := c.PostForm("month_year")
	notes := c.PostForm("notes")

	if amountStr == "" || monthYear == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount dan month_year wajib"})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount harus angka valid"})
		return
	}

	file, err := c.FormFile("bukti")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File bukti wajib diupload"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka file"})
		return
	}
	defer src.Close()

	r2 := storage.GetR2()
	buktiKey, err := r2.UploadFile(c.Request.Context(), src, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload ke R2", "details": err.Error()})
		return
	}

	client := supabase.GetClient()
	paymentData := map[string]interface{}{
		"user_id":      userID.String(),
		"amount":       amount,
		"month_year":   monthYear,
		"status":       "pending",
		"bukti_key":    buktiKey,
		"notes":        notes,
		"payment_date": time.Now().Format("2006-01-02"),
	}

	_, _, err = client.From("payments").
		Insert(paymentData, false, "", "none", "none").
		Execute()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal simpan ke DB",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Bukti bayar berhasil diupload dan disimpan",
		"bukti_key": buktiKey,
	})
}

func getBukti(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	userIDStr, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	client := supabase.GetClient()
	var payment struct {
		UserID   string `json:"user_id"`
		BuktiKey string `json:"bukti_key"`
	}
	_, err = client.From("payments").
		Select("user_id, bukti_key", "exact", false).
		Eq("id", id.String()).
		Single().
		ExecuteTo(&payment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found or no access"})
		return
	}

	if userRole != "bendahara" && payment.UserID != userIDStr.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tidak berhak akses bukti ini"})
		return
	}

	r2 := storage.GetR2()
	presignedURL, err := r2.GetPresignedURL(c.Request.Context(), payment.BuktiKey, time.Hour*24)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate link bukti", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bukti_url": presignedURL,
	})
}

func listPendingPayments(c *gin.Context) {
	client := supabase.GetClient()

	var payments []struct {
		ID          string  `json:"id"`
		UserID      string  `json:"user_id"`
		Amount      float64 `json:"amount"`
		MonthYear   string  `json:"month_year"`
		Status      string  `json:"status"`
		BuktiKey    string  `json:"bukti_key"`
		Notes       string  `json:"notes"`
		PaymentDate string  `json:"payment_date"`
		CreatedAt   string  `json:"created_at"`
	}

	_, err := client.From("payments").
		Select("*", "exact", false).
		Eq("status", "pending").
		Order("created_at", &postgrest.OrderOpts{
			Ascending: true,
		}).  // fix: postgrest.Asc
		ExecuteTo(&payments)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengambil list pending",
			"details": err.Error(),
		})
		return
	}

	r2 := storage.GetR2()
	response := []map[string]interface{}{}
	for _, p := range payments {
		presignedURL := ""
		if p.BuktiKey != "" {
			url, err := r2.GetPresignedURL(c.Request.Context(), p.BuktiKey, time.Hour*24)
			if err == nil {
				presignedURL = url
			}
		}

		response = append(response, map[string]interface{}{
			"id":           p.ID,
			"user_id":      p.UserID,
			"amount":       p.Amount,
			"month_year":   p.MonthYear,
			"status":       p.Status,
			"bukti_url":    presignedURL,
			"notes":        p.Notes,
			"payment_date": p.PaymentDate,
			"created_at":   p.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": response,
	})
}

func verifyPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes,omitempty"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Status != "verified" && req.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status harus verified atau rejected"})
		return
	}

	client := supabase.GetClient()
	updateData := map[string]interface{}{
		"status":     req.Status,
		"notes":      req.Notes,
		"updated_at": time.Now().Format(time.RFC3339),
	}

	_, _, err = client.From("payments").
		Update(updateData, "none", "none").  // fix: false (upsert), "none" (returning)
		Eq("id", id.String()).
		Execute()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal update status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment berhasil diverifikasi",
		"status":  req.Status,
	})
}

func listAllPayments(c *gin.Context) {
	client := supabase.GetClient()

	var payments []struct {
		ID          string  `json:"id"`
		UserID      string  `json:"user_id"`
		Amount      float64 `json:"amount"`
		MonthYear   string  `json:"month_year"`
		Status      string  `json:"status"`
		BuktiKey    string  `json:"bukti_key"`
		Notes       string  `json:"notes"`
		PaymentDate string  `json:"payment_date"`
		CreatedAt   string  `json:"created_at"`
	}

	_, err := client.From("payments").
		Select("*", "exact", false).
		Order("created_at", &postgrest.OrderOpts{
			Ascending: false,
		}). // urut terbaru dulu
		ExecuteTo(&payments)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengambil list payments",
			"details": err.Error(),
		})
		return
	}

	r2 := storage.GetR2()
	response := []map[string]interface{}{}
	for _, p := range payments {
		presignedURL := ""
		if p.BuktiKey != "" {
			url, err := r2.GetPresignedURL(c.Request.Context(), p.BuktiKey, time.Hour*24)
			if err == nil {
				presignedURL = url
			}
		}

		response = append(response, map[string]interface{}{
			"id":           p.ID,
			"user_id":      p.UserID,
			"amount":       p.Amount,
			"month_year":   p.MonthYear,
			"status":       p.Status,
			"bukti_url":    presignedURL,
			"notes":        p.Notes,
			"payment_date": p.PaymentDate,
			"created_at":   p.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": response,
	})
}

// createExpense - hanya bendahara
func createExpense(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	amountStr := c.PostForm("amount")
	description := c.PostForm("description")
	expenseDate := c.PostForm("expense_date") // optional, default now

	if amountStr == "" || description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount dan description wajib"})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount harus angka"})
		return
	}

	var buktiKey string
	file, err := c.FormFile("bukti_key")
	if err == nil {
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal buka file bukti"})
			return
		}
		defer src.Close()

		r2 := storage.GetR2()
		buktiKey, err = r2.UploadFile(c.Request.Context(), src, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload bukti", "details": err.Error()})
			return
		}
	}

	client := supabase.GetClient()
	expenseData := map[string]interface{}{
		"amount":       amount,
		"description":  description,
		"expense_date": expenseDate,
		"bukti_key":    buktiKey,
		"created_by":   userID.String(),
	}

	_, _, err = client.From("expenses").
		Insert(expenseData, false, "", "none", "none").
		Execute()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal simpan pengeluaran", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pengeluaran berhasil dicatat"})
}

// listExpenses - semua authenticated
func listExpenses(c *gin.Context) {
	client := supabase.GetClient()

	var expenses []struct {
		ID          string  `json:"id"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		ExpenseDate string  `json:"expense_date"`
		BuktiKey    string  `json:"bukti_key"`
		CreatedBy   string  `json:"created_by"`
		CreatedAt   string  `json:"created_at"`
	}

	_, err := client.From("expenses").
		Select("*", "exact", false).
		Order("expense_date", &postgrest.OrderOpts{
			Ascending: false,
		}).
		ExecuteTo(&expenses)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil list pengeluaran", "details": err.Error()})
		return
	}

	r2 := storage.GetR2()
	response := []map[string]interface{}{}
	for _, e := range expenses {
		presignedURL := ""
		if e.BuktiKey != "" {
			url, err := r2.GetPresignedURL(c.Request.Context(), e.BuktiKey, time.Hour*24)
			if err == nil {
				presignedURL = url
			}
		}

		response = append(response, map[string]interface{}{
			"id":           e.ID,
			"amount":       e.Amount,
			"description":  e.Description,
			"expense_date": e.ExpenseDate,
			"bukti_url":    presignedURL,
			"created_by":   e.CreatedBy,
			"created_at":   e.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"expenses": response})
}