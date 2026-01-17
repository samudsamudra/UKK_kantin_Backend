package api

import (
	"github.com/gin-gonic/gin"

	adminpkg "github.com/samudsamudra/UKK_kantin/internal/api/admin"
	authpkg "github.com/samudsamudra/UKK_kantin/internal/api/auth"
	siswapkg "github.com/samudsamudra/UKK_kantin/internal/api/siswa"
	userpkg "github.com/samudsamudra/UKK_kantin/internal/api/user"

)

// --- auth / user ---
func RegisterUser(c *gin.Context) { userpkg.RegisterUser(c) }
func Login(c *gin.Context)        { authpkg.Login(c) }

// --- siswa ---
func SiswaListMenus(c *gin.Context)     { siswapkg.SiswaListMenus(c) }
func SiswaCreateOrder(c *gin.Context)   { siswapkg.SiswaCreateOrder(c) }
// func SiswaOrdersByMonth(c *gin.Context) { siswapkg.SiswaOrdersByMonth(c) }
// func SiswaGetReceiptPDF(c *gin.Context) { siswapkg.SiswaGetReceiptPDF(c) }
func SiswaGetMenu(c *gin.Context) { siswapkg.SiswaGetMenu(c) }
func SiswaGetOrderReceiptPDF(c *gin.Context) { siswapkg.SiswaGetOrderReceiptPDF(c)}

// --- admin / stan (menus) ---
func AdminCreateMenu(c *gin.Context) { adminpkg.AdminCreateMenu(c) }
func AdminUpdateMenu(c *gin.Context) { adminpkg.AdminUpdateMenu(c) }
func AdminDeleteMenu(c *gin.Context) { adminpkg.AdminDeleteMenu(c) }
func AdminListMenus(c *gin.Context)  { adminpkg.AdminListMenus(c) }
func AdminGetMenu(c *gin.Context)    { adminpkg.AdminGetMenu(c) }

// --- admin / stan (discounts) ---
func AdminCreateDiscount(c *gin.Context) { adminpkg.AdminCreateDiscount(c) }
func AdminListDiscounts(c *gin.Context)  { adminpkg.AdminListDiscounts(c) }
func AdminGetDiscount(c *gin.Context)    { adminpkg.AdminGetDiscount(c) }
func AdminUpdateDiscount(c *gin.Context) { adminpkg.AdminUpdateDiscount(c) }
func AdminDeleteDiscount(c *gin.Context) { adminpkg.AdminDeleteDiscount(c) }

// --- admin / stan (orders & reports) ---
func AdminUpdateOrderStatus(c *gin.Context) { adminpkg.AdminUpdateOrderStatus(c) }
func AdminMonthlyReport(c *gin.Context)     { adminpkg.AdminMonthlyReport(c) }
func AdminRekapTransaksi(c *gin.Context)		{ adminpkg.AdminRekapTransaksi(c) }



// --- admin / stan (stan management) ---
func RegisterStan(c *gin.Context) { adminpkg.RegisterStan(c) }

func SiswaGetWallet(c *gin.Context) { siswapkg.SiswaGetWallet(c) }
func SiswaTopupByAdmin(c *gin.Context) { siswapkg.SiswaTopupByAdmin(c) }

func SiswaOrdersByMonth(c *gin.Context) { siswapkg.SiswaOrdersByMonth(c) }

