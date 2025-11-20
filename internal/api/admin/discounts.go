package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samudsamudra/UKK_kantin/internal/app"
	
)

// payloads
type createDiscountPayload struct {
	NamaDiskon       string   `json:"nama_diskon" binding:"required"`
	PersentaseDiskon float64  `json:"persentase_diskon" binding:"required,gte=0,lte=100"`
	TanggalAwalStr   *string  `json:"tanggal_awal,omitempty"` // RFC3339
	TanggalAkhirStr  *string  `json:"tanggal_akhir,omitempty"` // RFC3339
  MenuPublicIDs    []string `json:"menu_public_ids,omitempty"`
}

type updateDiscountPayload struct {
	NamaDiskon       *string   `json:"nama_diskon,omitempty"`
	PersentaseDiskon *float64  `json:"persentase_diskon,omitempty"`
	TanggalAwalStr   *string   `json:"tanggal_awal,omitempty"`
	TanggalAkhirStr  *string   `json:"tanggal_akhir,omitempty"`
	MenuPublicIDs    *[]string `json:"menu_public_ids,omitempty"` // replace assignment if present
}

// helper: parse optional RFC3339 to *time.Time
func parseOptionalTime(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil, err
	}
	tt := t.UTC()
	return &tt, nil
}

// ----------------------
// CREATE
// ----------------------

// AdminCreateDiscount -> POST /api/admin/discounts
// Admin can create a discount and assign menus (must own menus)
func AdminCreateDiscount(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	var p createDiscountPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tAwal, err := parseOptionalTime(p.TanggalAwalStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tanggal_awal, use RFC3339"})
		return
	}
	tAkhir, err := parseOptionalTime(p.TanggalAkhirStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tanggal_akhir, use RFC3339"})
		return
	}
	if tAwal != nil && tAkhir != nil && tAwal.After(*tAkhir) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tanggal_awal must be before tanggal_akhir"})
		return
	}

	// start transaction
	tx := app.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	d := app.Diskon{
		PublicID:         uuid.NewString(),
		NamaDiskon:       p.NamaDiskon,
		PersentaseDiskon: p.PersentaseDiskon,
		TanggalAwal:      tAwal,
		TanggalAkhir:     tAkhir,
	}

	if err := tx.Create(&d).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create discount"})
		return
	}

	// if menus assigned, validate ownership and attach
	if len(p.MenuPublicIDs) > 0 {
		var menus []app.Menu
		for _, pub := range p.MenuPublicIDs {
			var m app.Menu
			if err := tx.Where("public_id = ?", pub).First(&m).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "menu not found: " + pub})
				return
			}
			if m.StanID == nil || *m.StanID != stan.ID {
				tx.Rollback()
				c.JSON(http.StatusForbidden, gin.H{"error": "cannot assign menu not owned: " + pub})
				return
			}
			menus = append(menus, m)
		}
		// associate (many2many)
		if err := tx.Model(&d).Association("Menus").Append(&menus); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign menus to discount"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "discount created",
		"diskon_id":     d.PublicID,
		"nama_diskon":   d.NamaDiskon,
		"persentase":    d.PersentaseDiskon,
		"tanggal_awal":  d.TanggalAwal,
		"tanggal_akhir": d.TanggalAkhir,
	})
}

// ----------------------
// LIST
// ----------------------

// AdminListDiscounts -> GET /api/admin/discounts
// Lists discounts that are assigned to menus owned by this admin's stan
func AdminListDiscounts(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	var diskons []app.Diskon
	// join through menu_diskons -> menus -> filter by stan_id
	if err := app.DB.
		Preload("Menus").
		Joins("JOIN menu_diskons md ON md.diskon_id = diskons.id").
		Joins("JOIN menus m ON m.id = md.menu_id").
		Where("m.stan_id = ?", stan.ID).
		Group("diskons.id").
		Find(&diskons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list discounts"})
		return
	}

	out := make([]gin.H, 0, len(diskons))
	for _, d := range diskons {
		menuPub := []string{}
		for _, m := range d.Menus {
			menuPub = append(menuPub, m.PublicID)
		}
		out = append(out, gin.H{
			"diskon_id":         d.PublicID,
			"nama_diskon":       d.NamaDiskon,
			"persentase_diskon": d.PersentaseDiskon,
			"tanggal_awal":      d.TanggalAwal,
			"tanggal_akhir":     d.TanggalAkhir,
			"menus":             menuPub,
		})
	}

	c.JSON(http.StatusOK, gin.H{"discounts": out})
}

// ----------------------
// GET
// ----------------------

// AdminGetDiscount -> GET /api/admin/discounts/:id
func AdminGetDiscount(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}
	pub := c.Param("id")
	if pub == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	var d app.Diskon
	if err := app.DB.Preload("Menus").
		Where("public_id = ?", pub).
		First(&d).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "discount not found"})
		return
	}

	// ensure at least one menu belongs to this stan (otherwise admin shouldn't access)
	found := false
	for _, m := range d.Menus {
		if m.StanID != nil && *m.StanID == stan.ID {
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "discount not found"})
		return
	}

	menuPub := []string{}
	for _, m := range d.Menus {
		menuPub = append(menuPub, m.PublicID)
	}

	c.JSON(http.StatusOK, gin.H{
		"diskon_id":         d.PublicID,
		"nama_diskon":       d.NamaDiskon,
		"persentase_diskon": d.PersentaseDiskon,
		"tanggal_awal":      d.TanggalAwal,
		"tanggal_akhir":     d.TanggalAkhir,
		"menus":             menuPub,
	})
}

// ----------------------
// UPDATE (partial) - PATCH semantics
// ----------------------

// AdminUpdateDiscount -> PATCH /api/admin/discounts/:id
// Accepts partial fields (name/percent/date/menu list replace)
func AdminUpdateDiscount(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}
	pub := c.Param("id")
	if pub == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	var d app.Diskon
	if err := app.DB.Preload("Menus").Where("public_id = ?", pub).First(&d).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "discount not found"})
		return
	}

	// ensure admin has at least one menu in this discount (ownership)
	owned := false
	for _, m := range d.Menus {
		if m.StanID != nil && *m.StanID == stan.ID {
			owned = true
			break
		}
	}
	if !owned {
		c.JSON(http.StatusNotFound, gin.H{"error": "discount not found"})
		return
	}

	var p updateDiscountPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// start tx
	tx := app.DB.Begin()

	if p.NamaDiskon != nil {
		d.NamaDiskon = *p.NamaDiskon
	}
	if p.PersentaseDiskon != nil {
		if *p.PersentaseDiskon < 0 || *p.PersentaseDiskon > 100 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "persentase_diskon out of range"})
			return
		}
		d.PersentaseDiskon = *p.PersentaseDiskon
	}
	if p.TanggalAwalStr != nil {
		t, err := parseOptionalTime(p.TanggalAwalStr)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tanggal_awal"})
			return
		}
		d.TanggalAwal = t
	}
	if p.TanggalAkhirStr != nil {
		t, err := parseOptionalTime(p.TanggalAkhirStr)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tanggal_akhir"})
			return
		}
		d.TanggalAkhir = t
	}
	// if menu_public_ids provided -> replace assignment (validate ownership)
	if p.MenuPublicIDs != nil {
		// clear existing association first
		if err := tx.Model(&d).Association("Menus").Clear(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear menu assignments"})
			return
		}
		// validate and append
		var menus []app.Menu
		for _, pub := range *p.MenuPublicIDs {
			var m app.Menu
			if err := tx.Where("public_id = ?", pub).First(&m).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "menu not found: " + pub})
				return
			}
			if m.StanID == nil || *m.StanID != stan.ID {
				tx.Rollback()
				c.JSON(http.StatusForbidden, gin.H{"error": "cannot assign menu not owned: " + pub})
				return
			}
			menus = append(menus, m)
		}
		if err := tx.Model(&d).Association("Menus").Append(&menus); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign menus"})
			return
		}
	}

	if err := tx.Save(&d).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update discount"})
		return
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "discount updated", "diskon_id": d.PublicID})
}

// ----------------------
// PATCH only menus list (replace) - /api/admin/discounts/:id/menus
// ----------------------

func AdminUpdateDiscountMenus(c *gin.Context) {
	pub := c.Param("id")
	if pub == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}

	var d app.Diskon
	if err := app.DB.Preload("Menus").Where("public_id = ?", pub).First(&d).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "discount not found"})
		return
	}

	// ensure this discount touches at least one menu of this stan (ownership)
	owned := false
	for _, m := range d.Menus {
		if m.StanID != nil && *m.StanID == stan.ID {
			owned = true
			break
		}
	}
	if !owned {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed: discount does not belong to your stan"})
		return
	}

	var payload struct {
		MenuPublicIDs []string `json:"menu_public_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := app.DB.Begin()
	// clear existing associations
	if err := tx.Model(&d).Association("Menus").Clear(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear existing menu associations"})
		return
	}

	// validate & append only menus that belong to this stan
	var toAssign []app.Menu
	for _, pubID := range payload.MenuPublicIDs {
		var m app.Menu
		if err := tx.Where("public_id = ? AND stan_id = ?", pubID, stan.ID).First(&m).Error; err != nil {
			// skip menu not found or not owned; do not abort entire op for UX
			continue
		}
		toAssign = append(toAssign, m)
	}

	if len(toAssign) > 0 {
		if err := tx.Model(&d).Association("Menus").Append(&toAssign); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign menus to discount"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "discount menus updated",
		"discount_id":    d.PublicID,
		"menus_assigned": payload.MenuPublicIDs,
	})
}

// ----------------------
// DELETE
// ----------------------

// AdminDeleteDiscount -> DELETE /api/admin/discounts/:id
func AdminDeleteDiscount(c *gin.Context) {
	stan, ok := requireStanOrAbort(c)
	if !ok {
		return
	}
	pub := c.Param("id")
	if pub == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	var d app.Diskon
	if err := app.DB.Preload("Menus").Where("public_id = ?", pub).First(&d).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "discount not found"})
		return
	}

	// ensure admin owns at least one menu in this discount
	owned := false
	for _, m := range d.Menus {
		if m.StanID != nil && *m.StanID == stan.ID {
			owned = true
			break
		}
	}
	if !owned {
		c.JSON(http.StatusNotFound, gin.H{"error": "discount not found"})
		return
	}

	tx := app.DB.Begin()
	if err := tx.Model(&d).Association("Menus").Clear(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear associations"})
		return
	}
	if err := tx.Delete(&d).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete discount"})
		return
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "discount deleted"})
}
