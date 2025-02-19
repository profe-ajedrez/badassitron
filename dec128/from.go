package dec128

import (
	"math"
	"strconv"

	deferr "errors"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func NewFromString(s string) (Dec128, error) {
	v := FromString(s)

	if v.IsNaN() {
		return Dec128{}, deferr.New("invalid decimal number")
	}

	return v, nil
}

func NewFromInt(i int) Dec128 {
	return FromInt(i)
}

// FromString creates a new Dec128 from a string.
// The string must be in the format of [+-][0-9]+(.[0-9]+)?
// In case of empty string, it returns Zero.
// In case of errors, it returns NaN with the corresponding error.
func FromString[S string | []byte](s S) Dec128 {
	sz := len(s)

	switch sz {
	case 0:
		return Zero
	case 1:
		switch s[0] {
		case '0':
			return Zero
		case '+', '-', '.':
			return NaN(errors.InvalidFormat)
		}
	case 2:
		if (s[0] == '+' || s[0] == '-') && s[1] == '.' {
			return NaN(errors.InvalidFormat)
		}
	}

	var i, prec int
	var neg bool

	switch s[0] {
	case '+':
		i++
	case '-':
		neg = true
		i++
	}

	if sz <= uint128.MaxSafeStrLen64 {
		// safe to parse with uint64 as coef
		var u uint64
		for ; i < sz; i++ {
			if s[i] == '.' {
				if prec != 0 {
					return NaN(errors.InvalidFormat)
				}
				prec = sz - i - 1
				continue
			}
			if s[i] < '0' || s[i] > '9' {
				return NaN(errors.InvalidFormat)
			}
			u = u*10 + uint64(s[i]-'0')
		}
		if u == 0 && prec == 0 {
			return Zero
		}
		return Dec128{coef: uint128.FromUint64(u), exp: uint8(prec), neg: neg}
	}

	j := 0
	for ; j < sz; j++ {
		if s[j] == '.' {
			break
		}
	}

	if j == sz {
		coef, err := uint128.FromString(s[i:])
		if err != errors.None {
			return NaN(err)
		}
		return Dec128{coef: coef, exp: 0, neg: neg}
	}

	if j == sz-1 {
		return NaN(errors.InvalidFormat)
	}

	prec = sz - j - 1
	if prec > uint128.MaxSafeStrLen64 {
		return NaN(errors.PrecisionOutOfRange)
	}

	ipart, err := uint128.FromString(s[i:j])
	if err != errors.None {
		return NaN(err)
	}

	fpart, err := uint128.FromString(s[j+1:])
	if err != errors.None {
		return NaN(err)
	}

	// max prec is 19, so the fpart.Hi is always 0 and prec is always <= len(pow10)
	coef, err := ipart.Mul64(Pow10Uint64[prec])
	if err != errors.None {
		return NaN(err)
	}

	coef, err = coef.Add64(fpart.Lo)
	if err != errors.None {
		return NaN(err)
	}

	if coef.IsZero() && prec == 0 {
		return Zero
	}

	return Dec128{coef: coef, exp: uint8(prec), neg: neg}
}

// FromInt creates a new Dec128 from an int.
func FromInt(i int) Dec128 {
	if i == 0 {
		return Zero
	}

	if i > 0 {
		return Dec128{coef: uint128.FromUint64(uint64(i))}
	}

	return Dec128{coef: uint128.FromUint64(uint64(-i)), neg: true}
}

// FromInt64 creates a new Dec128 from an int64.
func FromInt64(i int64) Dec128 {
	if i == 0 {
		return Zero
	}

	if i > 0 {
		return Dec128{coef: uint128.FromUint64(uint64(i))}
	}

	return Dec128{coef: uint128.FromUint64(uint64(-i)), neg: true}
}

// DecodeFromUint128 decodes a Dec128 from a Uint128 and an exponent.
func DecodeFromUint128(coef uint128.Uint128, exp uint8) Dec128 {
	return New(coef, exp, false)
}

// DecodeFromUint64 decodes a Dec128 from a uint64 and an exponent.
func DecodeFromUint64(coef uint64, exp uint8) Dec128 {
	return New(uint128.FromUint64(coef), exp, false)
}

// FromFloat64 returns a decimal from float64.
func FromFloat64(f float64) Dec128 {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return NaN(errors.NotANumber)
	}
	return FromString(strconv.FormatFloat(f, 'f', -1, 64))
}
