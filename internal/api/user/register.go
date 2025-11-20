package user

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// payload for register
type registerPayload struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=6"`
	// optional for stan/admin flows (ignored for regular siswa)
	Role      string `json:"role,omitempty"`       // "siswa" by default
	NamaSiswa string `json:"nama_siswa,omitempty"` // optional friendly name for siswa
	Telp      string `json:"telp,omitempty"`
	Alamat    string `json:"alamat,omitempty"`
}

// RegisterUser -> POST /api/auth/register
// Creates user record and a Siswa profile (if role == siswa or not provided).
func RegisterUser(c *gin.Context) {
	var p registerPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// normalize role
	role := app.RoleSiswa
	if p.Role != "" {
		role = app.UserRole(p.Role)
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[REGISTER] bcrypt err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// create in transaction: user + optional siswa
	tx := app.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	u := app.User{
		PublicID: uuid.NewString(),
		Username: p.Username,
		Password: string(hashed),
		Role:     role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.Create(&u).Error; err != nil {
		tx.Rollback()
		// uniqueness error likely
		c.JSON(http.StatusBadRequest, gin.H{"error": "username already taken or invalid"})
		return
	}

	// if role is siswa (default), create a Siswa profile attached to this user
	if role == app.RoleSiswa {
		s := app.Siswa{
			PublicID: uuid.NewString(),
			Nama:     p.NamaSiswa,
			Alamat:   p.Alamat,
			Telp:     p.Telp,
			UserID:   &u.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if s.Nama == "" {
			s.Nama = p.Username
		}
		if err := tx.Create(&s).Error; err != nil {
			// log but rollback and return error to client
			tx.Rollback()
			log.Printf("[REGISTER] failed to create siswa profile: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create siswa profile"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "user created",
		"user_id":    u.PublicID,
		"username":   u.Username,
		"role":       u.Role,
	})
}
