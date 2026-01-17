package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/samudsamudra/UKK_kantin/internal/api"
)

//
// =========================
// Middleware Helpers
// =========================
//

// simpleRateLimiter limits requests per IP in a fixed window.
// NOTE: This is sufficient for UKK / demo purpose (in-memory).
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
			store[ip] = &entry{
				count:     1,
				expiresAt: time.Now().Add(window),
			}
		} else {
			e.count++
			if e.count > max {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "too many requests",
				})
				return
			}
		}
		c.Next()
	}
}

// requireJSON enforces application/json for write methods
// and limits request body size.
func requireJSON(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodPatch {

			if !strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
					"error": "content-type must be application/json",
				})
				return
			}
			c.Request.Body = http.MaxBytesReader(
				c.Writer,
				c.Request.Body,
				maxBytes,
			)
		}
		c.Next()
	}
}

//
// =========================
// Route Registration
// =========================
//

func Register(r *gin.Engine) {
	// root API group
	apiGroup := r.Group("/api")

	// global protections
	apiGroup.Use(requireJSON(1 << 20))                // 1 MB JSON
	apiGroup.Use(simpleRateLimiter(200, time.Minute)) // general rate limit

	// =========================
	// AUTH
	// =========================
	apiGroup.POST(
		"/auth/register",
		simpleRateLimiter(5, 2*time.Minute),
		api.RegisterUser,
	)
	apiGroup.POST(
		"/auth/login",
		simpleRateLimiter(10, time.Minute),
		api.Login,
	)

	// =========================
	// SISWA
	// =========================
	siswa := apiGroup.Group("/siswa")

	// public endpoints (no auth)
	siswa.GET("/menus", api.SiswaListMenus)
	siswa.GET("/menus/:id", api.SiswaGetMenu)

	// protected siswa endpoints
	siswaAuth := siswa.Group("")
	siswaAuth.Use(api.JWTAuth(), api.RequireRole("siswa"))
	{
		// wallet
		siswaAuth.GET("/wallet", api.SiswaGetWallet)

		// order
		siswaAuth.POST("/order", api.SiswaCreateOrder)

		// GET /api/siswa/orders?month=YYYY-MM
		siswaAuth.GET("/orders", api.SiswaOrdersByMonth)

		// receipt
		siswaAuth.GET("/orders/:id/receipt/pdf", api.SiswaGetOrderReceiptPDF)

		// (UKK opsional lanjutan)
		// siswaAuth.GET("/orders/:id/receipt", api.SiswaGetReceipt)
	}

	// =========================
	// ADMIN STAN
	// =========================
	admin := apiGroup.Group("/admin")

	// NOTE:
	// register stan biasanya oleh super admin.
	// Untuk UKK, endpoint ini boleh ada meski role super_admin belum diaktifkan penuh.
	admin.POST(
		"/stan/register",
		api.JWTAuth(),
		api.RequireRole("super_admin"),
		api.RegisterStan,
	)

	adminAuth := admin.Group("")
	adminAuth.Use(api.JWTAuth(), api.RequireRole("admin_stan"))
	{
		// ----- menu -----
		adminAuth.POST("/menus", api.AdminCreateMenu)
		adminAuth.PUT("/menus/:id", api.AdminUpdateMenu)
		adminAuth.DELETE("/menus/:id", api.AdminDeleteMenu)
		adminAuth.GET("/menus", api.AdminListMenus)
		adminAuth.GET("/menus/:id", api.AdminGetMenu)

		// ----- discount -----
		adminAuth.PATCH("/discounts", api.AdminCreateDiscount)
		adminAuth.GET("/discounts", api.AdminListDiscounts)
		adminAuth.GET("/discounts/:id", api.AdminGetDiscount)
		adminAuth.PUT("/discounts/:id", api.AdminUpdateDiscount)
		adminAuth.DELETE("/discounts/:id", api.AdminDeleteDiscount)

		// ----- orders -----
		adminAuth.PATCH("/orders/:id/status", api.AdminUpdateOrderStatus)

		// ----- reports -----
		// adminAuth.GET("/reports/monthly", api.AdminMonthlyReport)
		adminAuth.GET("/reports/rekap", api.AdminRekapTransaksi)
	}
}
