package dec128candidate

import (
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

const MaxStrLen = lu.MaxStrLen + 2

var (
	Decimal0         = Zero
	Decimal1         = One
	Decimal2         = FromInt(2)
	Decimal3         = FromInt(3)
	Decimal4         = FromInt(4)
	Decimal5         = FromInt(5)
	Decimal6         = FromInt(6)
	Decimal7         = FromInt(7)
	Decimal8         = FromInt(8)
	Decimal9         = FromInt(9)
	Decimal10        = FromInt(10)
	Decimal100       = FromInt(100)
	Decimal365       = FromInt(365)
	Decimal366       = FromInt(366)
	Decimal1000      = FromInt(1000)
	Ten              = Decimal10
	Hundred          = Decimal100
	ZeroStr          = "0"
	ZeroStrBytes     = []byte("0")
	ZeroJsonStrBytes = []byte(`"0"`)
	NaNStr           = "NaN"
	NaNStrBytes      = []byte("NaN")
	NaNJsonStrBytes  = []byte(`"NaN"`)
	Pow10Uint64      = lu.Pow10Uint64
	Pow10Uint128     = lu.Pow10Uint128
)

func DecodeFromUint128(coef lu.Uint128, scale uint8) Dec128 { return New(coef, scale, false) }
func DecodeFromUint64(coef uint64, scale uint8) Dec128 {
	return New(lu.Uint128{Lo: coef}, scale, false)
}
func Max(a Dec128, bs ...Dec128) Dec128 {
	for _, b := range bs {
		if b.Compare(a) > 0 {
			a = b
		}
	}
	return a
}
func Min(a Dec128, bs ...Dec128) Dec128 {
	for _, b := range bs {
		if b.Compare(a) < 0 {
			a = b
		}
	}
	return a
}
func Sum(a Dec128, bs ...Dec128) Dec128 {
	for _, b := range bs {
		a = a.Add(b)
	}
	return a
}
func Avg(a Dec128, bs ...Dec128) Dec128 {
	if len(bs) == 0 {
		return a
	}
	sum := Sum(a, bs...)
	return sum.Div(FromInt(len(bs) + 1))
}
