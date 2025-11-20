package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/samudsamudra/UKK_kantin/internal/api"
)

// simpleRateLimiterMap is a tiny in-memory limiter (per-IP, not for prod).
// limits N requests per window. Good as a basic DDOS/brute-force mitigation during dev.
func simpleRateLimiter(max int, window time.Duration) gin.HandlerFunc {
	type entry struct {
		count     int
		expiresAt time.Time
	}
	store := map[string]*entry{}
	return func(c *gin.Context) {
		ip := c.ClientIP()
		e, ok := store[ip]
		if !ok || time.Now().After(e.expiresAt) {
			store[ip] = &entry{count: 1, expiresAt: time.Now().Add(window)}
		} else {
			e.count++
			if e.count > max {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
				return
			}
		}
		c.Next()
	}
}

// requireJSON ensures requests that should be JSON have the right header and size.
// maxBytes = max allowed body bytes.
func requireJSON(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// only check for methods that carry a body
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
			if !strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{"error": "content-type must be application/json"})
				return
			}
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

// Register registers all API routes to the Gin engine with secured admin surface.
func Register(r *gin.Engine) {
	apiGroup := r.Group("/api")

	// global small protections
	apiGroup.Use(requireJSON(1<<20))                   // max 1MiB payload for JSON endpoints
	apiGroup.Use(simpleRateLimiter(200, time.Minute)) // protect from spam (global gentle limit)

	// AUTH & USER (public registration/login but rate-limited)
	// keep register/login public but protected by rate limiter to avoid abuse
	apiGroup.POST("/auth/register", simpleRateLimiter(5, 2*time.Minute), api.RegisterUser)
	apiGroup.POST("/auth/login", simpleRateLimiter(10, 1*time.Minute), api.Login)

	// SISWA
	siswa := apiGroup.Group("/siswa")
	{
		// public: list menu (browse)
		siswa.GET("/menus", api.SiswaListMenus)

		// protected siswa actions
		siswaAuth := siswa.Group("")
		siswaAuth.Use(api.JWTAuth(), api.RequireRole("siswa"))
		{
			siswaAuth.POST("/order", api.SiswaCreateOrder)
			siswaAuth.GET("/orders", api.SiswaOrdersByMonth)
			siswaAuth.GET("/order/:id/receipt", api.SiswaGetReceiptPDF)
		}
	}

	// ADMIN STAN
	admin := apiGroup.Group("/admin")

	// ---- IMPORTANT: STAN REGISTRATION ----
	// Only super_admin should be allowed to create admin stan accounts.
	// NOTE: create initial super_admin via seed/script/env before locking this in production.
	admin.POST("/stan/register", api.JWTAuth(), api.RequireRole("super_admin"), api.RegisterStan)

	// Protected admin actions (only admin_stan role)
	adminAuth := admin.Group("")
	adminAuth.Use(api.JWTAuth(), api.RequireRole("admin_stan"))
	{
		// menu CRUD (admin only for their own stan)
		adminAuth.POST("/menus", api.AdminCreateMenu)
		adminAuth.PUT("/menus/:id", api.AdminUpdateMenu)
		adminAuth.DELETE("/menus/:id", api.AdminDeleteMenu)
		adminAuth.GET("/menus", api.AdminListMenus)
		adminAuth.GET("/menus/:id", api.AdminGetMenu)

		// discounts CRUD
		adminAuth.PATCH("/discounts", api.AdminCreateDiscount)
		adminAuth.GET("/discounts", api.AdminListDiscounts)
		adminAuth.GET("/discounts/:id", api.AdminGetDiscount)
		adminAuth.PUT("/discounts/:id", api.AdminUpdateDiscount)
		adminAuth.DELETE("/discounts/:id", api.AdminDeleteDiscount)
		adminAuth.PATCH("/discounts/:id/menus", func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
		})
		
		// orders - admin confirms/updates status
		adminAuth.PATCH("/orders/:id/status", api.AdminUpdateOrderStatus)
		adminAuth.GET("/orders", api.AdminOrdersByMonth)

		// reports - monthly recap
		adminAuth.GET("/reports/monthly", api.AdminMonthlyReport)
	}
}
