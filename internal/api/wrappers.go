package api

import (
	"github.com/gin-gonic/gin"

	authpkg "github.com/samudsamudra/UKK_kantin/internal/api/auth"
	adminpkg "github.com/samudsamudra/UKK_kantin/internal/api/admin"
	siswapkg "github.com/samudsamudra/UKK_kantin/internal/api/siswa"
	userpkg "github.com/samudsamudra/UKK_kantin/internal/api/user"
)

// --- auth / user ---
func RegisterUser(c *gin.Context) { userpkg.RegisterUser(c) }
func Login(c *gin.Context)        { authpkg.Login(c) }

// --- siswa ---
func SiswaListMenus(c *gin.Context)     { siswapkg.SiswaListMenus(c) }
func SiswaCreateOrder(c *gin.Context)   { siswapkg.SiswaCreateOrder(c) }
func SiswaOrdersByMonth(c *gin.Context) { siswapkg.SiswaOrdersByMonth(c) }
func SiswaGetReceiptPDF(c *gin.Context) { siswapkg.SiswaGetReceiptPDF(c) }

// --- admin / stan ---
func RegisterStan(c *gin.Context)            { adminpkg.RegisterStan(c) }
func AdminCreateMenu(c *gin.Context)         { adminpkg.AdminCreateMenu(c) }
func AdminUpdateMenu(c *gin.Context)         { adminpkg.AdminUpdateMenu(c) }
func AdminDeleteMenu(c *gin.Context)         { adminpkg.AdminDeleteMenu(c) }
func AdminListMenus(c *gin.Context)          { adminpkg.AdminListMenus(c) }
func AdminUpdateOrderStatus(c *gin.Context)  { adminpkg.AdminUpdateOrderStatus(c) }
func AdminOrdersByMonth(c *gin.Context)      { adminpkg.AdminOrdersByMonth(c) }
func AdminMonthlyReport(c *gin.Context)      { adminpkg.AdminMonthlyReport(c) }
