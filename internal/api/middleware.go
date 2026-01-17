package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

//
// =========================
// JWT Secret
// =========================
//

func jwtSecret() []byte {
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		// dev fallback only
		sec = "dev_jwt_secret_change_me"
	}
	return []byte(sec)
}

//
// =========================
// JWT Middleware (SECURE)
// =========================
//
// Prinsip:
// - JWT hanya bukti identitas
// - Role diambil dari DATABASE
// - Payload JWT TIDAK dipercaya
//

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			return
		}

		parts := strings.Fields(h)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header",
			})
			return
		}
		tokenStr := parts[1]

		claims := &jwt.RegisteredClaims{}

		token, err := jwt.ParseWithClaims(
			tokenStr,
			claims,
			func(t *jwt.Token) (interface{}, error) {
				// 🔒 HARD CHECK SIGNING METHOD
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return jwtSecret(), nil
			},
		)

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or tampered token",
			})
			return
		}

		// =========================
		// VALIDATE SUBJECT
		// =========================
		if claims.Subject == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token subject",
			})
			return
		}

		// =========================
		// LOAD USER FROM DATABASE
		// =========================
		var user app.User
		if err := app.DB.
			Where("public_id = ?", claims.Subject).
			First(&user).Error; err != nil {

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "user not found",
			})
			return
		}

		// =========================
		// SET CONTEXT (SAFE)
		// =========================
		c.Set("user_id", user.ID)
		c.Set("public_id", user.PublicID)
		c.Set("role", fmt.Sprintf("%v", user.Role)) // ⬅️ EXPLICIT STRING CONVERSION

		c.Next()
	}
}

//
// =========================
// ROLE GUARD (STRING-BASED, AMAN)
// =========================
//

func RequireRole(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rv, ok := c.Get("role")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden",
			})
			return
		}

		role, ok := rv.(string)
		if !ok || role != expected {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden",
			})
			return
		}

		c.Next()
	}
}

// =========================
// SUPER ADMIN GUARD (HARD)
// =========================
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		rv, ok := c.Get("role")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden",
			})
			return
		}

		role, ok := rv.(string)
		if !ok || role != "super_admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "super admin only",
			})
			return
		}

		c.Next()
	}
}
