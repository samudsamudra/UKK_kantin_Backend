package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// GET /api/admin/reports/rekap
// Rekap transaksi yang SUDAH SAMPAI (urut lama → terbaru)
func AdminRekapTransaksi(c *gin.Context) {
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 🔑 ambil stan berdasarkan user_id (JWT)
	var stan app.Stan
	if err := app.DB.
		Where("user_id = ?", user.ID).
		First(&stan).Error; err != nil {

		c.JSON(http.StatusForbidden, gin.H{"error": "user is not admin stan"})
		return
	}

	// =========================
	// Ambil transaksi (HANYA yang sudah sampai)
	// =========================
	var transaksis []app.Transaksi
	err := app.DB.
		Preload("Details").
		Where(
			"stan_id = ? AND status = ?",
			stan.ID,
			app.StatusSampai,
		).
		Order("created_at ASC"). // TERLAMA → TERBARU
		Find(&transaksis).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}

	// =========================
	// Hitung rekap
	// =========================
	var totalPemasukan float64
	out := make([]gin.H, 0, len(transaksis))

	for _, trx := range transaksis {
		var totalTrx float64
		for _, d := range trx.Details {
			totalTrx += float64(d.Qty) * d.HargaBeli
		}

		totalPemasukan += totalTrx

		out = append(out, gin.H{
			"transaksi_id": trx.PublicID,
			"tanggal":      trx.CreatedAt, // raw timestamp (aman)
			"tanggal_real": trx.CreatedAt.Format("02 Jan 2006 15:04"),
			"total":        app.Round2(totalTrx),
		})
	}

	// =========================
	// Response rapi
	// =========================
	c.JSON(http.StatusOK, gin.H{
		"total_transaksi": len(transaksis),
		"total_pemasukan": app.Round2(totalPemasukan),
		"orders":          out, // urut lama → terbaru
	})
}
