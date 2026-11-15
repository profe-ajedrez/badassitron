package internal

import "github.com/profe-ajedrez/badassitron/dec128"

func Importe(base dec128.Dec128, scale int) dec128.Dec128 {
	sc := dec128.Decimal100.PowInt(scale) //PowInt(int32(scale))
	sc = sc.Div(dec128.Decimal2)
	base.Sub(sc)
	return base
}

func NumDecPlaces(v dec128.Dec128) int {
	return int(v.Precision())
}
