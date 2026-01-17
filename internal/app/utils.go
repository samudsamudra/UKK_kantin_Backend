package app

import (
	"fmt"
	"math"
	"time"
)

// Round2 membulatkan float ke 2 desimal
func Round2(f float64) float64 {
	return math.Round(f*100) / 100
}

// FormatTimeHuman mengembalikan representasi human-friendly untuk waktu.
// - jika dalam range +/-24 jam -> relative (mis. "7 menit lalu" / "in 2 hours")
// - jika di luar -> format pendek "02 Jan 2006 15:04" (waktu lokal Asia/Jakarta)
func FormatTimeHuman(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	// convert to Jakarta timezone
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err == nil {
		t = t.In(loc)
	}

	now := time.Now().In(t.Location())
	diff := now.Sub(t)

	// future
	if diff < 0 {
		// abs duration
		abs := -diff
		switch {
		case abs < time.Minute:
			return "sebentar lagi"
		case abs < time.Hour:
			return fmt.Sprintf("dalam %d menit", int(abs.Minutes()))
		case abs < 24*time.Hour:
			return fmt.Sprintf("dalam %d jam", int(abs.Hours()))
		default:
			return t.Format("02 Jan 2006 15:04")
		}
	}

	// past
	switch {
	case diff < time.Minute:
		return "baru saja"
	case diff < time.Hour:
		return fmt.Sprintf("%d menit lalu", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d jam lalu", int(diff.Hours()))
	default:
		return t.Format("02 Jan 2006 15:04")
	}
}

// FormatTimePretty returns a short human-friendly representation for a *time.Time.
// If t is nil, returns empty string. This reuses FormatTimeHuman.
func FormatTimePretty(t *time.Time) string {
	if t == nil {
		return ""
	}
	return FormatTimeHuman((*t).In(time.FixedZone("WIB", 7*3600)))
}

// FormatISOOrNil returns RFC3339 UTC string or nil if t is nil.
// Return type is interface{} so when marshalled to JSON it becomes null when nil.
func FormatISOOrNil(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}
// FormatTimeWithClock returns:
// "12:05 (5 menit lalu)" or "12:05 (baru saja)"
// If older than 24h, fallback to full date.
func FormatTimeWithClock(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err == nil {
		t = t.In(loc)
	}

	now := time.Now().In(t.Location())
	diff := now.Sub(t)

	clock := t.Format("15:04")

	// future
	if diff < 0 {
		abs := -diff
		switch {
		case abs < time.Minute:
			return clock + " (sebentar lagi)"
		case abs < time.Hour:
			return clock + " (dalam " + fmt.Sprintf("%d", int(abs.Minutes())) + " menit)"
		case abs < 24*time.Hour:
			return clock + " (dalam " + fmt.Sprintf("%d", int(abs.Hours())) + " jam)"
		default:
			return t.Format("02 Jan 2006 15:04")
		}
	}

	// past
	switch {
	case diff < time.Minute:
		return clock + " (baru saja)"
	case diff < time.Hour:
		return clock + " (" + fmt.Sprintf("%d", int(diff.Minutes())) + " menit lalu)"
	case diff < 24*time.Hour:
		return clock + " (" + fmt.Sprintf("%d", int(diff.Hours())) + " jam lalu)"
	default:
		return t.Format("02 Jan 2006 15:04")
	}
}
// =========================
// DISCOUNT HELPERS (UKK)
// =========================

// GetActiveDiscount mengambil 1 diskon aktif (GLOBAL).
// Diskon aktif jika:
// - tanggal_awal NULL atau <= sekarang
// - tanggal_akhir NULL atau >= sekarang
// Jika tidak ada, return nil.
func GetActiveDiscount() *Diskon {
	var d Diskon
	now := time.Now().UTC()

	err := DB.
		Where(
			"(tanggal_awal IS NULL OR tanggal_awal <= ?) AND (tanggal_akhir IS NULL OR tanggal_akhir >= ?)",
			now, now,
		).
		Order("created_at DESC").
		First(&d).Error

	if err != nil {
		return nil
	}
	return &d
}

// ApplyDiscount menghitung harga setelah diskon persen
func ApplyDiscount(price float64, percent float64) float64 {
	if percent <= 0 {
		return Round2(price)
	}
	disc := price * (percent / 100)
	return Round2(price - disc)
}
// FormatDateID formats date into Indonesian human-readable date.
// short = true  -> "17 Jan 2026"
// short = false -> "17 Januari 2026"
func FormatDateID(t *time.Time, short bool) string {
	if t == nil {
		return ""
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err == nil {
		t = func(tt time.Time) *time.Time {
			lt := tt.In(loc)
			return &lt
		}(*t)
	}

	day := t.Day()
	year := t.Year()

	monthsFull := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	monthsShort := []string{
		"Jan", "Feb", "Mar", "Apr", "Mei", "Jun",
		"Jul", "Agu", "Sep", "Okt", "Nov", "Des",
	}

	month := ""
	if short {
		month = monthsShort[int(t.Month())-1]
	} else {
		month = monthsFull[int(t.Month())-1]
	}

	return fmt.Sprintf("%d %s %d", day, month, year)
}
