package dec128bridge

import (
	jk "github.com/jokruger/dec128"
	ju "github.com/jokruger/dec128/uint128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// AddStrict computes the exact sum using Jokruger's explicit, global-free
// arithmetic. If alignment overflows, it retries the historical canonical
// operands rather than implicitly rounding away nonzero digits.
func AddStrict(a, b legacy.Dec128) legacy.Dec128 {
	return addSubStrict(a, b, false)
}

// SubStrict computes the exact difference, preserving the historical
// scale and NaN precedence. Neither operation reads upstream global policy.
func SubStrict(a, b legacy.Dec128) legacy.Dec128 {
	return addSubStrict(a, b, true)
}

func addSubStrict(a, b legacy.Dec128, sub bool) legacy.Dec128 {
	if a.IsNaN() {
		return a
	}
	if b.IsNaN() {
		return b
	}
	// Direct imported uint128 primitives for the common equal-scale case.
	// Preserve historical exact arithmetic and avoid the generic adapter.
	if a.Precision() == b.Precision() && a.IsNegative() == b.IsNegative() && a.StoredNegative() == b.StoredNegative() {
		ac, bc := a.Coefficient(), b.Coefficient()
		ua, ub := ju.Uint128{Lo: ac.Lo, Hi: ac.Hi}, ju.Uint128{Lo: bc.Lo, Hi: bc.Hi}
		if sub {
			if ac == bc {
				return legacy.Zero
			}
			if ac.Hi == 0 && ac.Lo == 0 || bc.Hi == 0 && bc.Lo == 0 {
				// Preserve mixed signed-zero behavior through the generic path.
			} else {
				v, borrow := ua.SubBorrow(ub)
				neg := a.IsNegative()
				if borrow != 0 {
					v, _ = (ju.Uint128{}).SubBorrow(v)
					neg = !neg
				}
				return legacy.New(lu.Uint128{Lo: v.Lo, Hi: v.Hi}, a.Precision(), neg)
			}
		} else if !a.IsZero() && !b.IsZero() {
			v, carry := ua.AddCarry(ub)
			if carry == 0 {
				return legacy.New(lu.Uint128{Lo: v.Lo, Hi: v.Hi}, a.Precision(), a.IsNegative())
			}
		}
	}
	op := func(x, y legacy.Dec128) legacy.Dec128 {
		scale := x.Precision()
		if y.Precision() > scale {
			scale = y.Precision()
		}
		ux, uy := ToUpstream(x), ToUpstream(y)
		// Historical alignment rejects a scaled operand that exceeds 128 bits,
		// even when the wider exact sum would fit after cancellation.
		if ux.Rescale(scale).IsNaN() || uy.Rescale(scale).IsNaN() {
			return legacy.NaN(le.Overflow)
		}
		var v jk.Dec128
		if sub {
			v = ux.SubRound(uy, scale, jk.ROUND_NAN)
		} else {
			v = ux.AddRound(uy, scale, jk.ROUND_NAN)
		}
		if v.IsNaN() {
			return legacy.NaN(le.Overflow)
		}
		// The historical same-sign zero path retains both stored scale
		// and sign. Opposite-sign cancellation is canonical positive Zero.
		if v.IsZero() {
			leftNeg, rightNeg := x.StoredNegative(), y.StoredNegative()
			if sub {
				if leftNeg != rightNeg {
					return legacyFiniteZero(scale, leftNeg)
				}
			} else if leftNeg == rightNeg {
				return legacyFiniteZero(scale, leftNeg)
			}
			return legacy.Zero
		}
		return FromUpstream(v)
	}
	res := op(a, b)
	if !res.IsNaN() {
		return res
	}
	res = op(canonicalOperand(a), canonicalOperand(b))
	if !res.IsNaN() {
		return res
	}
	return legacy.NaN(le.Overflow)
}
