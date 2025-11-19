package api

import (
	"github.com/gin-gonic/gin"
	"github.com/samudsamudra/UKK_kantin/internal/app"
	"golang.org/x/crypto/bcrypt"
)

type UserRegisterPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nama     string `json:"nama"`
	Alamat   string `json:"alamat"`
	Telp     string `json:"telp"`
	Foto     string `json:"foto"`
}

func RegisterUser(c *gin.Context) {
	var p UserRegisterPayload
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
		Role:     app.RoleSiswa,
	}
	app.DB.Create(&user)

	siswa := app.Siswa{
		Nama:   p.Nama,
		Alamat: p.Alamat,
		Telp:   p.Telp,
		Foto:   p.Foto,
		UserID: &user.ID,
	}
	app.DB.Create(&siswa)

	c.JSON(201, gin.H{
		"message":  "register siswa success",
		"user_id":  user.PublicID,
		"siswa_id": siswa.PublicID,
		"nama":     siswa.Nama,
		"username": user.Username,
	})
}
