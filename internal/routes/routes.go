package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/samudsamudra/UKK_kantin/internal/api"
)

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

func requireJSON(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
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

func Register(r *gin.Engine) {
	apiGroup := r.Group("/api")

	apiGroup.Use(requireJSON(1<<20))
	apiGroup.Use(simpleRateLimiter(200, time.Minute))

	apiGroup.POST("/auth/register", simpleRateLimiter(5, 2*time.Minute), api.RegisterUser)
	apiGroup.POST("/auth/login", simpleRateLimiter(10, 1*time.Minute), api.Login)

	// SISWA
	siswa := apiGroup.Group("/siswa")
	{
		// public: browse menus
		siswa.GET("/menus", api.SiswaListMenus)
		siswa.GET("/menus/:id", api.SiswaGetMenu)

		// wallet and topup should be protected:
		// - wallet: require siswa
		// - topup: require admin (or operator) — left here but recommended to protect
		// move sensitive endpoints under auth group below

		// protected siswa actions
		siswaAuth := siswa.Group("")
		siswaAuth.Use(api.JWTAuth(), api.RequireRole("siswa"))
		{
			siswaAuth.GET("/wallet", api.SiswaGetWallet)
			siswaAuth.POST("/order", api.SiswaCreateOrder) // <-- only registered here
			// siswaAuth.GET("/orders", api.SiswaOrdersByMonth)
			// siswaAuth.GET("/order/:id/receipt", api.SiswaGetReceiptPDF)
		}

		// if topup is admin-only, register under admin routes instead; for dev you can keep it here but protect:
		// siswa.POST("/topup", api.SiswaTopupByAdmin) // DO NOT expose publicly in prod
	}

	// ADMIN STAN
	admin := apiGroup.Group("/admin")

	admin.POST("/stan/register", api.JWTAuth(), api.RequireRole("super_admin"), api.RegisterStan)

	adminAuth := admin.Group("")
	adminAuth.Use(api.JWTAuth(), api.RequireRole("admin_stan"))
	{
		adminAuth.POST("/menus", api.AdminCreateMenu)
		adminAuth.PUT("/menus/:id", api.AdminUpdateMenu)
		adminAuth.DELETE("/menus/:id", api.AdminDeleteMenu)
		adminAuth.GET("/menus", api.AdminListMenus)
		adminAuth.GET("/menus/:id", api.AdminGetMenu)

		adminAuth.PATCH("/discounts", api.AdminCreateDiscount)
		adminAuth.GET("/discounts", api.AdminListDiscounts)
		adminAuth.GET("/discounts/:id", api.AdminGetDiscount)
		adminAuth.PUT("/discounts/:id", api.AdminUpdateDiscount)
		adminAuth.DELETE("/discounts/:id", api.AdminDeleteDiscount)
		adminAuth.PATCH("/discounts/:id/menus", func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
		})

		adminAuth.PATCH("/orders/:id/status", api.AdminUpdateOrderStatus)
		adminAuth.GET("/orders", api.AdminOrdersByMonth)

		adminAuth.GET("/reports/monthly", api.AdminMonthlyReport)
	}
}
