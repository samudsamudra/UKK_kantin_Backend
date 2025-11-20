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
