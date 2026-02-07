package seed

import (
	"log"

	"github.com/samudsamudra/UKK_kantin/internal/app"
	"golang.org/x/crypto/bcrypt"
)

func SeedStans() {
	db := app.DB

	type stanSeed struct {
		Email       string
		Password    string
		NamaStan    string
		NamaPemilik string
		Telp        string
	}

	stans := []stanSeed{
		{
			Email:       "admin.stan1@kantin.local",
			Password:    "password123",
			NamaStan:    "Kantin Bu Siti",
			NamaPemilik: "Siti Aminah",
			Telp:        "081234567801",
		},
		{
			Email:       "admin.stan2@kantin.local",
			Password:    "password123",
			NamaStan:    "Kantin Pak Budi",
			NamaPemilik: "Budi Santoso",
			Telp:        "081234567802",
		},
	}

	for _, s := range stans {
		var exist app.User
		if err := db.Where("email = ?", s.Email).First(&exist).Error; err == nil {
			log.Println("[SEED] stan admin exists:", s.Email)
			continue
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte(s.Password), bcrypt.DefaultCost)

		user := app.User{
			Email:              s.Email,
			PasswordHash:       string(hash),
			Role:               app.RoleAdminStan,
			MustChangePassword: true,
		}
		if err := db.Create(&user).Error; err != nil {
			log.Println("[SEED] failed create admin stan:", err)
			continue
		}

		stan := app.Stan{
			NamaStan:    s.NamaStan,
			NamaPemilik: s.NamaPemilik,
			Telp:        s.Telp,
			UserID:      user.ID,
		}
		if err := db.Create(&stan).Error; err != nil {
			log.Println("[SEED] failed create stan:", err)
			continue
		}

		log.Println("[SEED] stan created:", stan.NamaStan)
	}
}
