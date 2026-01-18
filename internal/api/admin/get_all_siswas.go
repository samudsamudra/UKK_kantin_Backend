package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

type siswaResp struct {
	PublicID string `json:"public_id"`
	Nama     string `json:"nama_lengkap"`
	Email    string `json:"email"`
}

// AdminGetAllSiswas
// GET /api/admin/system/siswas
// 🔒 SUPER ADMIN ONLY
func AdminGetAllSiswas(c *gin.Context) {
	// defense-in-depth
	roleAny, ok := c.Get("role")
	if !ok || roleAny.(string) != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "super admin only"})
		return
	}

	// ambil siswa
	var siswas []app.Siswa
	if err := app.DB.Find(&siswas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch siswas",
		})
		return
	}

	resp := make([]siswaResp, 0, len(siswas))

	for _, s := range siswas {
		var user app.User
		if err := app.DB.
			Where("id = ?", s.UserID).
			First(&user).Error; err != nil {
			// skip jika user tidak ada (data rusak)
			continue
		}

		resp = append(resp, siswaResp{
			PublicID: s.PublicID,
			Nama:     s.Nama,
			Email:    user.Email,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":  len(resp),
		"siswas": resp,
	})
}
