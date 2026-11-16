package internal

import (
	"math"
	"strconv"
	"strings"
)

// Importe subtracts half of 100 raised to scale from base.
func Importe(base float64, scale int) float64 {
	return base - math.Pow(100, float64(scale))/2
}

// NumDecPlaces returns the number of fractional digits in the shortest
// fixed-point decimal representation of v.
func NumDecPlaces(v float64) int {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if i < 0 {
		return 0
	}
	return len(s) - i - 1
}
