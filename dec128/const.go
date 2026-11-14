package dec128

import (
	"sync/atomic"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// MaxPrecision is the maximum number of digits after the decimal point that can be represented.
// MaxPrecision = 19
const MaxPrecision = uint8(uint128.MaxSafeStrLen64)

// MaxStrLen is the maximum number of characters that can be in a string representation of a Dec128.
// MaxStrLen = uint128.MaxStrLen + dot + sign
const MaxStrLen = uint128.MaxStrLen + 2

var (
	// Zero is the predefined decimal value zero.
	Zero = Dec128{}
	// One is the predefined decimal value one.
	One = FromInt(1)
	// NegativeOne is the predefined decimal value minus one.
	NegativeOne = FromInt(-1)

	// Decimal0 is the predefined decimal value 0.
	Decimal0 = Zero
	// Decimal1 is the predefined decimal value 1.
	Decimal1 = One
	// Decimal2 is the predefined decimal value 2.
	Decimal2 = FromInt(2)
	// Decimal3 is the predefined decimal value 3.
	Decimal3 = FromInt(3)
	// Decimal4 is the predefined decimal value 4.
	Decimal4 = FromInt(4)
	// Decimal5 is the predefined decimal value 5.
	Decimal5 = FromInt(5)
	// Decimal6 is the predefined decimal value 6.
	Decimal6 = FromInt(6)
	// Decimal7 is the predefined decimal value 7.
	Decimal7 = FromInt(7)
	// Decimal8 is the predefined decimal value 8.
	Decimal8 = FromInt(8)
	// Decimal9 is the predefined decimal value 9.
	Decimal9 = FromInt(9)
	// Decimal10 is the predefined decimal value 10.
	Decimal10 = FromInt(10)
	// Decimal100 is the predefined decimal value 100.
	Decimal100 = FromInt(100)
	// Decimal365 is the predefined decimal value 365.
	Decimal365 = FromInt(365)
	// Decimal366 is the predefined decimal value 366.
	Decimal366 = FromInt(366)
	// Decimal1000 is the predefined decimal value 1000.
	Decimal1000 = FromInt(1000)

	// Ten is a predefined decimal value equal to ten.
	Ten = Decimal10
	// Hundred is a predefined decimal value equal to one hundred.
	Hundred = Decimal100

	// ZeroStr is the decimal text for zero.
	ZeroStr = "0"
	// ZeroStrBytes contains the mutable bytes of the text zero. Treat this exported variable as read only.
	ZeroStrBytes = []byte(ZeroStr)
	// ZeroJsonStrBytes contains the mutable JSON string bytes for zero. Treat this exported variable as read only.
	ZeroJsonStrBytes = []byte(`"0"`)

	// NaNStr is the decimal text for an invalid value.
	NaNStr = "NaN"
	// NaNStrBytes contains the mutable bytes of the text NaN. Treat this exported variable as read only.
	NaNStrBytes = []byte(NaNStr)
	// NaNJsonStrBytes contains the mutable JSON string bytes for NaN. Treat this exported variable as read only.
	NaNJsonStrBytes = []byte(`"NaN"`)

	// Pow10Uint64 holds precomputed powers of ten for this coefficient width. Treat the exported table as read only.
	Pow10Uint64 = uint128.Pow10Uint64
	// Pow10Uint128 holds precomputed powers of ten for this coefficient width. Treat the exported table as read only.
	Pow10Uint128 = uint128.Pow10Uint128

	defaultPrecision atomic.Uint32
)

func init() {
	defaultPrecision.Store(uint32(MaxPrecision))
}

// SetDefaultPrecision sets the default precision for all Dec128 instances, where precision is the number of digits after the decimal point.
func SetDefaultPrecision(prec uint8) {
	if prec > MaxPrecision {
		panic(errors.PrecisionOutOfRange.Value())
	}
	defaultPrecision.Store(uint32(prec))
}
