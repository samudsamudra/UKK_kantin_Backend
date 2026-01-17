package user

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

//
// =========================
// Payload
// =========================
//

type registerPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nama     string `json:"nama_lengkap" binding:"required"`
}

//
// =========================
// REGISTER SISWA
// =========================
//

// RegisterUser -> POST /api/auth/register
// Public endpoint: ONLY for siswa
func RegisterUser(c *gin.Context) {
	var p registerPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[REGISTER] bcrypt error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	tx := app.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// create user (role fixed: siswa)
	u := app.User{
		Email:              p.Email,
		PasswordHash:       string(hashed),
		Role:               app.RoleSiswa,
		MustChangePassword: false,
	}

	if err := tx.Create(&u).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	// create siswa profile
	s := app.Siswa{
		Nama:   p.Nama,
		UserID: u.ID,
	}

	if err := tx.Create(&s).Error; err != nil {
		tx.Rollback()
		log.Printf("[REGISTER] failed create siswa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create siswa profile"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "register success",
		"user_id": u.PublicID,
		"email":   u.Email,
		"role":    u.Role,
	})
}
