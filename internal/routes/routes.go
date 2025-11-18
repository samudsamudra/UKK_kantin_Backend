package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/samudsamudra/UKK_kantin/internal/api"
)

// Register mendaftarkan semua route API ke engine Gin.
func Register(r *gin.Engine) {
	apiGroup := r.Group("/api")

	// AUTH & USER
	apiGroup.POST("/auth/register", api.RegisterUser)
	apiGroup.POST("/auth/login", api.Login) // jika login belum implement, pastikan api.Login ada (placeholder ok)

	// SISWA
	siswa := apiGroup.Group("/siswa")
	{
		// public
		siswa.GET("/menus", api.SiswaListMenus)

		// protected siswa (JWTAuth & RequireRole harus ada di internal/api)
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

	// public: calon pemilik stan register
	admin.POST("/stan/register", api.RegisterStan)

	// admin protected
	adminAuth := admin.Group("")
	adminAuth.Use(api.JWTAuth(), api.RequireRole("admin_stan"))
	{
		// menu CRUD
		adminAuth.POST("/menus", api.AdminCreateMenu)
		adminAuth.PUT("/menus/:id", api.AdminUpdateMenu)
		adminAuth.DELETE("/menus/:id", api.AdminDeleteMenu)
		adminAuth.GET("/menus", api.AdminListMenus)

		// orders
		adminAuth.PATCH("/orders/:id/status", api.AdminUpdateOrderStatus)
		adminAuth.GET("/orders", api.AdminOrdersByMonth)

		// reports
		adminAuth.GET("/reports/monthly", api.AdminMonthlyReport)
	}
}
