package internal

import (
	"math"
	"strconv"
	"strings"
)

func RoundHalfUp(x float64, prec int) float64 {
	if x == 0 {

		return 0
	}

	if prec >= 0 && x == math.Trunc(x) {
		return x
	}

	pow := math.Pow10(prec)
	intermed := x * pow
	if math.IsInf(intermed, 0) {
		return x
	}

	if NumDecPlaces(intermed) > NumDecPlaces(x) {
		intermed = RoundHalfUp(intermed, prec)
	}

	x = math.Round(intermed)

	if x == 0 {
		return 0
	}

	return x / pow
}

func NumDecPlaces(v float64) int {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if i > -1 {
		return len(s) - i - 1
	}
	return 0
}
