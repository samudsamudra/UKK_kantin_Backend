package admin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

//
// =========================
// Payloads
// =========================
//

type createDiscountPayload struct {
	Nama         string  `json:"nama_diskon" binding:"required"`
	Persentase   float64 `json:"persentase_diskon" binding:"required,gte=0,lte=100"`
	TanggalAwal  *string `json:"tanggal_awal,omitempty"`
	TanggalAkhir *string `json:"tanggal_akhir,omitempty"`
}

type updateDiscountPayload struct {
	Nama         *string  `json:"nama_diskon,omitempty"`
	Persentase   *float64 `json:"persentase_diskon,omitempty"`
	TanggalAwal  *string  `json:"tanggal_awal,omitempty"`
	TanggalAkhir *string  `json:"tanggal_akhir,omitempty"`
}

//
// =========================
// Helpers
// =========================
//

func parseOptionalTime(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	str := strings.TrimSpace(*s)
	if str == "" {
		return nil, nil
	}

	// RFC3339
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		tt := t.UTC()
		return &tt, nil
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	layouts := []string{
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, str, loc); err == nil {
			tt := t.UTC()
			return &tt, nil
		}
	}

	return nil, fmt.Errorf("unsupported date format")
}

//
// =========================
// CREATE
// =========================
//

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

	tAwal, err := parseOptionalTime(p.TanggalAwal)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tanggal_awal"})
		return
	}
	tAkhir, err := parseOptionalTime(p.TanggalAkhir)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tanggal_akhir"})
		return
	}
	if tAwal != nil && tAkhir != nil && tAwal.After(*tAkhir) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tanggal_awal must be before tanggal_akhir"})
		return
	}

	d := app.Diskon{
		PublicID:     uuid.NewString(),
		StanID:       stan.ID,
		Nama:         p.Nama,
		Persentase:   p.Persentase,
		TanggalAwal:  tAwal,
		TanggalAkhir: tAkhir,
	}

	if err := app.DB.Create(&d).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create discount"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"diskon_id":         d.PublicID,
		"stan_id":           stan.PublicID,
		"nama_diskon":       d.Nama,
		"persentase_diskon": d.Persentase,
		"tanggal_awal":      app.FormatISOOrNil(d.TanggalAwal),
		"tanggal_akhir":     app.FormatISOOrNil(d.TanggalAkhir),
	})

}

//
// =========================
// LIST
// =========================
//

func AdminListDiscounts(c *gin.Context) {
	var diskons []app.Diskon
	if err := app.DB.Order("created_at DESC").Find(&diskons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list discounts"})
		return
	}

	out := make([]gin.H, 0, len(diskons))
	for _, d := range diskons {
		out = append(out, gin.H{
			"diskon_id":         d.PublicID,
			"nama_diskon":       d.Nama,
			"persentase_diskon": d.Persentase,

			"tanggal_awal":  app.FormatISOOrNil(d.TanggalAwal),
			"tanggal_akhir": app.FormatISOOrNil(d.TanggalAkhir),

			"tanggal_awal_human":  app.FormatDateID(d.TanggalAwal, false),
			"tanggal_awal_short":  app.FormatDateID(d.TanggalAwal, true),
			"tanggal_akhir_human": app.FormatDateID(d.TanggalAkhir, false),
			"tanggal_akhir_short": app.FormatDateID(d.TanggalAkhir, true),
		})
	}

	c.JSON(http.StatusOK, gin.H{"discounts": out})
}

//
// =========================
// GET
// =========================
//

func AdminGetDiscount(c *gin.Context) {
	pub := c.Param("id")

	var d app.Diskon
	if err := app.DB.Where("public_id = ?", pub).First(&d).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "discount not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"diskon_id":         d.PublicID,
		"nama_diskon":       d.Nama,
		"persentase_diskon": d.Persentase,
		"tanggal_awal":      d.TanggalAwal,
		"tanggal_akhir":     d.TanggalAkhir,
	})
}

//
// =========================
// UPDATE
// =========================
//

func AdminUpdateDiscount(c *gin.Context) {
	pub := c.Param("id")

	var d app.Diskon
	if err := app.DB.Where("public_id = ?", pub).First(&d).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "discount not found"})
		return
	}

	var p updateDiscountPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if p.Nama != nil {
		d.Nama = *p.Nama
	}
	if p.Persentase != nil {
		d.Persentase = *p.Persentase
	}
	if p.TanggalAwal != nil {
		t, err := parseOptionalTime(p.TanggalAwal)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tanggal_awal"})
			return
		}
		d.TanggalAwal = t
	}
	if p.TanggalAkhir != nil {
		t, err := parseOptionalTime(p.TanggalAkhir)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tanggal_akhir"})
			return
		}
		d.TanggalAkhir = t
	}

	if err := app.DB.Save(&d).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update discount"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "discount updated"})
}

//
// =========================
// DELETE
// =========================
//

func AdminDeleteDiscount(c *gin.Context) {
	pub := c.Param("id")

	var d app.Diskon
	if err := app.DB.Where("public_id = ?", pub).First(&d).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "discount not found"})
		return
	}

	if err := app.DB.Delete(&d).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete discount"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "discount deleted"})
}
