package seed

import (
	"log"

	"github.com/samudsamudra/UKK_kantin/internal/app"
	"golang.org/x/crypto/bcrypt"
)

func SeedSiswas() {
	db := app.DB

	type siswaSeed struct {
		Nama  string
		Email string
	}

	siswas := []siswaSeed{
		{"Andi Pratama", "andi@school.local"},
		{"Budi Hartono", "budi@school.local"},
		{"Citra Lestari", "citra@school.local"},
	}

	for _, s := range siswas {
		var exist app.User
		if err := db.Where("email = ?", s.Email).First(&exist).Error; err == nil {
			continue
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

		user := app.User{
			Email:        s.Email,
			PasswordHash: string(hash),
			Role:         app.RoleSiswa,
			Saldo:        50000, // 🔥 biar bisa langsung order
		}
		if err := db.Create(&user).Error; err != nil {
			log.Println("[SEED] failed create siswa user")
			continue
		}

		siswa := app.Siswa{
			Nama:   s.Nama,
			UserID: user.ID,
		}
		_ = db.Create(&siswa).Error

		log.Println("[SEED] siswa created:", s.Nama)
	}
}
