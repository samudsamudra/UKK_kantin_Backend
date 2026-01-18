package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

//
// =========================
// Payload
// =========================
//

type clearDBPayload struct {
	Confirm string `json:"confirm" binding:"required"`
}

//
// =========================
// Handler
// =========================
//

// AdminClearDatabase
// POST /api/admin/system/clear-database
// SUPER ADMIN ONLY
func AdminClearDatabase(c *gin.Context) {
	// defense-in-depth
	roleAny, ok := c.Get("role")
	if !ok || roleAny.(string) != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "super admin only"})
		return
	}

	var p clearDBPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if p.Confirm != "DELETE_ALL_DATA" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid confirmation",
			"hint":  "set confirm = DELETE_ALL_DATA",
		})
		return
	}

	db := app.DB

	// DELETE instead of TRUNCATE to preserve super_admin
	queries := []string{
		"SET FOREIGN_KEY_CHECKS = 0",

		// transaksi & keuangan
		"DELETE FROM detail_transaksis",
		"DELETE FROM wallet_transactions",
		"DELETE FROM transaksis",

		// bisnis
		"DELETE FROM diskons",
		"DELETE FROM menus",

		// siswa & stan
		"DELETE FROM siswas",
		"DELETE FROM stans",

		// users: KEEP super_admin
		"DELETE FROM users WHERE role != 'super_admin'",

		"SET FOREIGN_KEY_CHECKS = 1",
	}

	for _, q := range queries {
		if err := db.Exec(q).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "failed to clear database",
				"detail": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "DATABASE CLEARED (super admin preserved)",
		"warning": "this action is irreversible",
	})
}
