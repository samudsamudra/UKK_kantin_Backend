package siswa

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

//
// =========================
// HISTORI TRANSAKSI SISWA
// =========================
//

// GET /api/siswa/orders
// Histori transaksi siswa berdasarkan akun login (JWT)
func SiswaOrdersByMonth(c *gin.Context) {
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 🔑 Ambil siswa langsung dari DB (jangan percaya preload context)
	var siswa app.Siswa
	if err := app.DB.
		Where("user_id = ?", user.ID).
		First(&siswa).Error; err != nil {

		c.JSON(http.StatusForbidden, gin.H{"error": "user is not siswa"})
		return
	}

	var trxs []app.Transaksi
	err := app.DB.
		Preload("Details.Menu").
		Where("siswa_id = ?", siswa.ID).
		Order("created_at DESC").
		Find(&trxs).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	out := make([]gin.H, 0, len(trxs))

	for _, t := range trxs {
		var total float64
		items := make([]gin.H, 0, len(t.Details))

		for _, d := range t.Details {
			sub := float64(d.Qty) * d.HargaBeli
			total += sub

			items = append(items, gin.H{
				"menu":       d.Menu.NamaMakanan,
				"qty":        d.Qty,
				"harga_beli": app.Round2(d.HargaBeli),
				"subtotal":   app.Round2(sub),
			})
		}

		out = append(out, gin.H{
			"transaksi_id": t.PublicID,
			"tanggal":      t.CreatedAt,
			"tanggal_real": app.FormatTimeHuman(t.CreatedAt),
			"status":       t.Status,
			"total":        app.Round2(total),
			"items":        items,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": out,
	})
}

//
// =========================
// CREATE ORDER (FINAL FIX)
// =========================
//

// POST /api/siswa/order
func SiswaCreateOrder(c *gin.Context) {
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var p CreateOrderPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := app.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 🔑 ambil siswa berdasarkan user_id (WAJIB)
	var siswa app.Siswa
	if err := tx.
		Where("user_id = ?", user.ID).
		First(&siswa).Error; err != nil {

		tx.Rollback()
		c.JSON(http.StatusForbidden, gin.H{"error": "user is not siswa"})
		return
	}

	var stanID uint
	var total float64
	var details []app.DetailTransaksi

	for _, it := range p.Items {
		var menu app.Menu
		if err := tx.
			Where("public_id = ?", it.MenuID).
			First(&menu).Error; err != nil {

			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "menu not found"})
			return
		}

		if menu.StanID == nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "menu has no stan"})
			return
		}

		if stanID == 0 {
			stanID = *menu.StanID
		} else if *menu.StanID != stanID {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "mixed stans not allowed"})
			return
		}

		// =========================
		// 🔥 HITUNG DISKON SAAT ORDER
		// =========================

		var diskons []app.Diskon
		if err := tx.
			Model(&app.Diskon{}).
			Joins("JOIN menu_diskons md ON md.diskon_id = diskons.id").
			Where(`
				md.menu_id = ?
				AND (diskons.tanggal_awal IS NULL OR diskons.tanggal_awal <= CURRENT_TIMESTAMP())
				AND (diskons.tanggal_akhir IS NULL OR diskons.tanggal_akhir >= CURRENT_TIMESTAMP())
			`, menu.ID).
			Find(&diskons).Error; err != nil {

			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch discount"})
			return
		}

		latest := pickLatestApplicableDiscount(diskons)

		hargaBeli := menu.Harga
		if latest != nil {
			hargaBeli = menu.Harga * (1 - latest.PersentaseDiskon/100)
		}

		hargaBeli = app.Round2(hargaBeli)
		sub := float64(it.Qty) * hargaBeli
		total += sub

		details = append(details, app.DetailTransaksi{
			MenuID:    menu.ID,
			Qty:       it.Qty,
			HargaBeli: hargaBeli, // ✅ harga SETELAH diskon
			CreatedAt: time.Now(),
		})
	}

	// 🔑 transaksi TERIKAT ke siswa
	trx := app.Transaksi{
		PublicID:  uuid.NewString(),
		StanID:    stanID,
		SiswaID:   siswa.ID,
		Status:    app.StatusBelumDikonfirm,
		CreatedAt: time.Now(),
	}

	if err := tx.Create(&trx).Error; err != nil {
		tx.Rollback()
		log.Printf("[SiswaCreateOrder] failed create transaksi: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaksi"})
		return
	}

	for i := range details {
		details[i].TransaksiID = trx.ID
		if err := tx.Create(&details[i]).Error; err != nil {
			tx.Rollback()
			log.Printf("[SiswaCreateOrder] failed create detail: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaksi detail"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"transaksi_id": trx.PublicID,
		"status":       trx.Status,
		"total":        app.Round2(total),
	})
}
