package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/supabase-community/postgrest-go"

	// Tambah ini
	"github.com/rhyssh/kos-finance-backend/internal/handler"
	"github.com/rhyssh/kos-finance-backend/internal/middleware"
	"github.com/rhyssh/kos-finance-backend/internal/storage"
	"github.com/rhyssh/kos-finance-backend/internal/supabase"
	supa "github.com/rhyssh/kos-finance-backend/internal/supabase" // sesuaikan module name kalau beda
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Tidak menemukan file .env")
	}

	r := gin.Default()
	storage.InitR2()
	// CORS (sementara)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Backend Golang berjalan!",
			"env":     os.Getenv("PORT"),
		})
	})

	// Endpoint baru: test koneksi Supabase
	r.GET("/supabase-test", func(c *gin.Context) {
		client := supa.GetClient()

		// Contoh query sederhana: ambil 1 row dari table profiles (kalau ada data)
		var result []map[string]interface{}
		_, err := client.From("profiles").Select("*", "exact", false).Limit(1, "").ExecuteTo(&result)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Gagal koneksi ke Supabase",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "connected",
			"supabase_url": os.Getenv("SUPABASE_URL"),
			"sample_data":  result, // bisa kosong kalau belum ada data
		})
	})

		// Protected routes
	authGroup := r.Group("/api")
	authGroup.Use(middleware.AuthMiddleware())
	handler.RegisterPaymentRoutes(authGroup)
	handler.RegisterProfileRoutes(authGroup)
	// Contoh endpoint hanya untuk user login
	authGroup.GET("/me", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello authenticated user!",
			"user_id": userID,
		})
	})

	// Rekap publik - tanpa auth
r.GET("/api/rekap", func(c *gin.Context) {
	client := supabase.GetClient()

	var summary []struct {
		Month          string  `json:"month"`
		TotalPemasukan float64 `json:"total_pemasukan"`
		TotalPengeluaran float64 `json:"total_pengeluaran"`
	}

	_, err := client.From("public_summary").
		Select("*", "exact", false).
			Order("month", &postgrest.OrderOpts{
			Ascending: false,
		}).
		ExecuteTo(&summary)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil rekap", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rekap": summary})
})

	// Contoh endpoint khusus bendahara
	authGroup.GET("/admin-only", middleware.RoleMiddleware("bendahara"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome, Bendahara!"})
	})

	// Contoh endpoint khusus member
	authGroup.GET("/member-only", middleware.RoleMiddleware("member"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome, Member!"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server berjalan di http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal start server: %v", err)
	}
}