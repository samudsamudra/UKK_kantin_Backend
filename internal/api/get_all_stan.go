package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samudsamudra/UKK_kantin/internal/app"
)

func AdminGetAllStan(c *gin.Context) {
	var stans []app.Stan

	if err := app.DB.Find(&stans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	type StanResponse struct {
		StanID      string `json:"stan_id"`
		NamaStan    string `json:"nama_stan"`
		NamaPemilik string `json:"nama_pemilik"`
		Telp        string `json:"telp"`
		UserID      uint   `json:"user_id"`
	}

	var resp []StanResponse

	for _, s := range stans {
		resp = append(resp, StanResponse{
			StanID:      s.PublicID,
			NamaStan:    s.NamaStan,
			NamaPemilik: s.NamaPemilik,
			Telp:        s.Telp,
			UserID:      s.UserID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"stans": resp,
	})
}
