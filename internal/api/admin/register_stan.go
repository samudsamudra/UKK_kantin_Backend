package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/samudsamudra/UKK_kantin/internal/app"
	"golang.org/x/crypto/bcrypt"
)

type RegisterStanPayload struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	NamaStan    string `json:"nama_stan" binding:"required"`
	NamaPemilik string `json:"nama_pemilik" binding:"required"`
	Telp        string `json:"telp"`
}

func RegisterStan(c *gin.Context) {
	var p RegisterStanPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var ex app.User
	if err := app.DB.Where("username = ?", p.Username).First(&ex).Error; err == nil {
		c.JSON(409, gin.H{"error": "username already exists"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)

	user := app.User{
		Username: p.Username,
		Password: string(hash),
		Role:     app.RoleAdminStan,
	}
	app.DB.Create(&user)

	stan := app.Stan{
		NamaStan:    p.NamaStan,
		NamaPemilik: p.NamaPemilik,
		Telp:        p.Telp,
		UserID:      &user.ID,
	}
	app.DB.Create(&stan)

	c.JSON(201, gin.H{
		"message": "register stan success",
		"user_id": user.PublicID,
		"stan_id": stan.PublicID,
	})
}
