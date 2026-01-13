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

// POST /api/siswa/order
// NOTE: ensure this is the only SiswaCreateOrder in package siswa (remove/replace duplicates)
// func SiswaCreateOrder(c *gin.Context) {
// 	// debug useful: show ctx keys if things go wrong
// 	if v := c.Keys; v != nil {
// 		log.Printf("[AUTH DEBUG] context keys present")
// 	}

// 	user, ok := getUserFromContext(c)
// 	if !ok {
// 		log.Printf("[SiswaCreateOrder] getUserFromContext returned nil")
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 		return
// 	}

// 	var p CreateOrderPayload
// 	if err := c.ShouldBindJSON(&p); err != nil {
// 		log.Printf("[SiswaCreateOrder] bind error: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// idempotency key: prefer header, fallback to body field
// 	idemp := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
// 	if idemp == "" && p.IdempotencyKey != nil {
// 		idemp = strings.TrimSpace(*p.IdempotencyKey)
// 	}
// 	if idemp != "" {
// 		var ik app.IdempotencyKey
// 		if err := app.DB.Where("`key` = ? AND user_id = ?", idemp, user.ID).First(&ik).Error; err == nil {
// 			// case A: idempotency key already points to an existing transaksi -> return OK with that transaksi
// 			if ik.TransaksiID != nil {
// 				var trx app.Transaksi
// 				if err := app.DB.Preload("Details").Where("id = ?", *ik.TransaksiID).First(&trx).Error; err == nil {
// 					c.JSON(http.StatusOK, gin.H{"message": "duplicate request", "transaksi_id": trx.PublicID})
// 					return
// 				}
// 				// fallback: cannot resolve transaksi -> continue to in-progress handling
// 			}

// 			// case B: key exists but transaksi_id nil -> request IN-PROGRESS
// 			// if the key is very old, consider it stale -> delete and continue as new request
// 			if time.Since(ik.CreatedAt) > 5*time.Minute {
// 				_ = app.DB.Where("`key` = ? AND user_id = ?", idemp, user.ID).Delete(&app.IdempotencyKey{}).Error
// 			} else {
// 				// instead of 409, return 202 Accepted to be friendlier (request is being processed)
// 				c.JSON(http.StatusAccepted, gin.H{"message": "request in progress"})
// 				return
// 			}
// 		}

// 		ikNew := app.IdempotencyKey{
// 			Key:       idemp,
// 			UserID:    user.ID,
// 			CreatedAt: time.Now(),
// 		}
// 		// ignore duplicate-key error here; next request will handle it
// 		if err := app.DB.Create(&ikNew).Error; err != nil {
// 			log.Printf("[SiswaCreateOrder] failed to create idempotency key: %v (key=%s user=%d)", err, idemp, user.ID)
// 		}
// 	}

// 	tx := app.DB.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			log.Printf("[SiswaCreateOrder] panic recovered: %v", r)
// 			if tx != nil {
// 				_ = tx.Rollback().Error
// 			}
// 		}
// 	}()

// 	var stanID uint = 0
// 	var total float64 = 0
// 	var details []app.DetailTransaksi

// 	for _, it := range p.Items {
// 		var menu app.Menu
// 		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("public_id = ?", it.MenuID).First(&menu).Error; err != nil {
// 			// differentiate not found vs other DB errors
// 			if err == gorm.ErrRecordNotFound {
// 				log.Printf("[SiswaCreateOrder] menu not found: %s", it.MenuID)
// 				tx.Rollback()
// 				c.JSON(http.StatusBadRequest, gin.H{"error": "menu not found: " + it.MenuID})
// 				return
// 			}
// 			log.Printf("[SiswaCreateOrder] db error fetching menu %s: %v", it.MenuID, err)
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error fetching menu"})
// 			return
// 		}

// 		if menu.StanID == nil {
// 			log.Printf("[SiswaCreateOrder] menu has no stan: menu_public=%s id=%d", menu.PublicID, menu.ID)
// 			tx.Rollback()
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "menu has no stan"})
// 			return
// 		}
// 		if stanID == 0 {
// 			stanID = *menu.StanID
// 		} else if *menu.StanID != stanID {
// 			log.Printf("[SiswaCreateOrder] mixed stans in order: expected %d got %d (menu=%s)", stanID, *menu.StanID, menu.PublicID)
// 			tx.Rollback()
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "mixed stans in order not allowed"})
// 			return
// 		}

// 		var diskons []app.Diskon
// 		if err := tx.
// 			Model(&app.Diskon{}).
// 			Joins("JOIN menu_diskons md ON md.diskon_id = diskons.id").
// 			Where("md.menu_id = ? AND (diskons.tanggal_awal IS NULL OR diskons.tanggal_awal <= CURRENT_TIMESTAMP()) AND (diskons.tanggal_akhir IS NULL OR diskons.tanggal_akhir >= CURRENT_TIMESTAMP())",
// 				menu.ID).
// 			Find(&diskons).Error; err != nil {
// 			// ignore discount fetch error but log it for debugging
// 			log.Printf("[SiswaCreateOrder] warning: failed to fetch discounts for menu %d: %v", menu.ID, err)
// 		}
// 		latest := pickLatestApplicableDiscount(diskons)
// 		var pct float64
// 		if latest != nil {
// 			pct = latest.PersentaseDiskon
// 		} else {
// 			pct = 0
// 		}

// 		hargaBeli := app.Round2(menu.Harga * (1 - pct/100.0))
// 		if hargaBeli < 0 {
// 			hargaBeli = 0
// 		}

// 		lineTotal := float64(it.Qty) * hargaBeli
// 		total += lineTotal

// 		detail := app.DetailTransaksi{
// 			MenuID:    menu.ID,
// 			Qty:       it.Qty,
// 			HargaBeli: hargaBeli,
// 			CreatedAt: time.Now(),
// 		}
// 		details = append(details, detail)
// 	}

// 	trx := app.Transaksi{
// 		PublicID:  uuid.NewString(),
// 		Tanggal:   time.Now().UTC(),
// 		StanID:    stanID,
// 		Status:    app.StatusBelumDikonfirm,
// 		CreatedAt: time.Now(),
// 	}

// 	var u app.User
// 	// NOTE: Preload must match exported field name "Siswa"
// 	if err := tx.Preload("Siswa").Where("id = ?", user.ID).First(&u).Error; err != nil {
// 		log.Printf("[SiswaCreateOrder] failed to resolve user id=%d: %v", user.ID, err)
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve user"})
// 		return
// 	}
// 	if u.Siswa == nil {
// 		log.Printf("[SiswaCreateOrder] user is not a siswa: user_id=%d", user.ID)
// 		tx.Rollback()
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "user is not a siswa"})
// 		return
// 	}
// 	trx.SiswaID = u.Siswa.ID

// 	if p.PaymentMethod == "wallet" {
// 		var cur app.User
// 		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", user.ID).First(&cur).Error; err != nil {
// 			log.Printf("[SiswaCreateOrder] failed to lock user balance id=%d: %v", user.ID, err)
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lock user balance"})
// 			return
// 		}
// 		// log balance + total for reproducibility
// 		log.Printf("[SiswaCreateOrder] user_id=%d current_saldo=%.2f total=%.2f", user.ID, cur.Saldo, total)
// 		if cur.Saldo < total {
// 			tx.Rollback()
// 			c.JSON(http.StatusPaymentRequired, gin.H{"error": "saldo tidak cukup"})
// 			return
// 		}

// 		trx.Status = app.StatusDimasak
// 		if err := tx.Create(&trx).Error; err != nil {
// 			log.Printf("[SiswaCreateOrder] failed to create transaksi: %v trx=%+v", err, trx)
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaksi"})
// 			return
// 		}

// 		for i := range details {
// 			details[i].TransaksiID = trx.ID
// 			if err := tx.Create(&details[i]).Error; err != nil {
// 				log.Printf("[SiswaCreateOrder] failed to create transaksi detail: %v detail=%+v", err, details[i])
// 				tx.Rollback()
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaksi detail"})
// 				return
// 			}
// 		}

// 		if err := tx.Model(&app.User{}).Where("id = ?", user.ID).UpdateColumn("saldo", gorm.Expr("saldo - ?", total)).Error; err != nil {
// 			log.Printf("[SiswaCreateOrder] failed to debit saldo for user %d: %v", user.ID, err)
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to debit saldo"})
// 			return
// 		}

// 		wtx := app.WalletTransaction{
// 			PublicID:  uuid.NewString(),
// 			UserID:    user.ID,
// 			Amount:    -total,
// 			Type:      "debit",
// 			Note:      "Pembayaran pesanan " + trx.PublicID,
// 			CreatedAt: time.Now(),
// 		}
// 		if err := tx.Create(&wtx).Error; err != nil {
// 			log.Printf("[SiswaCreateOrder] failed to create wallet tx: %v wtx=%+v", err, wtx)
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create wallet tx"})
// 			return
// 		}

// 	} else {
// 		// cash
// 		trx.Status = app.StatusBelumDikonfirm
// 		if err := tx.Create(&trx).Error; err != nil {
// 			log.Printf("[SiswaCreateOrder] failed to create transaksi (cash): %v trx=%+v", err, trx)
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaksi"})
// 			return
// 		}
// 		for i := range details {
// 			details[i].TransaksiID = trx.ID
// 			if err := tx.Create(&details[i]).Error; err != nil {
// 				log.Printf("[SiswaCreateOrder] failed to create transaksi detail (cash): %v detail=%+v", err, details[i])
// 				tx.Rollback()
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaksi detail"})
// 				return
// 			}
// 		}
// 	}

// 	if idemp != "" {
// 		if err := tx.Model(&app.IdempotencyKey{}).Where("`key` = ? AND user_id = ?", idemp, user.ID).Updates(map[string]interface{}{
// 			"transaksi_id": trx.ID,
// 			"response":     "",
// 		}).Error; err != nil {
// 			log.Printf("[SiswaCreateOrder] warning: failed to update idempotency key (key=%s user=%d): %v", idemp, user.ID, err)
// 			// not fatal
// 		}
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		log.Printf("[SiswaCreateOrder] commit failed: %v", err)
// 		_ = tx.Rollback().Error
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
// 		return
// 	}

// 	log.Printf("[SiswaCreateOrder] success transaksi_public_id=%s user=%d total=%.2f", trx.PublicID, user.ID, app.Round2(total))
// 	c.JSON(http.StatusCreated, gin.H{
// 		"transaksi_id": trx.PublicID,
// 		"status":       trx.Status,
// 		"total":        app.Round2(total),
// 	})
// }

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
