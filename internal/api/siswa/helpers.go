package siswa

import (
	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// getStanPublicIDByID returns stan.public_id
func getStanPublicIDByID(id uint) string {
	var pub string
	app.DB.Table("stans").
		Select("public_id").
		Where("id = ?", id).
		Scan(&pub)
	return pub
}

// getStanNameByID returns stan.nama_stan
func getStanNameByID(id uint) string {
	var name string
	app.DB.Table("stans").
		Select("nama_stan").
		Where("id = ?", id).
		Scan(&name)
	return name
}
