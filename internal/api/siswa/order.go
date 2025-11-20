package siswa

import (
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// payloads
type orderItemPayload struct {
	MenuID string `json:"menu_id" binding:"required"` // public id
	Qty    int    `json:"qty" binding:"required,gt=0"`
}

type createOrderPayload struct {
	StanID string             `json:"stan_id" binding:"required"` // public id
	Items  []orderItemPayload `json:"items" binding:"required,min=1"`
}

// SiswaCreateOrder -> POST /api/siswa/order
// Applies active discounts automatically when creating the order.
func SiswaCreateOrder(c *gin.Context) {
	var p createOrderPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get current user (siswa) from context (set by JWTAuth)
	uidVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	userID, ok := uidVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in context"})
		return
	}

	// find siswa profile for this user
	var siswa app.Siswa
	if err := app.DB.Where("user_id = ?", userID).First(&siswa).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "siswa profile not found"})
		return
	}

	// lookup stan by public id
	var stan app.Stan
	if err := app.DB.Where("public_id = ?", p.StanID).First(&stan).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stan not found"})
		return
	}

	// begin transaction
	tx := app.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// create transaksi
	tr := app.Transaksi{
		PublicID: uuid.NewString(),
		StanID:   stan.ID,
		SiswaID:  siswa.ID,
		Status:   app.StatusBelumDikonfirm,
	}
	if err := tx.Create(&tr).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaksi"})
		return
	}

	now := time.Now().UTC()

	for _, it := range p.Items {
		// load menu by public id
		var menu app.Menu
		if err := tx.Where("public_id = ?", it.MenuID).First(&menu).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "menu not found: " + it.MenuID})
			return
		}
		// ensure menu belongs to chosen stan
		if menu.StanID == nil || *menu.StanID != stan.ID {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "menu does not belong to chosen stan: " + it.MenuID})
			return
		}

		// find active discounts for this menu (choose best percent)
		var diskons []app.Diskon
		if err := tx.
			Joins("JOIN menu_diskons md ON md.diskon_id = diskons.id").
			Where("md.menu_id = ? AND (diskons.tanggal_awal IS NULL OR diskons.tanggal_awal <= ?) AND (diskons.tanggal_akhir IS NULL OR diskons.tanggal_akhir >= ?)",
				menu.ID, now, now).
			Find(&diskons).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query discounts"})
			return
		}

		best := 0.0
		for _, d := range diskons {
			if d.PersentaseDiskon > best {
				best = d.PersentaseDiskon
			}
		}

		hargaAfter := menu.Harga
		if best > 0 {
			hargaAfter = math.Round(menu.Harga*(1.0-best/100.0)*100) / 100
		}

		dt := app.DetailTransaksi{
			TransaksiID: tr.ID,
			MenuID:      menu.ID,
			Qty:         it.Qty,
			HargaBeli:   hargaAfter,
		}
		if err := tx.Create(&dt).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create detail transaksi"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "order created", "transaksi_id": tr.PublicID})
}

// SiswaOrdersByMonth -> GET /api/siswa/orders?month=2025-11
// Returns transaksi list for the current siswa filtered by month (YYYY-MM)
func SiswaOrdersByMonth(c *gin.Context) {
	uidVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	userID, ok := uidVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in context"})
		return
	}

	var siswa app.Siswa
	if err := app.DB.Where("user_id = ?", userID).First(&siswa).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "siswa profile not found"})
		return
	}

	monthStr := c.Query("month")
	var start, end time.Time
	if monthStr == "" {
		now := time.Now().UTC()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)
	} else {
		t, err := time.Parse("2006-01", monthStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "month must be YYYY-MM"})
			return
		}
		start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)
	}

	var trans []app.Transaksi
	if err := app.DB.Where("siswa_id = ? AND tanggal >= ? AND tanggal < ?", siswa.ID, start, end).
		Order("tanggal desc").Find(&trans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}

	out := make([]gin.H, 0, len(trans))
	for _, t := range trans {
		out = append(out, gin.H{
			"transaksi_id": t.PublicID,
			"tanggal":      t.Tanggal,
			"status":       t.Status,
		})
	}

	c.JSON(http.StatusOK, gin.H{"transactions": out})
}

// SiswaGetReceiptPDF -> GET /api/siswa/order/:id/receipt
// For now returns transaction detail as JSON (you can later wire PDF generator)
func SiswaGetReceiptPDF(c *gin.Context) {
	uidVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	userID, ok := uidVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in context"})
		return
	}

	pub := c.Param("id")
	if pub == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	// find siswa
	var siswa app.Siswa
	if err := app.DB.Where("user_id = ?", userID).First(&siswa).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "siswa profile not found"})
		return
	}

	var tr app.Transaksi
	if err := app.DB.Preload("Details").Preload("Details.Menu").Where("public_id = ? AND siswa_id = ?", pub, siswa.ID).First(&tr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	// build detail
	items := make([]gin.H, 0, len(tr.Details))
	var total float64
	for _, d := range tr.Details {
		items = append(items, gin.H{
			"menu_id":   d.Menu.PublicID,
			"nama":      d.Menu.NamaMakanan,
			"qty":       d.Qty,
			"harga_beli": d.HargaBeli,
			"subtotal":  float64(d.Qty) * d.HargaBeli,
		})
		total += float64(d.Qty) * d.HargaBeli
	}

	c.JSON(http.StatusOK, gin.H{
		"transaksi_id": tr.PublicID,
		"tanggal":      tr.Tanggal,
		"status":       tr.Status,
		"items":        items,
		"total":        total,
	})
}
