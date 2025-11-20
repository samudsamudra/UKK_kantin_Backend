package siswa

import (
	"log"
	"math"
	"net/http"
	"time"

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
		var diskons []app.Diskon
		if err := db.
			Model(&app.Diskon{}).
			Joins("JOIN menu_diskons md ON md.diskon_id = diskons.id").
			Where("md.menu_id = ? AND (diskons.tanggal_awal IS NULL OR diskons.tanggal_awal <= CURRENT_TIMESTAMP()) AND (diskons.tanggal_akhir IS NULL OR diskons.tanggal_akhir >= CURRENT_TIMESTAMP())",
				m.ID).
			Find(&diskons).Error; err != nil {
			log.Printf("[SiswaListMenus] error querying discounts for menu %s: %v", m.PublicID, err)
		}

		latest := pickLatestApplicableDiscount(diskons)

		var pct float64
		var diskonInfo gin.H
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
		} else {
			pct = 0
			diskonInfo = nil
		}

		hargaAkhir := round2(m.Harga * (1 - pct/100.0))

		out = append(out, gin.H{
			"menu_id":            m.PublicID,
			"nama_makanan":       m.NamaMakanan,
			"harga_asli":         round2(m.Harga),
			"persentase_diskon":  pct,
			"harga_akhir":        hargaAkhir,
			"jenis":              m.Jenis,
			"deskripsi":          m.Deskripsi,
			"stan_id":            func() string { if m.StanID == nil { return "" }; return getStanPublicIDByID(*m.StanID) }(),
			"stan_name":          func() string { if m.StanID == nil { return "" }; return getStanNameByID(*m.StanID) }(),
			"diskon":             diskonInfo,
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

	var diskons []app.Diskon
	if err := app.DB.
		Model(&app.Diskon{}).
		Joins("JOIN menu_diskons md ON md.diskon_id = diskons.id").
		Where("md.menu_id = ? AND (diskons.tanggal_awal IS NULL OR diskons.tanggal_awal <= CURRENT_TIMESTAMP()) AND (diskons.tanggal_akhir IS NULL OR diskons.tanggal_akhir >= CURRENT_TIMESTAMP())",
			m.ID).
		Find(&diskons).Error; err != nil {
		log.Printf("[SiswaGetMenu] error querying discounts for menu %s: %v", m.PublicID, err)
	}

	latest := pickLatestApplicableDiscount(diskons)

	var pct float64
	var diskonInfo gin.H
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
	} else {
		pct = 0
		diskonInfo = nil
	}

	hargaAkhir := round2(m.Harga * (1 - pct/100.0))

	c.JSON(http.StatusOK, gin.H{
		"menu_id":           m.PublicID,
		"nama_makanan":      m.NamaMakanan,
		"harga_asli":        round2(m.Harga),
		"persentase_diskon": pct,
		"harga_akhir":       hargaAkhir,
		"jenis":             m.Jenis,
		"deskripsi":         m.Deskripsi,
		"stan_id":           func() string { if m.StanID == nil { return "" }; return getStanPublicIDByID(*m.StanID) }(),
		"stan_name":         func() string { if m.StanID == nil { return "" }; return getStanNameByID(*m.StanID) }(),
		"diskon":            diskonInfo,
	})
}

func pickLatestApplicableDiscount(diskons []app.Diskon) *app.Diskon {
	var best *app.Diskon
	for i := range diskons {
		d := &diskons[i]
		// defensive: skip discounts outside date window (should already be filtered by DB)
		now := time.Now().UTC()
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

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

func getStanPublicIDByID(id uint) string {
	var pub string
	app.DB.Table("stans").Select("public_id").Where("id = ?", id).Scan(&pub)
	return pub
}

func getStanNameByID(id uint) string {
	var name string
	app.DB.Table("stans").Select("nama_stan").Where("id = ?", id).Scan(&name)
	return name
}
