package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

//
// =========================
// HELPER STATUS LABEL (ADMIN)
// =========================
//

func adminStatusLabel(s app.TransaksiStatus) string {
	switch s {
	case app.StatusBelumDikonfirm:
		return "Menunggu konfirmasi"
	case app.StatusDimasak:
		return "Sedang dimasak"
	case app.StatusDiantar:
		return "Sedang diantar"
	case app.StatusSampai:
		return "Pesanan sudah sampai"
	default:
		return "Status tidak diketahui"
	}
}

//
// =========================
// LIST ORDERS (ADMIN STAN)
// =========================
// GET /api/admin/orders
//

func AdminOrders(c *gin.Context) {
	uidv, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := uidv.(uint)

	// ambil stan milik admin
	var stan app.Stan
	if err := app.DB.Where("user_id = ?", userID).First(&stan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stan not found"})
		return
	}

	var trxs []app.Transaksi
	if err := app.DB.
		Where("stan_id = ?", stan.ID).
		Preload("Details.Menu").
		Order("created_at DESC").
		Find(&trxs).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	out := make([]gin.H, 0, len(trxs))
	for _, t := range trxs {
		items := make([]gin.H, 0, len(t.Details))
		var total float64

		for _, d := range t.Details {
			sub := float64(d.Qty) * d.HargaBeli
			total += sub

			items = append(items, gin.H{
				"menu_id":      d.Menu.PublicID,
				"nama_makanan": d.Menu.NamaMakanan,
				"qty":          d.Qty,
				"harga_beli":   app.Round2(d.HargaBeli),
				"subtotal":     app.Round2(sub),
			})
		}

		out = append(out, gin.H{
			"transaksi_id": t.PublicID,

			// STATUS
			"status":       t.Status,
			"status_label": adminStatusLabel(t.Status),

			// WAKTU (UX)
			"created_at":        t.CreatedAt,
			"created_at_human": app.FormatTimeWithClock(t.CreatedAt),
			"updated_at_human": app.FormatTimeWithClock(t.UpdatedAt),

			// DATA
			"total": app.Round2(total),
			"items": items,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"stan_id": stan.PublicID,
		"orders":  out,
	})
}

//
// =========================
// UPDATE STATUS ORDER (ADMIN)
// =========================
// PATCH /api/admin/orders/:id/status
//

func AdminUpdateOrderStatus(c *gin.Context) {
	trxPub := c.Param("id")
	if trxPub == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing transaksi id"})
		return
	}

	uidv, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := uidv.(uint)

	// cek stan admin
	var stan app.Stan
	if err := app.DB.Where("user_id = ?", userID).First(&stan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stan not found"})
		return
	}

	var payload struct {
		Status app.TransaksiStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var trx app.Transaksi
	if err := app.DB.
		Where("public_id = ? AND stan_id = ?", trxPub, stan.ID).
		First(&trx).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "transaksi not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	// idempotent
	if trx.Status == payload.Status {
		c.JSON(http.StatusOK, gin.H{
			"message": "already in target status",
			"status":  trx.Status,
		})
		return
	}

	// valid transitions
	allowed := map[app.TransaksiStatus][]app.TransaksiStatus{
		app.StatusBelumDikonfirm: {app.StatusDimasak},
		app.StatusDimasak:        {app.StatusDiantar},
		app.StatusDiantar:        {app.StatusSampai},
	}

	cur := trx.Status
	target := payload.Status

	valid := false
	if nexts, ok := allowed[cur]; ok {
		for _, n := range nexts {
			if n == target {
				valid = true
				break
			}
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid status transition",
			"from":  cur,
			"to":    target,
		})
		return
	}

	res := app.DB.Model(&app.Transaksi{}).
		Where("id = ? AND status = ?", trx.ID, trx.Status).
		Updates(map[string]interface{}{
			"status":     target,
			"updated_at": time.Now(),
		})

	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "status already changed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "status updated",
		"transaksi_id": trxPub,
		"new_status":   target,
	})
}
