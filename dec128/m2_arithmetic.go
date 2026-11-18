package dec128

import (
	jk "github.com/jokruger/dec128"
	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
	"math/bits"
)

// Mul preserves the historical exact-or-NaN multiplication contract.
// Its integer arithmetic comes from the imported Jokruger module. In
// particular, the 256-bit product and divide are not copied into this repo.
// Unlike upstream Mul, it does not read any process-wide rounding setting.
func (a Dec128) Mul(b Dec128) Dec128 {
	// Match the frozen 64-bit exact path before entering the 256-bit engine.
	if a.state < 2 && b.state < 2 && a.coef.Hi == 0 && b.coef.Hi == 0 && a.coef.Lo != 0 && b.coef.Lo != 0 && a.scale+b.scale <= MaxPrecision {
		hi, lo := bits.Mul64(a.coef.Lo, b.coef.Lo)
		return Dec128{coef: ju.Uint128{Lo: lo, Hi: hi}, scale: a.scale + b.scale, state: a.state ^ b.state}
	}
	if a.IsNaN() {
		return a
	}
	if b.IsNaN() {
		return b
	}
	if a.IsZero() || b.IsZero() {
		return Zero
	}
	if out, ok := mulExact(a, b); ok {
		return out
	}
	// Canonicalization can only help if one operand loses trailing zeros.
	// Never repeat an identical 256-bit multiplication and quotient attempt.
	ca, cb := a.Canonical(), b.Canonical()
	if ca.scale != a.scale || cb.scale != b.scale {
		if out, ok := mulExact(ca, cb); ok {
			return out
		}
	}
	return NaN(le.Overflow)
}

func mulExact(a, b Dec128) (Dec128, bool) {
	scale := a.Precision() + b.Precision() // 0..38
	lo, hi := a.coef.MulCarry(b.coef)
	neg := a.negative() != b.negative()
	result := func(c ju.Uint128, s uint8) Dec128 {
		return Dec128{coef: c, scale: s, state: signState(neg)}
	}

	if hi.IsZero() {
		if scale <= jk.MaxScale {
			return result(lo, scale), true
		}
		// The legacy implementation canonicalizes the entire product when
		// its scale exceeds the maximum, rather than merely limiting to 19.
		for scale > 0 {
			q, r, st := lo.QuoRem64(10)
			if st != state.OK || r != 0 {
				break
			}
			lo = q
			scale--
		}
		if scale <= jk.MaxScale {
			return result(lo, scale), true
		}
		return Dec128{}, false
	}

	// When the full product does not fit in 128 bits, Badassitron chooses
	// the *largest exact decimal reduction* first, not necessarily the
	// maximum retained scale. Preserve that observable scale contract.
	for drop := scale; drop > 0; drop-- {
		q, r, st := ju.QuoRem256By128(lo, hi, ju.Pow10Uint128[drop])
		if st == state.OK && r.IsZero() {
			return result(q, scale-drop), true
		}
		if st == state.Overflow {
			return Dec128{}, false
		}
		if scale-drop > jk.MaxScale {
			return Dec128{}, false
		}
	}
	return Dec128{}, false
}

// Add computes the exact sum using Jokruger's explicit, global-free
// arithmetic. If alignment overflows, it retries the historical canonical
// operands rather than implicitly rounding away nonzero digits.
func (a Dec128) Add(b Dec128) Dec128 {
	if a.scale == b.scale && a.state == b.state && a.state < 2 {
		lo, carry := bits.Add64(a.coef.Lo, b.coef.Lo, 0)
		hi, carry := bits.Add64(a.coef.Hi, b.coef.Hi, carry)
		if carry == 0 {
			return Dec128{coef: ju.Uint128{Lo: lo, Hi: hi}, scale: a.scale, state: signState(a.negative())}
		}
	}
	return addSubStrict(a, b, false)
}

// Sub computes the exact difference, preserving the historical
// scale and NaN precedence. Neither operation reads upstream global policy.
func (a Dec128) Sub(b Dec128) Dec128 {
	if a.scale == b.scale && a.state == b.state && a.state < 2 {
		lo, borrow := bits.Sub64(a.coef.Lo, b.coef.Lo, 0)
		hi, borrow := bits.Sub64(a.coef.Hi, b.coef.Hi, borrow)
		neg := a.negative()
		if borrow != 0 {
			var carry uint64
			lo, carry = bits.Add64(^lo, 1, 0)
			hi, _ = bits.Add64(^hi, 0, carry)
			neg = !neg
		}
		if lo == 0 && hi == 0 {
			return Zero
		}
		return Dec128{coef: ju.Uint128{Lo: lo, Hi: hi}, scale: a.scale, state: signState(neg)}
	}
	return addSubStrict(a, b, true)
}

// addSubStrict reproduces the historical exact scale-alignment contract.
// Coefficient multiplication, addition and subtraction use exported Jokruger
// uint128 primitives; arithmetic must never pick up upstream global rounding.
func addSubStrict(a, b Dec128, sub bool) Dec128 {
	if a.IsNaN() {
		return a
	}
	if b.IsNaN() {
		return b
	}
	if result, ok := alignedAddSub(a, b, sub); ok {
		return result
	}
	// Reducing exact trailing decimal zeros can make otherwise overflowing
	// alignment representable. Preserve the legacy retry ordering.
	ca, cb := a.Canonical(), b.Canonical()
	if ca.scale != a.scale || cb.scale != b.scale {
		if result, ok := alignedAddSub(ca, cb, sub); ok {
			return result
		}
	}
	return NaN(le.Overflow)
}

func alignedAddSub(a, b Dec128, sub bool) (Dec128, bool) {
	scale := a.scale
	if b.scale > scale {
		scale = b.scale
	}
	ua, ub := a.coef, b.coef
	if a.scale < scale {
		var st state.State
		ua, st = ua.Mul(ju.Pow10Uint128[scale-a.scale])
		if st != state.OK {
			return Dec128{}, false
		}
	}
	if b.scale < scale {
		var st state.State
		ub, st = ub.Mul(ju.Pow10Uint128[scale-b.scale])
		if st != state.OK {
			return Dec128{}, false
		}
	}
	aNeg, bNeg := a.negative(), b.negative()
	if sub {
		bNeg = !bNeg
	}
	if aNeg == bNeg {
		sum, carry := ua.AddCarry(ub)
		if carry != 0 {
			return Dec128{}, false
		}
		return Dec128{coef: sum, scale: scale, state: signState(aNeg)}, true
	}
	switch cmp := ua.Compare(ub); {
	case cmp == 0:
		return Zero, true
	case cmp > 0:
		diff, _ := ua.SubBorrow(ub)
		return Dec128{coef: diff, scale: scale, state: signState(aNeg)}, true
	default:
		diff, _ := ub.SubBorrow(ua)
		return Dec128{coef: diff, scale: scale, state: signState(bNeg)}, true
	}
}

// DivAtPrecision computes historical truncated division at an explicit scale.
// Div reads the candidate-local atomic default, never upstream global state.
func (a Dec128) DivAtPrecision(b Dec128, defaultPrecision uint8) Dec128 {
	if a.IsNaN() {
		return a
	}
	if b.IsNaN() {
		return b
	}
	if b.IsZero() {
		return NaN(le.DivisionByZero)
	}
	if a.IsZero() {
		return Zero
	}
	if defaultPrecision > MaxPrecision {
		return NaN(le.PrecisionOutOfRange)
	}
	if r, ok := divAttempt(a, b, defaultPrecision); ok {
		return r
	}
	if r, ok := divAttempt(a.Canonical(), b.Canonical(), defaultPrecision); ok {
		return r
	}
	return NaN(le.Overflow)
}

func divAttempt(a, b Dec128, target uint8) (Dec128, bool) {
	scale := a.Precision()
	factor := b.Precision()
	if scale < target {
		factor += target - scale
		scale = target
	}
	u, carry := a.coef.MulCarry(ju.Pow10Uint128[factor])
	q, _, st := ju.QuoRem256By128(u, carry, b.coef)
	if st != state.OK {
		return Dec128{}, false
	}
	return Dec128{coef: q, scale: scale, state: signState(a.negative() != b.negative())}, true
}

func (d Dec128) Div(b Dec128) Dec128 { return d.DivAtPrecision(b, uint8(defaultPrecision.Load())) }
func (d Dec128) AddInt(i int) Dec128 { return d.Add(FromInt(i)) }
func (d Dec128) SubInt(i int) Dec128 { return d.Sub(FromInt(i)) }
func (d Dec128) MulInt(i int) Dec128 { return d.Mul(FromInt(i)) }
func (d Dec128) DivInt(i int) Dec128 { return d.Div(FromInt(i)) }
