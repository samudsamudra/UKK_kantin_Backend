package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

//
// =========================
// Payloads
// =========================
//

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

//
// =========================
// CREATE MENU (ADMIN STAN)
// =========================
//

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
		StanID:      stan.ID, // 🔑 IKAT KE STAN
		NamaMakanan: p.NamaMakanan,
		Harga:       p.Harga,
		Jenis:       app.MenuJenis(p.Jenis),
		Deskripsi:   p.Deskripsi,
	}

	if err := app.DB.Create(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create menu"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "menu created",
		"menu_id":      menu.PublicID,
		"nama_makanan": menu.NamaMakanan,
		"harga":        menu.Harga,
		"jenis":        menu.Jenis,
	})
}

//
// =========================
// LIST MENU (ADMIN STAN)
// =========================
//

func AdminListMenus(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	var menus []app.Menu
	if err := app.DB.
		Where("stan_id = ?", stan.ID).
		Order("created_at DESC").
		Find(&menus).Error; err != nil {

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

//
// =========================
// GET MENU DETAIL (ADMIN)
// =========================
//

func AdminGetMenu(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	pub := c.Param("id")

	var menu app.Menu
	if err := app.DB.
		Where("public_id = ? AND stan_id = ?", pub, stan.ID).
		First(&menu).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "menu not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"menu_id":      menu.PublicID,
		"nama_makanan": menu.NamaMakanan,
		"harga":        menu.Harga,
		"jenis":        menu.Jenis,
		"deskripsi":    menu.Deskripsi,
	})
}

//
// =========================
// UPDATE MENU (ADMIN)
// =========================
//

func AdminUpdateMenu(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	pub := c.Param("id")

	var menu app.Menu
	if err := app.DB.
		Where("public_id = ? AND stan_id = ?", pub, stan.ID).
		First(&menu).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "menu not found"})
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

	if err := app.DB.Save(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "menu updated",
		"menu_id": menu.PublicID,
	})
}

//
// =========================
// DELETE MENU (ADMIN)
// =========================
//

func AdminDeleteMenu(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	pub := c.Param("id")

	var menu app.Menu
	if err := app.DB.
		Where("public_id = ? AND stan_id = ?", pub, stan.ID).
		First(&menu).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "menu not found"})
		return
	}

	if err := app.DB.Delete(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "menu deleted"})
}
