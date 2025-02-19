package dec128

import "github.com/profe-ajedrez/badassitron/dec128/errors"

func (decimal Dec128) Round(prec uint8) Dec128 {
	return decimal.RoundHalfAwayFromZero(prec)
}

// RoundDown (or Floor) rounds the decimal to the specified precision using Round Down method (https://en.wikipedia.org/wiki/Rounding#Rounding_down).
//
// Examples:
//
//	RoundDown(1.236, 2) = 1.23
//	RoundDown(1.235, 2) = 1.23
//	RoundDown(1.234, 2) = 1.23
//	RoundDown(-1.234, 2) = -1.24
//	RoundDown(-1.235, 2) = -1.24
//	RoundDown(-1.236, 2) = -1.24
func (decimal Dec128) RoundDown(prec uint8) Dec128 {
	if decimal.err != errors.None || prec >= decimal.exp {
		return decimal
	}

	q, r, err := decimal.coef.QuoRem64(Pow10Uint64[decimal.exp-prec])
	if err != errors.None {
		return NaN(err)
	}

	if decimal.neg && r != 0 {
		q, err = q.Add64(1)
		if err != errors.None {
			return NaN(err)
		}
	}

	return Dec128{coef: q, exp: prec, neg: decimal.neg}
}

// RoundUp (or Ceil) rounds the decimal to the specified precision using Round Up method (https://en.wikipedia.org/wiki/Rounding#Rounding_up).
//
// Examples:
//
//	RoundUp(1.236, 2) = 1.24
//	RoundUp(1.235, 2) = 1.24
//	RoundUp(1.234, 2) = 1.24
//	RoundUp(-1.234, 2) = -1.23
//	RoundUp(-1.235, 2) = -1.23
//	RoundUp(-1.236, 2) = -1.23
func (decimal Dec128) RoundUp(prec uint8) Dec128 {
	if decimal.err != errors.None || prec >= decimal.exp {
		return decimal
	}

	q, r, err := decimal.coef.QuoRem64(Pow10Uint64[decimal.exp-prec])
	if err != errors.None {
		return NaN(err)
	}

	if !decimal.neg && r != 0 {
		q, err = q.Add64(1)
		if err != errors.None {
			return NaN(err)
		}
	}

	return Dec128{coef: q, exp: prec, neg: decimal.neg}
}

// RoundTowardZero rounds the decimal to the specified prec using Toward Zero method (https://en.wikipedia.org/wiki/Rounding#Rounding_toward_zero).
//
// Examples:
//
//	RoundTowardZero(1.236, 2) = 1.23
//	RoundTowardZero(1.235, 2) = 1.23
//	RoundTowardZero(1.234, 2) = 1.23
//	RoundTowardZero(-1.234, 2) = -1.23
//	RoundTowardZero(-1.235, 2) = -1.23
//	RoundTowardZero(-1.236, 2) = -1.23
func (decimal Dec128) RoundTowardZero(prec uint8) Dec128 {
	return decimal.Trunc(prec)
}

// RoundAwayFromZero rounds the decimal to the specified prec using Away From Zero method (https://en.wikipedia.org/wiki/Rounding#Rounding_away_from_zero).
//
// Examples:
//
//	RoundAwayFromZero(1.236, 2) = 1.24
//	RoundAwayFromZero(1.235, 2) = 1.24
//	RoundAwayFromZero(1.234, 2) = 1.24
//	RoundAwayFromZero(-1.234, 2) = -1.24
//	RoundAwayFromZero(-1.235, 2) = -1.24
//	RoundAwayFromZero(-1.236, 2) = -1.24
func (decimal Dec128) RoundAwayFromZero(prec uint8) Dec128 {
	if decimal.err != errors.None {
		return decimal
	}

	if prec >= decimal.exp {
		return decimal
	}

	q, r, err := decimal.coef.QuoRem64(Pow10Uint64[decimal.exp-prec])
	if err != errors.None {
		return NaN(err)
	}

	if r != 0 {
		q, err = q.Add64(1)
		if err != errors.None {
			return NaN(err)
		}
	}

	return Dec128{coef: q, exp: prec, neg: decimal.neg}
}

// RoundHalfTowardZero rounds the decimal to the specified prec using Half Toward Zero method (https://en.wikipedia.org/wiki/Rounding#Rounding_half_toward_zero).
//
// Examples:
//
//	RoundHalfTowardZero(1.236, 2) = 1.24
//	RoundHalfTowardZero(1.235, 2) = 1.23
//	RoundHalfTowardZero(1.234, 2) = 1.23
//	RoundHalfTowardZero(-1.234, 2) = -1.23
//	RoundHalfTowardZero(-1.235, 2) = -1.23
//	RoundHalfTowardZero(-1.236, 2) = -1.24
func (decimal Dec128) RoundHalfTowardZero(prec uint8) Dec128 {
	if decimal.err != errors.None {
		return decimal
	}

	if prec >= decimal.exp {
		return decimal
	}

	factor := Pow10Uint64[decimal.exp-prec]
	half := factor / 2

	q, r, err := decimal.coef.QuoRem64(factor)
	if err != errors.None {
		return NaN(err)
	}

	if half < r {
		q, err = q.Add64(1)
		if err != errors.None {
			return NaN(err)
		}
	}

	return Dec128{coef: q, exp: prec, neg: decimal.neg}
}

// RoundHalfAwayFromZero rounds the decimal to the specified prec using Half Away from Zero method (https://en.wikipedia.org/wiki/Rounding#Rounding_half_away_from_zero).
//
// Examples:
//
//	RoundHalfAwayFromZero(1.236, 2) = 1.24
//	RoundHalfAwayFromZero(1.235, 2) = 1.24
//	RoundHalfAwayFromZero(1.234, 2) = 1.23
//	RoundHalfAwayFromZero(-1.234, 2) = -1.23
//	RoundHalfAwayFromZero(-1.235, 2) = -1.24
//	RoundHalfAwayFromZero(-1.236, 2) = -1.24
func (decimal Dec128) RoundHalfAwayFromZero(prec uint8) Dec128 {
	if decimal.err != errors.None {
		return decimal
	}

	if prec >= decimal.exp {
		return decimal
	}

	factor := Pow10Uint64[decimal.exp-prec]
	half := factor / 2

	q, r, err := decimal.coef.QuoRem64(factor)
	if err != errors.None {
		return NaN(err)
	}

	if half <= r {
		q, err = q.Add64(1)
		if err != errors.None {
			return NaN(err)
		}
	}

	return Dec128{coef: q, exp: prec, neg: decimal.neg}
}

// RoundBank uses half up to even (banker's rounding) to round the decimal to the specified precision.
//
// Examples:
//
//	RoundBank(2.121, 2) = 2.12 ; rounded down
//	RoundBank(2.125, 2) = 2.12 ; rounded down, rounding digit is an even number
//	RoundBank(2.135, 2) = 2.14 ; rounded up, rounding digit is an odd number
//	RoundBank(2.1351, 2) = 2.14; rounded up
//	RoundBank(2.127, 2) = 2.13 ; rounded up
func (decimal Dec128) RoundBank(prec uint8) Dec128 {
	if decimal.err != errors.None {
		return decimal
	}

	if prec >= decimal.exp {
		return decimal
	}

	factor := Pow10Uint64[decimal.exp-prec]
	half := factor / 2

	q, r, err := decimal.coef.QuoRem64(factor)
	if err != errors.None {
		return NaN(err)
	}

	if half < r || (half == r && q.Lo%2 == 1) {
		q, err = q.Add64(1)
		if err != errors.None {
			return NaN(err)
		}
	}

	return Dec128{coef: q, exp: prec, neg: decimal.neg}
}

// Trunc returns 'decimal' after truncating the decimal to the specified precision.
//
// Examples:
//
//	Trunc(1.12345, 4) = 1.1234
//	Trunc(1.12335, 4) = 1.1233
func (decimal Dec128) Trunc(prec uint8) Dec128 {
	if decimal.err != errors.None {
		return decimal
	}

	if prec >= decimal.exp {
		return decimal
	}

	q, _, err := decimal.coef.QuoRem64(Pow10Uint64[decimal.exp-prec])
	if err != errors.None {
		return NaN(err)
	}

	return Dec128{coef: q, exp: prec, neg: decimal.neg}
}
