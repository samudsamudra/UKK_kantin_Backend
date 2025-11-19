package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// payloads
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

// AdminCreateMenu -> POST /api/admin/menus
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

// AdminListMenus -> GET /api/admin/menus
// returns menus owned by this admin's stan
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
	for _, m := range menus {
		out = append(out, gin.H{
			"menu_id":      m.PublicID,
			"nama_makanan": m.NamaMakanan,
			"harga":        m.Harga,
			"jenis":        m.Jenis,
			"deskripsi":    m.Deskripsi,
		})
	}

	c.JSON(http.StatusOK, gin.H{"menus": out})
}

// helper: find menu by public id and ensure ownership
func findMenuByPublicIDAndEnsureOwnership(menuPubID string, stanID uint) (*app.Menu, error) {
	var menu app.Menu
	if err := app.DB.Where("public_id = ?", menuPubID).First(&menu).Error; err != nil {
		return nil, err
	}
	if menu.StanID == nil || *menu.StanID != stanID {
		return nil, app.DB.Where("1 = 0").Error // force "not found" behavior
	}
	return &menu, nil
}

// AdminUpdateMenu -> PUT /api/admin/menus/:id
// :id = menu public id (UUID)
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

// AdminDeleteMenu -> DELETE /api/admin/menus/:id
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

	if err := app.DB.Delete(&app.Menu{}, menu.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "menu deleted"})
}
