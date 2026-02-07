package siswa

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

//
// =========================
// LIST MENUS (SISWA)
// =========================
//

// GET /api/siswa/menus
// optional: ?stan_id=<stan_public_id>
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

		if err := db.
			Where("stan_id = ?", stan.ID).
			Find(&menus).Error; err != nil {

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
		stanID := getStanPublicIDByID(m.StanID)
		stanName := getStanNameByID(m.StanID)

		price := app.Round2(m.Harga)
		priceFinal := price

		var diskonInfo interface{} = nil
		if diskon := app.GetActiveDiscountByStan(m.StanID); diskon != nil {
			priceFinal = app.ApplyDiscount(price, diskon.Persentase)
			diskonInfo = gin.H{
				"nama":       diskon.Nama,
				"persentase": diskon.Persentase,
			}
		}

		out = append(out, gin.H{
			"id":          m.PublicID,
			"name":        m.NamaMakanan,
			"description": m.Deskripsi,
			"type":        m.Jenis,

			"price":       price,
			"price_final": priceFinal,
			"diskon":      diskonInfo,

			"stan": gin.H{
				"id":   stanID,
				"name": stanName,
			},

			"created_at":       m.CreatedAt,
			"created_at_human": app.FormatTimeWithClock(m.CreatedAt),
			"updated_at":       m.UpdatedAt,
			"updated_at_human": app.FormatTimeWithClock(m.UpdatedAt),
		})
	}

	c.JSON(http.StatusOK, gin.H{"menus": out})
}

//
// =========================
// GET MENU DETAIL (SISWA)
// =========================
//

// GET /api/siswa/menus/:id
func SiswaGetMenu(c *gin.Context) {
	pub := c.Param("id")
	if pub == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing menu id"})
		return
	}

	var m app.Menu
	if err := app.DB.
		Where("public_id = ?", pub).
		First(&m).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "menu not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch menu"})
		return
	}

	stanID := getStanPublicIDByID(m.StanID)
	stanName := getStanNameByID(m.StanID)

	// 🔑 diskon aktif
	// diskon := app.GetActiveDiscount()

	// 🔑 diskon aktif per-stan
	diskon := app.GetActiveDiscountByStan(m.StanID)

	price := app.Round2(m.Harga)
	priceFinal := price

	var diskonInfo interface{} = nil
	if diskon != nil {
		priceFinal = app.ApplyDiscount(price, diskon.Persentase)
		diskonInfo = gin.H{
			"nama":       diskon.Nama,
			"persentase": diskon.Persentase,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          m.PublicID,
		"name":        m.NamaMakanan,
		"description": m.Deskripsi,
		"type":        m.Jenis,

		"price":       price,
		"price_final": priceFinal,
		"diskon":      diskonInfo,

		"stan": gin.H{
			"id":   stanID,
			"name": stanName,
		},

		"created_at":       m.CreatedAt,
		"created_at_human": app.FormatTimeWithClock(m.CreatedAt),
		"updated_at":       m.UpdatedAt,
		"updated_at_human": app.FormatTimeWithClock(m.UpdatedAt),
	})
}
