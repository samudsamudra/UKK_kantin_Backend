package admin

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// helper: pilih diskon terbaru yang berlaku sekarang (copy logic kecil)
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
// returns all transactions for the stan owned by the authenticated admin,
// with price & discount info computed similarly to SiswaListMenus.
func AdminOrdersByMonth(c *gin.Context) {
	uidv, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, _ := uidv.(uint)

	var stan app.Stan
	if err := app.DB.Where("user_id = ?", userID).First(&stan).Error; err != nil {
		log.Printf("[AdminOrdersByMonth] stan not found for user %d: %v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "stan not found"})
		return
	}

	var trxs []app.Transaksi
	// PRELOAD Diskons for each menu so we can compute applicable discount
	if err := app.DB.
		Where("stan_id = ?", stan.ID).
		Order("created_at DESC").
		Preload("Details.Menu.Diskons").
		Preload("Details.Menu").
		Find(&trxs).Error; err != nil {
		log.Printf("[AdminOrdersByMonth] failed fetch trxs stan=%d: %v", stan.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	outOrders := make([]gin.H, 0, len(trxs))
	for _, t := range trxs {
		detailsOut := make([]gin.H, 0, len(t.Details))
		var total float64 = 0
		for _, d := range t.Details {
			// compute current applicable discount from preloaded diskons (if any)
			var diskonInfo gin.H = nil
			var pct float64 = 0
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
						"diskon_id":         latest.PublicID,
						"nama_diskon":       latest.NamaDiskon,
						"persentase_diskon": latest.PersentaseDiskon,
						"tanggal_awal":      tAwal,
						"tanggal_akhir":     tAkhir,
					}
				}
			}

			hargaAkhir := app.Round2(d.Menu.Harga * (1 - pct/100.0))
			// total use recorded HargaBeli to reflect historical price, but show harga_akhir for reference
			total += d.HargaBeli

			detailsOut = append(detailsOut, gin.H{
				"qty":         d.Qty,
				"harga_beli":  d.HargaBeli, // stored at order time
				"menu": gin.H{
					"menu_id":           d.Menu.PublicID,
					"nama_makanan":      d.Menu.NamaMakanan,
					"harga_asli":        app.Round2(d.Menu.Harga),
					"persentase_diskon": pct,
					"harga_akhir":       hargaAkhir,
					"jenis":             d.Menu.Jenis,
					"deskripsi":         d.Menu.Deskripsi,
					"diskon":            diskonInfo,
				},
			})
		}

		outOrders = append(outOrders, gin.H{
			"transaksi_id": t.PublicID,
			"tanggal":      t.Tanggal,
			"status":       t.Status,
			"CreatedAt":    t.CreatedAt,
			"UpdatedAt":    t.UpdatedAt,
			"total":        app.Round2(total),
			"Details":      detailsOut,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"stan":   stan.PublicID,
		"orders": outOrders,
	})
}

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
	userID, _ := uidv.(uint)

	var stan app.Stan
	if err := app.DB.Where("user_id = ?", userID).First(&stan).Error; err != nil {
		log.Printf("[AdminUpdateOrderStatus] stan not found for user %d: %v", userID, err)
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
	if err := app.DB.Where("public_id = ? AND stan_id = ?", trxPub, stan.ID).First(&trx).Error; err != nil {
		log.Printf("[AdminUpdateOrderStatus] transaksi not found pub=%s stan=%d: %v", trxPub, stan.ID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "transaksi not found"})
		return
	}

	if err := app.DB.Model(&trx).Update("status", payload.Status).Error; err != nil {
		log.Printf("[AdminUpdateOrderStatus] failed update status trx=%s: %v", trxPub, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "status updated",
		"transaksi_id": trxPub,
		"new_status":   payload.Status,
	})
}