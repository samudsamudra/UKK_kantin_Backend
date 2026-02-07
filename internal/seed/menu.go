package seed

import (
	"log"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

func SeedMenus() {
	db := app.DB

	var stans []app.Stan
	if err := db.Find(&stans).Error; err != nil {
		log.Println("[SEED] failed fetch stans")
		return
	}

	for _, stan := range stans {
		menus := []app.Menu{
			{
				NamaMakanan: "Nasi Goreng",
				Harga:       15000,
				Jenis:       app.JenisMakanan,
				Deskripsi:   "Nasi goreng spesial",
				StanID:      stan.ID,
			},
			{
				NamaMakanan: "Mie Ayam",
				Harga:       12000,
				Jenis:       app.JenisMakanan,
				Deskripsi:   "Mie ayam gurih",
				StanID:      stan.ID,
			},
			{
				NamaMakanan: "Es Teh",
				Harga:       5000,
				Jenis:       app.JenisMinuman,
				Deskripsi:   "Es teh segar",
				StanID:      stan.ID,
			},
		}

		for _, m := range menus {
			var exist app.Menu
			if err := db.
				Where("nama_makanan = ? AND stan_id = ?", m.NamaMakanan, stan.ID).
				First(&exist).Error; err == nil {
				continue
			}

			if err := db.Create(&m).Error; err != nil {
				log.Println("[SEED] failed create menu:", err)
			}
		}

		log.Println("[SEED] menus created for:", stan.NamaStan)
	}
}
