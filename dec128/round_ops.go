package dec128

import (
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// RoundingMode specifies how an exact result is quantized to a decimal scale.
// It is independent of the process-wide default precision used by Div and Sqrt.
type RoundingMode uint8

const (
	// RoundingTowardZero drops fractional digits, regardless of sign.
	RoundingTowardZero RoundingMode = iota
	// RoundingAwayFromZero increases the magnitude on any nonzero remainder.
	RoundingAwayFromZero
	// RoundingFloor rounds toward negative infinity.
	RoundingFloor
	// RoundingCeil rounds toward positive infinity.
	RoundingCeil
	// RoundingHalfTowardZero rounds nearest, with exact ties toward zero.
	RoundingHalfTowardZero
	// RoundingHalfAwayFromZero rounds nearest, with exact ties away from zero.
	RoundingHalfAwayFromZero
	// RoundingHalfEven rounds nearest, with exact ties to an even coefficient.
	RoundingHalfEven
)

// compare256 compares two unsigned 256-bit numbers in (low, high) form.
func compare256(aLow, aHigh, bLow, bHigh uint128.Uint128) int {
	if c := aHigh.Compare(bHigh); c != 0 {
		return c
	}
	return aLow.Compare(bLow)
}

// quantizeQuotient rounds an unsigned integer quotient using an exact
// nonnegative remainder and a nonzero 256-bit denominator. The quotient and
// remainder are Uint128 values; the denominator may need 256 bits when
// rescaling a divisor with a large coefficient.
func quantizeQuotient(q, r, denLow, denHigh uint128.Uint128, neg bool, prec uint8, mode RoundingMode) (Dec128, bool) {
	if mode > RoundingHalfEven {
		return NaN(errors.InvalidFormat), false
	}
	inexact := !r.IsZero()
	if inexact {
		increment := false
		switch mode {
		case RoundingTowardZero:
		case RoundingAwayFromZero:
			increment = true
		case RoundingFloor:
			increment = neg
		case RoundingCeil:
			increment = !neg
		case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
			twiceLow, twiceHigh := r.MulCarry(uint128.FromUint64(2))
			cmp := compare256(twiceLow, twiceHigh, denLow, denHigh)
			increment = cmp > 0 || cmp == 0 && (mode == RoundingHalfAwayFromZero || mode == RoundingHalfEven && q.Lo&1 != 0)
		}
		if increment {
			var err errors.Error
			q, err = q.Add64(1)
			if err != errors.None {
				return NaN(err), inexact
			}
		}
	}
	return New(q, prec, neg && !q.IsZero()), inexact
}

// MulRound multiplies two decimals exactly before rounding the product once
// to prec fractional digits. The intermediate product uses 256-bit arithmetic
// rather than a 128-bit rounded intermediate. Neither input is mutated.
// It neither reads nor changes SetDefaultPrecision. Errors are returned as NaN.
func (decimal Dec128) MulRound(other Dec128, prec uint8, mode RoundingMode) Dec128 {
	result, _ := decimal.MulRoundWithStatus(other, prec, mode)
	return result
}

// MulRoundWithStatus behaves as MulRound and also reports inexact=true if a
// nonzero fraction of the target coefficient was discarded. A NaN result
// signals an error; the status alone must not be used to detect errors.
func (decimal Dec128) MulRoundWithStatus(other Dec128, prec uint8, mode RoundingMode) (Dec128, bool) {
	if decimal.errorCode() != errors.None {
		return decimal, false
	}
	if other.errorCode() != errors.None {
		return other, false
	}
	if prec > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange), false
	}
	if mode > RoundingHalfEven {
		return NaN(errors.InvalidFormat), false
	}

	low, high := decimal.coef.MulCarry(other.coef)
	scale := int(decimal.exp) + int(other.exp) - int(prec)
	neg := decimal.negative() != other.negative()
	if scale < 0 {
		// Increasing the scale cannot make a non-128-bit coefficient fit.
		if !high.IsZero() {
			return NaN(errors.Overflow), false
		}
		coef, err := low.Mul(Pow10Uint128[-scale])
		if err != errors.None {
			return NaN(err), false
		}
		return New(coef, prec, neg && !coef.IsZero()), false
	}

	denominator := Pow10Uint128[scale] // 0 <= scale <= 38
	q, r, err := uint128.QuoRem256By128(low, high, denominator)
	if err != errors.None {
		return NaN(err), false
	}
	return quantizeQuotient(q, r, denominator, uint128.Zero, neg, prec, mode)
}

// DivRound divides decimals at the explicitly requested precision and rounds
// the exact rational quotient once. Unlike Div, it does not use the process-
// wide defaultPrecision and does not always truncate.
func (decimal Dec128) DivRound(other Dec128, prec uint8, mode RoundingMode) Dec128 {
	result, _ := decimal.DivRoundWithStatus(other, prec, mode)
	return result
}

// DivRoundWithStatus behaves as DivRound and reports whether the exact
// quotient had a nonzero remainder at the requested scale. It does not change
// SetDefaultPrecision. Invalid precision/mode, overflow or division by zero
// produce NaN, consistent with existing decimal operations.
func (decimal Dec128) DivRoundWithStatus(other Dec128, prec uint8, mode RoundingMode) (Dec128, bool) {
	if decimal.errorCode() != errors.None {
		return decimal, false
	}
	if other.errorCode() != errors.None {
		return other, false
	}
	if prec > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange), false
	}
	if mode > RoundingHalfEven {
		return NaN(errors.InvalidFormat), false
	}
	if other.coef.IsZero() {
		return NaN(errors.DivisionByZero), false
	}

	scale := int(other.exp) + int(prec) - int(decimal.exp)
	neg := decimal.negative() != other.negative()
	if scale >= 0 {
		low, high := decimal.coef.MulCarry(Pow10Uint128[scale])
		q, r, err := uint128.QuoRem256By128(low, high, other.coef)
		if err != errors.None {
			return NaN(err), false
		}
		return quantizeQuotient(q, r, other.coef, uint128.Zero, neg, prec, mode)
	}

	// Rescale the denominator instead when the numerator has more fractional
	// digits than the requested quotient. The scaled denominator can exceed
	// 128 bits, but if so the integer quotient must be zero since the original
	// numerator fits in 128 bits. Keep the full denominator for half rounding.
	denLow, denHigh := other.coef.MulCarry(Pow10Uint128[-scale])
	if !denHigh.IsZero() {
		return quantizeQuotient(uint128.Zero, decimal.coef, denLow, denHigh, neg, prec, mode)
	}
	q, r, err := decimal.coef.QuoRem(denLow)
	if err != errors.None {
		return NaN(err), false
	}
	return quantizeQuotient(q, r, denLow, uint128.Zero, neg, prec, mode)
}
