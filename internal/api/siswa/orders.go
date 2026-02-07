package siswa

import (
	// "log"
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

func SiswaOrdersByMonth(c *gin.Context) {
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{"orders": out})
}

// =========================
// CREATE ORDER (FINAL CLEAN)
// =========================
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

	var siswa app.Siswa
	if err := tx.
		Where("user_id = ?", user.ID).
		First(&siswa).Error; err != nil {

		tx.Rollback()
		c.JSON(http.StatusForbidden, gin.H{"error": "user is not siswa"})
		return
	}


	var diskon *app.Diskon

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

		// ❌ campur stan tidak boleh
		if stanID == 0 {
			stanID = menu.StanID
			diskon = app.GetActiveDiscountByStan(stanID)
		} else if menu.StanID != stanID {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "mixed stans not allowed"})
			return
		}

		// 💰 harga final (apply diskon DI SINI)
		harga := app.Round2(menu.Harga)
		if diskon != nil {
			harga = app.ApplyDiscount(harga, diskon.Persentase)
		}

		sub := float64(it.Qty) * harga
		total += sub

		details = append(details, app.DetailTransaksi{
			MenuID:    menu.ID,
			Qty:       it.Qty,
			HargaBeli: harga, // 🔥 harga sudah diskon
			CreatedAt: time.Now(),
		})
	}

	trx := app.Transaksi{
		PublicID: uuid.NewString(),
		StanID:   stanID,
		SiswaID:  siswa.ID,
		Status:   app.StatusBelumDikonfirm,
	}

	if err := tx.Create(&trx).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaksi"})
		return
	}

	for i := range details {
		details[i].TransaksiID = trx.ID
		if err := tx.Create(&details[i]).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create detail"})
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
