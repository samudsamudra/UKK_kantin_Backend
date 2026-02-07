package seed

import (
	"log"
	"os"

	"github.com/samudsamudra/UKK_kantin/internal/app"
	"golang.org/x/crypto/bcrypt"
)

func SeedSuperAdmin() {
	db := app.DB

	email := "root@system.local"
	password := os.Getenv("SUPERADMIN_PASSWORD")

	if password == "" {
		log.Println("SUPERADMIN_PASSWORD not set")
		return
	}

	var existing app.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		log.Println("Superadmin already exists, skip seed")
		return
	}

	// HASH sesuai login logic
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("failed hashing password:", err)
		return
	}

	superAdmin := app.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         app.RoleSuperAdmin,
	}

	if err := db.Create(&superAdmin).Error; err != nil {
		log.Println("failed create superadmin:", err)
		return
	}

	log.Println("Superadmin seeded successfully")
}
