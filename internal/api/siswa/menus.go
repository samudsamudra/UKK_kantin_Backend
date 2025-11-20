package siswa

import (
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// SiswaListMenus - improved debug version
// returns menus + real-time best discount (based on DB CURRENT_TIMESTAMP())
func SiswaListMenus(c *gin.Context) {
	stanPub := c.Query("stan_id")

	db := app.DB

	// fetch menus (optionally filter by stan)
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

	// For each menu, query active discounts (use CURRENT_TIMESTAMP() for DB-time consistency)
	for _, m := range menus {
		var diskons []app.Diskon
		if err := db.
			Model(&app.Diskon{}).
			Joins("JOIN menu_diskons md ON md.diskon_id = diskons.id").
			Where("md.menu_id = ? AND (diskons.tanggal_awal IS NULL OR diskons.tanggal_awal <= CURRENT_TIMESTAMP()) AND (diskons.tanggal_akhir IS NULL OR diskons.tanggal_akhir >= CURRENT_TIMESTAMP())",
				m.ID).
			Find(&diskons).Error; err != nil {
			log.Printf("[MENU DEBUG] error querying discounts for menu %s: %v", m.PublicID, err)
		}

		// debug: print what diskons found for this menu
		if len(diskons) == 0 {
			log.Printf("[MENU DEBUG] menu %s (%s) -> no active discounts", m.PublicID, m.NamaMakanan)
		} else {
			for _, d := range diskons {
				log.Printf("[MENU DEBUG] menu %s -> diskon found: id=%s pct=%.2f start=%v end=%v",
					m.PublicID, d.PublicID, d.PersentaseDiskon, d.TanggalAwal, d.TanggalAkhir)
			}
		}

		// pick best percent
		best := 0.0
		var bestDiskon app.Diskon
		for _, d := range diskons {
			if d.PersentaseDiskon > best {
				best = d.PersentaseDiskon
				bestDiskon = d
			}
		}

		var hargaDiskon float64 = 0
		if best > 0 {
			hargaDiskon = math.Round(m.Harga*(1.0-best/100.0)*100) / 100
		}

		var diskonInfo interface{}
		if best > 0 {
			var tAwal, tAkhir *string
			if bestDiskon.TanggalAwal != nil {
				s := bestDiskon.TanggalAwal.UTC().Format(time.RFC3339)
				tAwal = &s
			}
			if bestDiskon.TanggalAkhir != nil {
				s := bestDiskon.TanggalAkhir.UTC().Format(time.RFC3339)
				tAkhir = &s
			}
			diskonInfo = gin.H{
				"diskon_id":         bestDiskon.PublicID,
				"nama_diskon":       bestDiskon.NamaDiskon,
				"persentase_diskon": bestDiskon.PersentaseDiskon,
				"tanggal_awal":      tAwal,
				"tanggal_akhir":     tAkhir,
			}
		} else {
			diskonInfo = nil
		}

		out = append(out, gin.H{
			"menu_id":      m.PublicID,
			"nama_makanan": m.NamaMakanan,
			"harga":        m.Harga,
			"jenis":        m.Jenis,
			"deskripsi":    m.Deskripsi,
			"stan_id": func() string {
				if m.StanID == nil {
					return ""
				}
				return getStanPublicIDByID(*m.StanID)
			}(),
			"diskon":       diskonInfo,
			"harga_diskon": hargaDiskon,
		})
	}

	// also log summary - helpful to detect handler actually executed
	log.Printf("[MENU DEBUG] returned %d menus at %s", len(out), time.Now().UTC().Format(time.RFC3339))
	c.JSON(http.StatusOK, gin.H{"menus": out})
}

// getStanPublicIDByID returns the public_id for a stan by its numeric ID, or empty string on error/not found.
func getStanPublicIDByID(id uint) string {
	var stan app.Stan
	if err := app.DB.Select("public_id").Where("id = ?", id).First(&stan).Error; err != nil {
		log.Printf("[MENU DEBUG] failed to lookup stan id=%d: %v", id, err)
		return ""
	}
	return stan.PublicID
}
