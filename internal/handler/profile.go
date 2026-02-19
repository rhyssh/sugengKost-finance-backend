package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhyssh/kos-finance-backend/internal/middleware"
	"github.com/rhyssh/kos-finance-backend/internal/supabase"
)

func RegisterProfileRoutes(r *gin.RouterGroup) {
	profile := r.Group("/profile")
	profile.Use(middleware.AuthMiddleware()) // login wajib

	profile.PATCH("", updateProfile)
}

func updateProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}
	userID := userIDStr.(string)

	var req struct {
		Username  string `json:"username"`
		FullName  string `json:"full_name"`
		// Email tidak diizinkan edit langsung di profile (karena Supabase Auth handle email terpisah)
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validasi sederhana
	if req.Username == "" && req.FullName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setidaknya isi satu field: username atau full_name"})
		return
	}

	client := supabase.GetClient()

	updateData := map[string]interface{}{}
	if req.Username != "" {
		updateData["username"] = req.Username
	}
	if req.FullName != "" {
		updateData["full_name"] = req.FullName
	}
	updateData["updated_at"] = time.Now().Format(time.RFC3339)

	_, _, err := client.From("profiles").
		Update(updateData, "none", "none").
		Eq("id", userID).
		Execute()


	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal update profile",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile berhasil diperbarui",
	})
}

func getProfile(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID := userIDStr.(string)

	client := supabase.GetClient()
	var profile struct {
		Username  string `json:"username"`
		FullName  string `json:"full_name"`
		Role      string `json:"role"`
		Email     string `json:"email"` // optional, kalau ada di profiles
	}
	_, err := client.From("profiles").
		Select("username, full_name, role", "exact", false).
		Eq("id", userID).
		Single().
		ExecuteTo(&profile)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	// Optional: ambil email dari auth.users kalau tidak ada di profiles
	if profile.Email == "" {
		var authUser struct {
			Email string `json:"email"`
		}
		_, err = client.From("auth.users").
			Select("email", "exact", false).
			Eq("id", userID).
			Single().
			ExecuteTo(&authUser)
		if err == nil {
			profile.Email = authUser.Email
		}
	}

	c.JSON(http.StatusOK, profile)
}