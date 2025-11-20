package admin

import (
	"log"
	"net/http"
	"time"
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// pilih diskon terbaru yang berlaku sekarang
func latestApplicableDiskon(diskons []app.Diskon) *app.Diskon {
	var best *app.Diskon
	now := time.Now().UTC()
	for i := range diskons {
		d := &diskons[i]
		if d.TanggalAwal != nil && now.Before(*d.TanggalAwal) {
			continue
		}
		if d.TanggalAkhir != nil && now.After(*d.TanggalAkhir) {
			continue
		}
		if best == nil || d.CreatedAt.After(best.CreatedAt) || (d.CreatedAt.Equal(best.CreatedAt) && d.PersentaseDiskon > best.PersentaseDiskon) {
			best = d
		}
	}
	return best
}

// AdminOrdersByMonth -> GET /api/admin/orders
// Returns orders for the admin's stan (preloaded details and menus)
func AdminOrdersByMonth(c *gin.Context) {
	uidv, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, _ := uidv.(uint)

	var stan app.Stan
	if err := app.DB.Where("user_id = ?", userID).First(&stan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stan not found"})
		return
	}

	var trxs []app.Transaksi
	if err := app.DB.
		Where("stan_id = ?", stan.ID).
		Preload("Details.Menu.Diskons").
		Preload("Details.Menu").
		Order("created_at DESC").
		Find(&trxs).Error; err != nil {
		log.Printf("[AdminOrdersByMonth] db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	out := make([]gin.H, 0, len(trxs))
	for _, t := range trxs {
		details := make([]gin.H, 0, len(t.Details))
		var total float64

		for _, d := range t.Details {
			var pct float64
			var diskonInfo gin.H

			if len(d.Menu.Diskons) > 0 {
				latest := latestApplicableDiskon(d.Menu.Diskons)
				if latest != nil {
					pct = latest.PersentaseDiskon
					var tAwal, tAkhir *string
					if latest.TanggalAwal != nil {
						s := latest.TanggalAwal.UTC().Format(time.RFC3339)
						tAwal = &s
					}
					if latest.TanggalAkhir != nil {
						s := latest.TanggalAkhir.UTC().Format(time.RFC3339)
						tAkhir = &s
					}
					diskonInfo = gin.H{
						"diskon_id":           latest.PublicID,
						"nama_diskon":         latest.NamaDiskon,
						"persentase_diskon":   latest.PersentaseDiskon,
						"tanggal_awal":        tAwal,
						"tanggal_awal_real":   func() string { if latest.TanggalAwal == nil { return "" }; return app.FormatTimeHuman(*latest.TanggalAwal) }(),
						"tanggal_akhir":       tAkhir,
						"tanggal_akhir_real":  func() string { if latest.TanggalAkhir == nil { return "" }; return app.FormatTimeHuman(*latest.TanggalAkhir) }(),
					}
				}
			}

			hargaAkhir := app.Round2(d.Menu.Harga * (1 - pct/100.0))
			total += d.HargaBeli

			details = append(details, gin.H{
				"qty":              d.Qty,
				"harga_beli":       app.Round2(d.HargaBeli),
				"created_at":       d.CreatedAt,
				"created_at_real":  app.FormatTimeHuman(d.CreatedAt),
				"menu": gin.H{
					"menu_id":           d.Menu.PublicID,
					"nama_makanan":      d.Menu.NamaMakanan,
					"jenis":             d.Menu.Jenis,
					"deskripsi":         d.Menu.Deskripsi,
					"harga_asli":        app.Round2(d.Menu.Harga),
					"persentase_diskon": pct,
					"harga_akhir":       hargaAkhir,
					"diskon":            diskonInfo,
				},
			})
		}

		out = append(out, gin.H{
			"transaksi_id":     t.PublicID,
			"tanggal":          t.Tanggal,
			"tanggal_real":     app.FormatTimeHuman(t.Tanggal),
			"status":           t.Status,
			"CreatedAt":        t.CreatedAt,
			"created_at_real":  app.FormatTimeHuman(t.CreatedAt),
			"UpdatedAt":        t.UpdatedAt,
			"updated_at_real":  app.FormatTimeHuman(t.UpdatedAt),
			"total":            app.Round2(total),
			"Details":          details,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"stan":   stan.PublicID,
		"orders": out,
	})
}

// AdminUpdateOrderStatus -> PATCH /api/admin/orders/:id/status
// Accepts a target status and performs an atomic conditional update to
// avoid duplicate processing. Returns 200 if already in desired state.
// PATCH /api/admin/orders/:id/status
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

    // cek stan admin ini
    var stan app.Stan
    if err := app.DB.Where("user_id = ?", userID).First(&stan).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "stan not found"})
        return
    }

    // ambil payload status baru
    var payload struct {
        Status app.TransaksiStatus `json:"status" binding:"required"`
    }
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // load transaksi
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

    // idempotent success
    if trx.Status == payload.Status {
        c.JSON(http.StatusOK, gin.H{"message": "already in target status", "status": trx.Status})
        return
    }

    // validasi transition
    allowed := map[app.TransaksiStatus][]app.TransaksiStatus{
        app.StatusBelumDikonfirm: {app.StatusDimasak, "cancelled"},
        app.StatusDimasak:        {app.StatusDiantar, "cancelled"},
        app.StatusDiantar:        {app.StatusSampai},
    }

    cur := trx.Status
    target := payload.Status

    isAllowed := false
    if nexts, ok := allowed[cur]; ok {
        for _, n := range nexts {
            if n == target {
                isAllowed = true
                break
            }
        }
    }
    if !isAllowed {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "invalid status transition",
            "from":  cur,
            "to":    target,
        })
        return
    }

    // atomic update
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
        c.JSON(http.StatusConflict, gin.H{"error": "status already changed by another actor"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message":      "status updated",
        "transaksi_id": trxPub,
        "new_status":   target,
    })
}
