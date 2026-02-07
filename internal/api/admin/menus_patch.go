package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samudsamudra/UKK_kantin/internal/app"
)

func AdminPatchMenu(c *gin.Context) {
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

	changed := false

	if p.NamaMakanan != nil {
		menu.NamaMakanan = *p.NamaMakanan
		changed = true
	}
	if p.Harga != nil {
		menu.Harga = *p.Harga
		changed = true
	}
	if p.Jenis != nil {
		menu.Jenis = app.MenuJenis(*p.Jenis)
		changed = true
	}
	if p.Deskripsi != nil {
		menu.Deskripsi = *p.Deskripsi
		changed = true
	}

	if !changed {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no fields to update",
		})
		return
	}

	if err := app.DB.Save(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update menu",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "menu updated",
		"menu_id": menu.PublicID,
	})
}