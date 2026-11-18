package dec128candidate

import (
	jk "github.com/jokruger/dec128"
	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// MulRoundWithStatus performs one 256-bit product and one quantization.
// The coefficient operations come from the imported Jokruger package; the
// decision about rounding and errors preserves Badassitron's public contract.
// Inexact is true exactly when a nonzero fractional remainder was discarded.
func (a Dec128) MulRoundWithStatus(b Dec128, precision uint8, mode RoundingMode) (Dec128, bool) {
	if a.IsNaN() {
		return a, false
	}
	if b.IsNaN() {
		return b, false
	}
	if precision > MaxPrecision {
		return NaN(le.PrecisionOutOfRange), false
	}
	if _, ok := toRoundingMode(mode); !ok {
		return NaN(le.InvalidFormat), false
	}

	ac, bc := a.Coefficient(), b.Coefficient()
	lo, hi := (ju.Uint128{Lo: ac.Lo, Hi: ac.Hi}).MulCarry(ju.Uint128{Lo: bc.Lo, Hi: bc.Hi})
	shift := int(a.Precision()) + int(b.Precision()) - int(precision)
	negative := a.IsNegative() != b.IsNegative()
	var quotient, remainder, denominator ju.Uint128
	if shift < 0 {
		if !hi.IsZero() {
			return NaN(le.Overflow), false
		}
		var st state.State
		quotient, st = lo.Mul(ju.Pow10Uint128[-shift])
		if st != state.OK {
			return NaN(le.Overflow), false
		}
	} else {
		denominator = ju.Pow10Uint128[shift]
		var st state.State
		quotient, remainder, st = ju.QuoRem256By128(lo, hi, denominator)
		if st != state.OK {
			return NaN(le.Overflow), false
		}
	}
	inexact := !remainder.IsZero()
	if inexact {
		increment := false
		switch mode {
		case RoundingAwayFromZero:
			increment = true
		case RoundingFloor:
			increment = negative
		case RoundingCeil:
			increment = !negative
		case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
			// Compare 2*remainder with denominator without a 129-bit value.
			complement, _ := denominator.Sub(remainder)
			cmp := remainder.Compare(complement)
			increment = cmp > 0 || cmp == 0 && (mode == RoundingHalfAwayFromZero || mode == RoundingHalfEven && quotient.Lo&1 != 0)
		}
		if increment {
			var st state.State
			quotient, st = quotient.Add64(1)
			if st != state.OK {
				return NaN(le.Overflow), true
			}
		}
	}
	return New(lu.Uint128{Lo: quotient.Lo, Hi: quotient.Hi}, precision, negative && !quotient.IsZero()), inexact
}

// DivRoundWithStatus divides and rounds once using the imported Jokruger
// uint128 primitives. It bypasses upstream Dec128 conversion and its global
// settings, preserving Badassitron's mode and state vocabulary.
//
// The denominator may exceed 128 bits after scale alignment. In that case
// the quotient must be zero, but rounding still compares the full 256-bit
// denominator against twice the remainder.
func (a Dec128) DivRoundWithStatus(b Dec128, precision uint8, mode RoundingMode) (Dec128, bool) {
	if a.IsNaN() {
		return a, false
	}
	if b.IsNaN() {
		return b, false
	}
	if precision > MaxPrecision {
		return NaN(le.PrecisionOutOfRange), false
	}
	if _, ok := toRoundingMode(mode); !ok {
		return NaN(le.InvalidFormat), false
	}
	if b.IsZero() {
		return NaN(le.DivisionByZero), false
	}

	ac, bc := a.Coefficient(), b.Coefficient()
	numerator := ju.Uint128{Lo: ac.Lo, Hi: ac.Hi}
	denominator := ju.Uint128{Lo: bc.Lo, Hi: bc.Hi}
	shift := int(b.Precision()) + int(precision) - int(a.Precision())
	negative := a.IsNegative() != b.IsNegative()

	var q, r, denHigh ju.Uint128
	if shift >= 0 {
		// Most financial operands stay inside 128 bits after scale alignment.
		// Avoid 256/128 division unless the scaled numerator truly requires it.
		if shift == 0 {
			q, r = quotientAndRemainder128(numerator, denominator)
		} else {
			low, high := numerator.MulCarry(ju.Pow10Uint128[shift])
			if high.IsZero() {
				q, r = quotientAndRemainder128(low, denominator)
			} else {
				var status state.State
				q, r, status = ju.QuoRem256By128(low, high, denominator)
				if status != state.OK {
					return NaN(le.Overflow), false
				}
			}
		}
	} else {
		denominator, denHigh = denominator.MulCarry(ju.Pow10Uint128[-shift])
		if denHigh.IsZero() {
			q, r = quotientAndRemainder128(numerator, denominator)
		} else {
			// numerator is 128-bit and scaled denominator is >128-bit.
			r = numerator
		}
	}
	return roundedDivisionResult(q, r, denominator, denHigh, precision, mode, negative)
}

// quotientAndRemainder128 is called only with a nonzero divisor. The 64-bit
// divisor path avoids the general 128-bit and 256-bit division algorithms.
func quotientAndRemainder128(numerator, denominator ju.Uint128) (ju.Uint128, ju.Uint128) {
	if denominator.Hi == 0 {
		q, r, _ := numerator.QuoRem64(denominator.Lo)
		return q, ju.Uint128{Lo: r}
	}
	q, r, _ := numerator.QuoRem(denominator)
	return q, r
}

func roundedDivisionResult(q, r, denLow, denHigh ju.Uint128, scale uint8, mode RoundingMode, negative bool) (Dec128, bool) {
	inexact := !r.IsZero()
	if inexact {
		increase := false
		switch mode {
		case RoundingAwayFromZero:
			increase = true
		case RoundingFloor:
			increase = negative
		case RoundingCeil:
			increase = !negative
		case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
			var comparison int
			if denHigh.IsZero() {
				// Here 0 < remainder < denominator <= 2^128-1. Comparing
				// remainder with (denominator-remainder) avoids a wide
				// multiply and handles ties without a 129-bit temporary.
				complement, _ := denLow.Sub(r)
				comparison = r.Compare(complement)
			} else {
				// The scaled divisor can require 256 bits. Keep its high
				// limb for the exact half-way comparison.
				twiceLow, twiceHigh := r.MulCarry(ju.Uint128{Lo: 2})
				comparison = twiceHigh.Compare(denHigh)
				if comparison == 0 {
					comparison = twiceLow.Compare(denLow)
				}
			}
			increase = comparison > 0 || (comparison == 0 && (mode == RoundingHalfAwayFromZero || (mode == RoundingHalfEven && q.Lo&1 == 1)))
		}
		if increase {
			next, status := q.Add64(1)
			if status != state.OK {
				return NaN(le.Overflow), true
			}
			q = next
		}
	}
	return New(lu.Uint128{Lo: q.Lo, Hi: q.Hi}, scale, negative && !q.IsZero()), inexact
}

func toRoundingMode(m RoundingMode) (jk.RoundingMode, bool) {
	switch m {
	case RoundingTowardZero:
		return jk.ROUND_TOWARD_ZERO, true
	case RoundingAwayFromZero:
		return jk.ROUND_AWAY_FROM_ZERO, true
	case RoundingFloor:
		return jk.ROUND_DOWN, true
	case RoundingCeil:
		return jk.ROUND_UP, true
	case RoundingHalfTowardZero:
		return jk.ROUND_HALF_TOWARD_ZERO, true
	case RoundingHalfAwayFromZero:
		return jk.ROUND_HALF_AWAY_FROM_ZERO, true
	case RoundingHalfEven:
		return jk.ROUND_BANK, true
	default:
		return 0, false
	}
}
func (d Dec128) RescaleRoundWithStatus(scale uint8, mode RoundingMode) (Dec128, bool) {
	if d.IsNaN() {
		return d, false
	}
	if scale > MaxPrecision {
		return NaN(le.PrecisionOutOfRange), false
	}
	if mode > RoundingHalfEven {
		return NaN(le.InvalidFormat), false
	}
	if scale == d.scale {
		return d, false
	}
	if d.IsZero() {
		return NewStoredZero(scale, d.negative()), false
	}
	if scale > d.scale {
		return d.Rescale(scale), false
	}
	// The discarded portion of a 128-bit coefficient fits in uint64 since
	// a scale reduction never divides by more than 10^19. This avoids the
	// generic 256-bit quotient path while preserving the full remainder.
	factor := ju.Pow10Uint64[d.scale-scale]
	quotient, remainder, st := d.coef.QuoRem64(factor)
	if st != state.OK {
		return NaN(fromState(st.Error())), false
	}
	if remainder == 0 {
		return Dec128{coef: quotient, scale: scale, state: signState(d.negative())}, false
	}
	negative := d.negative()
	increment := false
	switch mode {
	case RoundingAwayFromZero:
		increment = true
	case RoundingFloor:
		increment = negative
	case RoundingCeil:
		increment = !negative
	case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
		// Compare against factor-remainder, avoiding an overflow of 2*r.
		complement := factor - remainder
		increment = remainder > complement || remainder == complement && (mode == RoundingHalfAwayFromZero || mode == RoundingHalfEven && quotient.Lo&1 != 0)
	}
	if increment {
		var st state.State
		quotient, st = quotient.Add64(1)
		if st != state.OK {
			return NaN(le.Overflow), false
		}
	}
	// A nonzero negative rounded to zero becomes a positive stored zero.
	return Dec128{coef: quotient, scale: scale, state: signState(negative && !quotient.IsZero())}, true
}
func (d Dec128) RescaleRound(scale uint8, mode RoundingMode) Dec128 {
	r, _ := d.RescaleRoundWithStatus(scale, mode)
	return r
}
func (d Dec128) MulRound(b Dec128, p uint8, m RoundingMode) Dec128 {
	r, _ := d.MulRoundWithStatus(b, p, m)
	return r
}
func (d Dec128) DivRound(b Dec128, p uint8, m RoundingMode) Dec128 {
	r, _ := d.DivRoundWithStatus(b, p, m)
	return r
}

// roundExisting implements the older Round-family contract. It differs from
// RescaleRound: an equal/higher requested scale is a no-op (even if >19), and
// a negative magnitude rounded to zero retains its stored negative sign.
func (d Dec128) roundExisting(p uint8, mode RoundingMode) Dec128 {
	if d.IsNaN() || p >= d.scale {
		return d
	}
	divisor := ju.Pow10Uint64[d.scale-p]
	q, r, st := d.coef.QuoRem64(divisor)
	if st != state.OK {
		return NaN(fromState(st.Error()))
	}
	advance := false
	if r != 0 {
		switch mode {
		case RoundingAwayFromZero:
			advance = true
		case RoundingFloor:
			advance = d.negative()
		case RoundingCeil:
			advance = !d.negative()
		case RoundingHalfTowardZero:
			advance = r > divisor/2
		case RoundingHalfAwayFromZero:
			advance = r >= divisor/2
		case RoundingHalfEven:
			advance = r > divisor/2 || r == divisor/2 && q.Lo&1 == 1
		}
	}
	if advance {
		q, st = q.Add64(1)
		if st != state.OK {
			return NaN(fromState(st.Error()))
		}
	}
	return Dec128{coef: q, scale: p, state: signState(d.negative())}
}
func (d Dec128) Round(p uint8) Dec128             { return d.roundExisting(p, RoundingHalfAwayFromZero) }
func (d Dec128) RoundBank(p uint8) Dec128         { return d.roundExisting(p, RoundingHalfEven) }
func (d Dec128) Trunc(p uint8) Dec128             { return d.roundExisting(p, RoundingTowardZero) }
func (d Dec128) RoundDown(p uint8) Dec128         { return d.roundExisting(p, RoundingFloor) }
func (d Dec128) RoundUp(p uint8) Dec128           { return d.roundExisting(p, RoundingCeil) }
func (d Dec128) RoundTowardZero(p uint8) Dec128   { return d.roundExisting(p, RoundingTowardZero) }
func (d Dec128) RoundAwayFromZero(p uint8) Dec128 { return d.roundExisting(p, RoundingAwayFromZero) }
func (d Dec128) RoundHalfTowardZero(p uint8) Dec128 {
	return d.roundExisting(p, RoundingHalfTowardZero)
}
func (d Dec128) RoundHalfAwayFromZero(p uint8) Dec128 {
	return d.roundExisting(p, RoundingHalfAwayFromZero)
}

// SqrtRoundWithStatus uses upstream's explicit sqrt mode, never its process
// defaults. Exactness is checked by squaring the rounded coefficient using
// upstream uint128 primitives, avoiding a second full square-root operation.
// This check is independent of rounding mode and detects zero underflow.
func (d Dec128) SqrtRoundWithStatus(scale uint8, mode RoundingMode) (Dec128, bool) {
	if d.IsNaN() {
		return d, false
	}
	if scale > MaxPrecision {
		return NaN(le.PrecisionOutOfRange), false
	}
	m, ok := toRoundingMode(mode)
	if !ok {
		return NaN(le.InvalidFormat), false
	}
	if d.IsZero() {
		return New(lu.Zero, scale, false), false
	}
	if d.negative() {
		return NaN(le.SqrtNegative), false
	}
	u := d.upstream()
	v := u.SqrtRound(scale, m)
	if v.IsNaN() {
		return NaN(fromState(v.ErrorDetails())), false
	}
	result := fromUpstream(v)
	// v represents rootCoef / 10^scale. Compare its exact square against
	// the original coefficient at the common power-of-ten scale. Both
	// sides fit in 256 bits (128-bit coefficient times at most 10^38).
	rootSquaredLow, rootSquaredHigh := result.coef.MulCarry(result.coef)
	shift := 2*int(scale) - int(d.scale)
	var originalLow, originalHigh ju.Uint128
	if shift >= 0 {
		originalLow, originalHigh = d.coef.MulCarry(ju.Pow10Uint128[shift])
	} else {
		// For a coarser output, root^2 must be multiplied by 10^(-shift).
		// Do not truncate a 256-bit result: a nonzero high half already
		// implies inexactness because the input has only 128 bits.
		if !rootSquaredHigh.IsZero() {
			return result, true
		}
		rootSquaredLow, rootSquaredHigh = rootSquaredLow.MulCarry(ju.Pow10Uint128[-shift])
		originalLow = d.coef
	}
	return result, rootSquaredLow != originalLow || rootSquaredHigh != originalHigh
}
func (d Dec128) SqrtRound(scale uint8, mode RoundingMode) Dec128 {
	v, _ := d.SqrtRoundWithStatus(scale, mode)
	return v
}
func (d Dec128) Sqrt() Dec128 {
	if d.IsNaN() {
		return d
	}
	if d.IsZero() {
		return Zero
	}
	if d.negative() {
		return NaN(le.SqrtNegative)
	}
	if d.Equal(One) {
		return One
	}
	return d.SqrtRound(uint8(defaultPrecision.Load()), RoundingTowardZero)
}
