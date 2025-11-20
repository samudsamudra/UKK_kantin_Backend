package siswa

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// SiswaListMenus -> GET /api/siswa/menus
// optional query: ?stan_id=<stan_public_id>
func SiswaListMenus(c *gin.Context) {
	stanPub := c.Query("stan_id")
	db := app.DB

	var menus []app.Menu
	if stanPub != "" {
		var stan app.Stan
		if err := db.Where("public_id = ?", stanPub).First(&stan).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stan not found"})
			return
		}
		if err := db.Where("stan_id = ?", stan.ID).Find(&menus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch menus"})
			return
		}
	} else {
		if err := db.Find(&menus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch menus"})
			return
		}
	}

	out := make([]gin.H, 0, len(menus))
	for _, m := range menus {
		// fetch applicable discounts for this menu
		var diskons []app.Diskon
		if err := db.
			Model(&app.Diskon{}).
			Joins("JOIN menu_diskons md ON md.diskon_id = diskons.id").
			Where("md.menu_id = ? AND (diskons.tanggal_awal IS NULL OR diskons.tanggal_awal <= CURRENT_TIMESTAMP()) AND (diskons.tanggal_akhir IS NULL OR diskons.tanggal_akhir >= CURRENT_TIMESTAMP())",
				m.ID).
			Find(&diskons).Error; err != nil {
			log.Printf("[SiswaListMenus] error querying discounts for menu %s: %v", m.PublicID, err)
		}

		// pick latest applicable discount via helper
		var discountObj gin.H = nil
		var latest *app.Diskon = nil
		if len(diskons) > 0 {
			latest = pickLatestApplicableDiscount(diskons)
			if latest != nil {
				discountObj = gin.H{
					"id":           latest.PublicID,
					"name":         latest.NamaDiskon,
					"percent":      latest.PersentaseDiskon,
					"start":        app.FormatISOOrNil(latest.TanggalAwal),
					"start_pretty": app.FormatTimePretty(latest.TanggalAwal),
					"end":          app.FormatISOOrNil(latest.TanggalAkhir),
					"end_pretty":   app.FormatTimePretty(latest.TanggalAkhir),
				}
			}
		}

		// compute final price using latest struct (if any)
		priceFinal := app.Round2(m.Harga)
		if latest != nil && latest.PersentaseDiskon > 0 {
			priceFinal = app.Round2(m.Harga * (1 - latest.PersentaseDiskon/100.0))
		}

		stanID := ""
		stanName := ""
		if m.StanID != nil {
			stanID = getStanPublicIDByID(*m.StanID)
			stanName = getStanNameByID(*m.StanID)
		}

		out = append(out, gin.H{
			"id":          m.PublicID,
			"name":        m.NamaMakanan,
			"description": m.Deskripsi,
			"type":        m.Jenis,

			"price":       app.Round2(m.Harga),
			"price_final": priceFinal,

			"discount": discountObj, // nil if none

			"stan": gin.H{
				"id":   stanID,
				"name": stanName,
			},

			"created_at":       m.CreatedAt,
			"created_at_real-time": app.FormatTimeHuman(m.CreatedAt),
			"updated_at":       m.UpdatedAt,
			"updated_at_real-time": app.FormatTimeHuman(m.UpdatedAt),
		})
	}

	c.JSON(http.StatusOK, gin.H{"menus": out})
}

// SiswaGetMenu -> GET /api/siswa/menus/:id
func SiswaGetMenu(c *gin.Context) {
	pub := c.Param("id")
	if pub == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing menu id"})
		return
	}

	var m app.Menu
	if err := app.DB.Where("public_id = ?", pub).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "menu not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch menu"})
		return
	}

	// fetch applicable discounts
	var diskons []app.Diskon
	if err := app.DB.
		Model(&app.Diskon{}).
		Joins("JOIN menu_diskons md ON md.diskon_id = diskons.id").
		Where("md.menu_id = ? AND (diskons.tanggal_awal IS NULL OR diskons.tanggal_awal <= CURRENT_TIMESTAMP()) AND (diskons.tanggal_akhir IS NULL OR diskons.tanggal_akhir >= CURRENT_TIMESTAMP())",
			m.ID).
		Find(&diskons).Error; err != nil {
		log.Printf("[SiswaGetMenu] error querying discounts for menu %s: %v", m.PublicID, err)
	}

	var discountObj gin.H = nil
	var latest *app.Diskon = nil
	if len(diskons) > 0 {
		latest = pickLatestApplicableDiscount(diskons)
		if latest != nil {
			discountObj = gin.H{
				"id":           latest.PublicID,
				"name":         latest.NamaDiskon,
				"percent":      latest.PersentaseDiskon,
				"start":        app.FormatISOOrNil(latest.TanggalAwal),
				"start_pretty": app.FormatTimePretty(latest.TanggalAwal),
				"end":          app.FormatISOOrNil(latest.TanggalAkhir),
				"end_pretty":   app.FormatTimePretty(latest.TanggalAkhir),
			}
		}
	}

	priceFinal := app.Round2(m.Harga)
	if latest != nil && latest.PersentaseDiskon > 0 {
		priceFinal = app.Round2(m.Harga * (1 - latest.PersentaseDiskon/100.0))
	}

	stanID := ""
	stanName := ""
	if m.StanID != nil {
		stanID = getStanPublicIDByID(*m.StanID)
		stanName = getStanNameByID(*m.StanID)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               m.PublicID,
		"name":             m.NamaMakanan,
		"description":      m.Deskripsi,
		"type":             m.Jenis,
		"price":            app.Round2(m.Harga),
		"price_final":      priceFinal,
		"discount":         discountObj,
		"stan":             gin.H{"id": stanID, "name": stanName},
		"created_at":       m.CreatedAt,
		"created_at_real-time": app.FormatTimeHuman(m.CreatedAt),
		"updated_at":       m.UpdatedAt,
		"updated_at_real-time": app.FormatTimeHuman(m.UpdatedAt),
	})
}
