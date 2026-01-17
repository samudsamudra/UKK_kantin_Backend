package siswa

import (
	// "log"
	"net/http"
	// "strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	// "gorm.io/gorm/clause"

	"github.com/google/uuid"
	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// Order payload
type OrderItemPayload struct {
	MenuID string `json:"menu_id" binding:"required"`
	Qty    int    `json:"qty" binding:"required,gt=0"`
}
type CreateOrderPayload struct {
	Items         []OrderItemPayload `json:"items" binding:"required,min=1"`
	PaymentMethod string             `json:"payment_method" binding:"required,oneof=wallet cash"`
	// optional: client can send idempotency key header instead
	IdempotencyKey *string `json:"idempotency_key,omitempty"`
}

// GET /api/siswa/wallet - get current user's saldo
func SiswaGetWallet(c *gin.Context) {
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var u app.User
	if err := app.DB.Select("saldo").Where("id = ?", user.ID).First(&u).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch saldo"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"saldo": u.Saldo})
}

// POST /api/siswa/topup - admin/operator topup (for quick testing or manual topup)
func SiswaTopupByAdmin(c *gin.Context) {
	var payload struct {
		UserPublicID string  `json:"user_public_id" binding:"required"`
		Amount       float64 `json:"amount" binding:"required,gt=0"`
		Note         string  `json:"note,omitempty"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user app.User
	if err := app.DB.Where("public_id = ?", payload.UserPublicID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	tx := app.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&app.User{}).Where("id = ?", user.ID).UpdateColumn("saldo", gorm.Expr("saldo + ?", payload.Amount)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to topup"})
		return
	}

	wtx := app.WalletTransaction{
		PublicID:  uuid.NewString(),
		UserID:    user.ID,
		Amount:    payload.Amount,
		Type:      "topup",
		Note:      payload.Note,
		CreatedAt: time.Now(),
	}
	if err := tx.Create(&wtx).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record wallet tx"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "topup successful"})
}

// robust getUserFromContext: accepts either *app.User, numeric user id (uint/float64) or public_id string.
func getUserFromContext(c *gin.Context) (*app.User, bool) {
	// 1) direct *app.User stored (future-proof)
	if v, ok := c.Get("user"); ok {
		if u, ok2 := v.(*app.User); ok2 {
			return u, true
		}
	}

	// 2) middleware sets numeric user_id (uint) under "user_id"
	if v, ok := c.Get("user_id"); ok {
		switch idv := v.(type) {
		case uint:
			var u app.User
			if err := app.DB.Preload("Siswa").Where("id = ?", idv).First(&u).Error; err == nil {
				return &u, true
			}
		case int:
			var u app.User
			if err := app.DB.Preload("Siswa").Where("id = ?", idv).First(&u).Error; err == nil {
				return &u, true
			}
		case float64: // sometimes JSON decodes numbers as float64
			var u app.User
			if err := app.DB.Preload("Siswa").Where("id = ?", uint(idv)).First(&u).Error; err == nil {
				return &u, true
			}
		}
	}

	// 3) middleware sets public_id under "public_id"
	if v, ok := c.Get("public_id"); ok {
		if pid, ok2 := v.(string); ok2 && pid != "" {
			var u app.User
			if err := app.DB.Preload("Siswa").Where("public_id = ?", pid).First(&u).Error; err == nil {
				return &u, true
			}
		}
	}

	// 4) middleware might set claims under "claims" or "jwt_claims"
	if v, ok := c.Get("claims"); ok {
		if m, ok2 := v.(map[string]interface{}); ok2 {
			if idf, ok := m["user_id"]; ok {
				switch idvv := idf.(type) {
				case float64:
					var u app.User
					if err := app.DB.Preload("Siswa").Where("id = ?", uint(idvv)).First(&u).Error; err == nil {
						return &u, true
					}
				case string:
					var u app.User
					if err := app.DB.Preload("Siswa").Where("public_id = ?", idvv).First(&u).Error; err == nil {
						return &u, true
					}
				}
			}
		}
	}

	return nil, false
}
