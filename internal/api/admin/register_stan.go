package admin

import (
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

type registerStanPayload struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	NamaStan    string `json:"nama_stan" binding:"required"`
	NamaPemilik string `json:"nama_pemilik" binding:"required"`
	Telp        string `json:"telp,omitempty"`
}

//
// =========================
// Handler
// =========================
//

// RegisterStan -> POST /api/admin/stan/register
// Only super_admin can create admin_stan
func RegisterStan(c *gin.Context) {
	// pastikan role super_admin
	role, ok := c.Get("role")
	if !ok || role != string(app.RoleSuperAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var p registerStanPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// cek email unik
	var ex app.User
	if err := app.DB.Where("email = ?", p.Email).First(&ex).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// create user admin stan
	user := app.User{
		Email:              p.Email,
		PasswordHash:       string(hash),
		Role:               app.RoleAdminStan,
		MustChangePassword: true,
	}

	if err := app.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// create stan
	stan := app.Stan{
		NamaStan:    p.NamaStan,
		NamaPemilik: p.NamaPemilik,
		Telp:        p.Telp,
		UserID:      user.ID,
	}

	if err := app.DB.Create(&stan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create stan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "register stan success",
		"user_id":  user.PublicID,
		"stan_id":  stan.PublicID,
		"email":    user.Email,
		"must_change_password": user.MustChangePassword,
	})
}
