package siswa

import (
	"time"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// pickLatestApplicableDiscount: pilih diskon paling relevan dari slice
// (fallback logic bila Anda ingin memilih dari slice)
func pickLatestApplicableDiscount(diskons []app.Diskon) *app.Diskon {
	var best *app.Diskon
	now := time.Now().UTC()
	for i := range diskons {
		d := &diskons[i]

		// skip belum mulai
		if d.TanggalAwal != nil && now.Before(*d.TanggalAwal) {
			continue
		}
		// skip sudah berakhir
		if d.TanggalAkhir != nil && now.After(*d.TanggalAkhir) {
			continue
		}
		// pilih berdasarkan created_at paling baru; jika tie, pilih persentase diskon lebih besar
		if best == nil || d.CreatedAt.After(best.CreatedAt) || (d.CreatedAt.Equal(best.CreatedAt) && d.PersentaseDiskon > best.PersentaseDiskon) {
			best = d
		}
	}
	return best
}

// getStanPublicIDByID returns stan.public_id for given numeric id
// returns empty string if not found.
func getStanPublicIDByID(id uint) string {
	var pub string
	app.DB.Table("stans").Select("public_id").Where("id = ?", id).Scan(&pub)
	return pub
}

// getStanNameByID returns stan.nama_stan for given numeric id
// returns empty string if not found.
func getStanNameByID(id uint) string {
	var name string
	app.DB.Table("stans").Select("nama_stan").Where("id = ?", id).Scan(&name)
	return name
}
