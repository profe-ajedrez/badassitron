package dec128

import (
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// RescaleRound changes the number of fractional decimal digits to prec,
// rounding the exact value once according to mode when reducing the scale.
// Unlike Rescale, reducing the scale does not always truncate toward zero.
// Increasing the scale is exact when the scaled coefficient fits Uint128.
// It never reads or changes the process-wide default precision.
// Invalid precision/mode and coefficient overflow produce NaN.
func (decimal Dec128) RescaleRound(prec uint8, mode RoundingMode) Dec128 {
	result, _ := decimal.RescaleRoundWithStatus(prec, mode)
	return result
}

// RescaleRoundWithStatus behaves like RescaleRound and reports inexact=true
// precisely when a nonzero fractional part was discarded at the requested
// scale, regardless of whether the rounded coefficient was incremented.
// An inexact result is not an error; check IsNaN separately. An error returns
// NaN and inexact=false. Exact zeros retain their existing sign bit internally.
func (decimal Dec128) RescaleRoundWithStatus(prec uint8, mode RoundingMode) (Dec128, bool) {
	if decimal.errorCode() != errors.None {
		return decimal, false
	}
	if prec > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange), false
	}
	if mode > RoundingHalfEven {
		return NaN(errors.InvalidFormat), false
	}
	if prec == decimal.exp {
		return decimal, false
	}
	if decimal.coef.IsZero() {
		return Dec128{exp: prec, state: signState(decimal.negative())}, false
	}

	if prec > decimal.exp {
		// Multiplying by 10^(prec-exp) is exact, unless the resulting
		// coefficient cannot be represented in 128 bits.
		coef, err := decimal.coef.Mul64(Pow10Uint64[prec-decimal.exp])
		if err != errors.None {
			return NaN(err), false
		}
		return Dec128{coef: coef, exp: prec, state: signState(decimal.negative())}, false
	}

	// A scale reduction divides by at most 10^19, so both the quotient
	// and the exact remainder can be computed with Uint128 and uint64.
	factor := Pow10Uint64[decimal.exp-prec]
	q, r, err := decimal.coef.QuoRem64(factor)
	if err != errors.None {
		return NaN(err), false
	}
	return quantizeQuotient(q, uint128.FromUint64(r), uint128.FromUint64(factor), uint128.Zero, decimal.negative(), prec, mode)
}
