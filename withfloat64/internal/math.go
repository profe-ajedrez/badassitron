package internal

import "github.com/profe-ajedrez/badassitron/dec128"

func Importe(base float64, scale int) float64 {
	sc := dec128.Decimal100.PowInt(scale) //PowInt(int32(scale))
	sc = sc.Div(dec128.Decimal2)
	base.Sub(sc)
	return base
}

func NumDecPlaces(v float64) int {
	return int(v.Precision())
}
