package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

type createMenuPayload struct {
	NamaMakanan string  `json:"nama_makanan" binding:"required,min=1"`
	Harga       float64 `json:"harga" binding:"required,gt=0"`
	Jenis       string  `json:"jenis" binding:"required,oneof=makanan minuman"`
	Deskripsi   string  `json:"deskripsi,omitempty"`
}

type updateMenuPayload struct {
	NamaMakanan *string  `json:"nama_makanan,omitempty"`
	Harga       *float64 `json:"harga,omitempty"`
	Jenis       *string  `json:"jenis,omitempty"`
	Deskripsi   *string  `json:"deskripsi,omitempty"`
}

func AdminCreateMenu(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	var p createMenuPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	menu := app.Menu{
		NamaMakanan: p.NamaMakanan,
		Harga:       p.Harga,
		Jenis:       app.MenuJenis(p.Jenis),
		Deskripsi:   p.Deskripsi,
		StanID:      &stan.ID,
	}

	if err := app.DB.Create(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create menu"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "menu created",
		"menu_id": menu.PublicID,
		"nama":    menu.NamaMakanan,
		"harga":   menu.Harga,
		"jenis":   menu.Jenis,
		"stan_id": stan.PublicID,
	})
}

func AdminListMenus(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	var menus []app.Menu
	if err := app.DB.Where("stan_id = ?", stan.ID).Find(&menus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch menus"})
		return
	}

	out := make([]gin.H, 0, len(menus))
	now := time.Now().UTC()
	for _, m := range menus {
		discountsForMenu, _ := fetchDiscountsForMenu(m.ID)
		latest := pickLatestApplicableDiscount(discountsForMenu, now)

		var persentase float64
		var diskonID *string
		var tAwal *time.Time
		var tAkhir *time.Time
		if latest != nil {
			persentase = latest.PersentaseDiskon
			diskonID = &latest.PublicID
			tAwal = latest.TanggalAwal
			tAkhir = latest.TanggalAkhir
		} else {
			persentase = 0
		}

		hargaAkhir := m.Harga * (1 - persentase/100.0)

		out = append(out, gin.H{
			"menu_id":           m.PublicID,
			"nama_makanan":      m.NamaMakanan,
			"harga_asli":        m.Harga,
			"persentase_diskon": persentase,
			"harga_akhir":       hargaAkhir,
			"jenis":             m.Jenis,
			"deskripsi":         m.Deskripsi,
			"diskon_id":         diskonID,
			"tanggal_awal":      tAwal,
			"tanggal_akhir":     tAkhir,
		})
	}

	c.JSON(http.StatusOK, gin.H{"menus": out})
}

func AdminGetMenu(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	menuID := c.Param("id")
	if menuID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing menu id"})
		return
	}

	menu, err := findMenuByPublicIDAndEnsureOwnership(menuID, stan.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "menu not found or not owned"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch menu"})
		return
	}

	discountsForMenu, _ := fetchDiscountsForMenu(menu.ID)
	latest := pickLatestApplicableDiscount(discountsForMenu, time.Now().UTC())

	var persentase float64
	var diskonID *string
	var tAwal *time.Time
	var tAkhir *time.Time
	if latest != nil {
		persentase = latest.PersentaseDiskon
		diskonID = &latest.PublicID
		tAwal = latest.TanggalAwal
		tAkhir = latest.TanggalAkhir
	} else {
		persentase = 0
	}

	hargaAkhir := menu.Harga * (1 - persentase/100.0)

	c.JSON(http.StatusOK, gin.H{
		"menu_id":           menu.PublicID,
		"nama_makanan":      menu.NamaMakanan,
		"harga_asli":        menu.Harga,
		"persentase_diskon": persentase,
		"harga_akhir":       hargaAkhir,
		"jenis":             menu.Jenis,
		"deskripsi":         menu.Deskripsi,
		"diskon_id":         diskonID,
		"tanggal_awal":      tAwal,
		"tanggal_akhir":     tAkhir,
	})
}

func fetchDiscountsForMenu(menuID uint) ([]app.Diskon, error) {
	var diskons []app.Diskon
	err := app.DB.
		Joins("JOIN menu_diskons md ON md.diskon_id = diskons.id").
		Where("md.menu_id = ?", menuID).
		Preload("Menus").
		Find(&diskons).Error
	if err != nil {
		return nil, err
	}
	return diskons, nil
}

func pickLatestApplicableDiscount(diskons []app.Diskon, now time.Time) *app.Diskon {
	var best *app.Diskon
	for i := range diskons {
		d := &diskons[i]

		// check date window: must be within start/end if those exist
		if d.TanggalAwal != nil && now.Before(*d.TanggalAwal) {
			continue
		}
		if d.TanggalAkhir != nil && now.After(*d.TanggalAkhir) {
			continue
		}

		if best == nil {
			best = d
			continue
		}

		// prefer newer CreatedAt
		if d.CreatedAt.After(best.CreatedAt) {
			best = d
			continue
		}
		// if same CreatedAt, prefer larger percentage
		if d.CreatedAt.Equal(best.CreatedAt) && d.PersentaseDiskon > best.PersentaseDiskon {
			best = d
		}
	}
	return best
}

func findMenuByPublicIDAndEnsureOwnership(menuPubID string, stanID uint) (*app.Menu, error) {
	var menu app.Menu
	if err := app.DB.Where("public_id = ?", menuPubID).First(&menu).Error; err != nil {
		return nil, err
	}
	if menu.StanID == nil || *menu.StanID != stanID {
		return nil, gorm.ErrRecordNotFound
	}
	return &menu, nil
}

func AdminUpdateMenu(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	menuID := c.Param("id")
	if menuID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing menu id"})
		return
	}

	menu, err := findMenuByPublicIDAndEnsureOwnership(menuID, stan.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu not found or not owned"})
		return
	}

	var p updateMenuPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if p.NamaMakanan != nil {
		menu.NamaMakanan = *p.NamaMakanan
	}
	if p.Harga != nil {
		menu.Harga = *p.Harga
	}
	if p.Jenis != nil {
		menu.Jenis = app.MenuJenis(*p.Jenis)
	}
	if p.Deskripsi != nil {
		menu.Deskripsi = *p.Deskripsi
	}

	if err := app.DB.Save(menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "menu updated",
		"menu_id": menu.PublicID,
	})
}

func AdminDeleteMenu(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	menuID := c.Param("id")
	if menuID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing menu id"})
		return
	}

	menu, err := findMenuByPublicIDAndEnsureOwnership(menuID, stan.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu not found or not owned"})
		return
	}

	// clear any associations in join table to be safe
	_ = app.DB.Model(&menu).Association("Diskons").Clear()

	if err := app.DB.Delete(&app.Menu{}, menu.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "menu deleted"})
}
