package app

import "math"

func Round2(f float64) float64 {
	return math.Round(f*100) / 100
}
