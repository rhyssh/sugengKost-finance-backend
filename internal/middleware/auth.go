package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	supa "github.com/rhyssh/kos-finance-backend/internal/supabase" // sesuaikan kalau module name beda
)

var (
	jwksSet jwk.Set
	jwksOnce sync.Once
	jwksErr  error
)

// getJWKS lazily loads and caches JWKS from Supabase
func getJWKS() (jwk.Set, error) {
	jwksOnce.Do(func() {
		// GANTI project ref ini dengan yang benar dari URL Supabase-mu
		// Contoh dari sebelumnya: iuamdgxkbmqtooozfxyz
		projectRef := "iuamdgxkbmqtooozfxyz" // ← PASTIKAN BENAR DI SINI

		jwksURL := "https://" + projectRef + ".supabase.co/auth/v1/.well-known/jwks.json"

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		jwksSet, jwksErr = jwk.Fetch(ctx, jwksURL)
	})

	return jwksSet, jwksErr
}

// AuthMiddleware dengan JWKS verification
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		tokenStr := parts[1]

		// Ambil JWKS
		jwks, err := getJWKS()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to load JWKS", "details": err.Error()})
			return
		}

		// Parse & verify token pakai JWKS
		token, err := jwt.Parse(
			[]byte(tokenStr),
			jwt.WithKeySet(jwks),
			jwt.WithValidate(true),
			jwt.WithIssuer("https://iuamdgxkbmqtooozfxyz.supabase.co/auth/v1"), // ← sesuaikan projectRef di sini juga
			jwt.WithAudience("authenticated"),
		)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid or expired token",
				"details": err.Error(),
			})
			return
		}

		// Ambil claims (di jwx, token sudah punya method Get)
		userIDAny, ok := token.Get("sub")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing sub claim in token"})
			return
		}

		userID, ok := userIDAny.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sub claim is not string"})
			return
		}

		c.Set("user_id", userID)

		c.Next()
	}
}

// RoleMiddleware cek role setelah auth
// RoleMiddleware sekarang query role real dari DB
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		client := supa.GetClient()

		// Di dalam RoleMiddleware
		var profiles []struct {
			Role string `json:"role"`
		}

		_, err := client.From("profiles").
			Select("role", "exact", false).
			Eq("id", userID.(string)).
			Limit(1, "").
			ExecuteTo(&profiles)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "Gagal mengambil role user",
				"details": err.Error(),
			})
			return
		}

		var userRole string
		if len(profiles) > 0 {
			userRole = profiles[0].Role
		} else {
			// Default kalau tidak ada row (atau log warning)
			userRole = "member" // atau abort kalau wajib ada
			// log.Printf("User %s tidak punya profile, default role: member", userID)
		}

		if userRole != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: diperlukan role " + requiredRole + ", kamu adalah " + userRole,
			})
			return
		}

		c.Set("user_role", userRole)
		c.Next()
	}
}