package dec128bridge

import (
	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// MulRoundWithStatus performs one 256-bit product and one quantization.
// The coefficient operations come from the imported Jokruger package; the
// decision about rounding and errors preserves Badassitron's public contract.
// Inexact is true exactly when a nonzero fractional remainder was discarded.
func MulRoundWithStatus(a, b legacy.Dec128, precision uint8, mode legacy.RoundingMode) (legacy.Dec128, bool) {
	if a.IsNaN() {
		return a, false
	}
	if b.IsNaN() {
		return b, false
	}
	if precision > legacy.MaxPrecision {
		return legacy.NaN(le.PrecisionOutOfRange), false
	}
	if _, ok := toRoundingMode(mode); !ok {
		return legacy.NaN(le.InvalidFormat), false
	}

	ac, bc := a.Coefficient(), b.Coefficient()
	lo, hi := (ju.Uint128{Lo: ac.Lo, Hi: ac.Hi}).MulCarry(ju.Uint128{Lo: bc.Lo, Hi: bc.Hi})
	shift := int(a.Precision()) + int(b.Precision()) - int(precision)
	negative := a.IsNegative() != b.IsNegative()
	var quotient, remainder, denominator ju.Uint128
	if shift < 0 {
		if !hi.IsZero() {
			return legacy.NaN(le.Overflow), false
		}
		var st state.State
		quotient, st = lo.Mul(ju.Pow10Uint128[-shift])
		if st != state.OK {
			return legacy.NaN(le.Overflow), false
		}
	} else {
		denominator = ju.Pow10Uint128[shift]
		var st state.State
		quotient, remainder, st = ju.QuoRem256By128(lo, hi, denominator)
		if st != state.OK {
			return legacy.NaN(le.Overflow), false
		}
	}
	inexact := !remainder.IsZero()
	if inexact {
		increment := false
		switch mode {
		case legacy.RoundingAwayFromZero:
			increment = true
		case legacy.RoundingFloor:
			increment = negative
		case legacy.RoundingCeil:
			increment = !negative
		case legacy.RoundingHalfTowardZero, legacy.RoundingHalfAwayFromZero, legacy.RoundingHalfEven:
			// Compare 2*remainder with denominator without a 129-bit value.
			complement, _ := denominator.Sub(remainder)
			cmp := remainder.Compare(complement)
			increment = cmp > 0 || cmp == 0 && (mode == legacy.RoundingHalfAwayFromZero || mode == legacy.RoundingHalfEven && quotient.Lo&1 != 0)
		}
		if increment {
			var st state.State
			quotient, st = quotient.Add64(1)
			if st != state.OK {
				return legacy.NaN(le.Overflow), true
			}
		}
	}
	return legacy.New(lu.Uint128{Lo: quotient.Lo, Hi: quotient.Hi}, precision, negative && !quotient.IsZero()), inexact
}

// DivRoundWithStatus divides and rounds once using the imported Jokruger
// uint128 primitives. It bypasses upstream Dec128 conversion and its global
// settings, preserving Badassitron's mode and state vocabulary.
//
// The denominator may exceed 128 bits after scale alignment. In that case
// the quotient must be zero, but rounding still compares the full 256-bit
// denominator against twice the remainder.
func DivRoundWithStatus(a, b legacy.Dec128, precision uint8, mode legacy.RoundingMode) (legacy.Dec128, bool) {
	if a.IsNaN() {
		return a, false
	}
	if b.IsNaN() {
		return b, false
	}
	if precision > legacy.MaxPrecision {
		return legacy.NaN(le.PrecisionOutOfRange), false
	}
	if _, ok := toRoundingMode(mode); !ok {
		return legacy.NaN(le.InvalidFormat), false
	}
	if b.IsZero() {
		return legacy.NaN(le.DivisionByZero), false
	}

	ac, bc := a.Coefficient(), b.Coefficient()
	numerator := ju.Uint128{Lo: ac.Lo, Hi: ac.Hi}
	denominator := ju.Uint128{Lo: bc.Lo, Hi: bc.Hi}
	shift := int(b.Precision()) + int(precision) - int(a.Precision())
	negative := a.IsNegative() != b.IsNegative()

	var q, r, denHigh ju.Uint128
	if shift >= 0 {
		low, high := numerator.MulCarry(ju.Pow10Uint128[shift])
		var status state.State
		q, r, status = ju.QuoRem256By128(low, high, denominator)
		if status != state.OK {
			return legacy.NaN(le.Overflow), false
		}
	} else {
		var status state.State
		denominator, denHigh = denominator.MulCarry(ju.Pow10Uint128[-shift])
		if denHigh.IsZero() {
			q, r, status = numerator.QuoRem(denominator)
			if status != state.OK {
				return legacy.NaN(le.Overflow), false
			}
		} else {
			// numerator is 128-bit and scaled denominator is >128-bit.
			r = numerator
		}
	}
	return roundedDivisionResult(q, r, denominator, denHigh, precision, mode, negative)
}

func roundedDivisionResult(q, r, denLow, denHigh ju.Uint128, scale uint8, mode legacy.RoundingMode, negative bool) (legacy.Dec128, bool) {
	inexact := !r.IsZero()
	if inexact {
		increase := false
		switch mode {
		case legacy.RoundingAwayFromZero:
			increase = true
		case legacy.RoundingFloor:
			increase = negative
		case legacy.RoundingCeil:
			increase = !negative
		case legacy.RoundingHalfTowardZero, legacy.RoundingHalfAwayFromZero, legacy.RoundingHalfEven:
			twiceLow, twiceHigh := r.MulCarry(ju.Uint128{Lo: 2})
			comparison := twiceHigh.Compare(denHigh)
			if comparison == 0 {
				comparison = twiceLow.Compare(denLow)
			}
			increase = comparison > 0 || (comparison == 0 && (mode == legacy.RoundingHalfAwayFromZero || (mode == legacy.RoundingHalfEven && q.Lo&1 == 1)))
		}
		if increase {
			next, status := q.Add64(1)
			if status != state.OK {
				return legacy.NaN(le.Overflow), true
			}
			q = next
		}
	}
	return legacy.New(lu.Uint128{Lo: q.Lo, Hi: q.Hi}, scale, negative && !q.IsZero()), inexact
}
