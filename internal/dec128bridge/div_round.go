package dec128bridge

import (
	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// DivAtPrecision preserves the historical Div calculation for a CALLER-SUPPLIED
// default precision. The legacy global defaultPrecision has no public getter;
// this bridge therefore deliberately does not claim to implement DivStrict.
// All multiplications and 256/128 divisions use imported Jokruger primitives.
// This is an internal migration probe, not a public replacement.
func DivAtPrecision(a, b legacy.Dec128, defaultPrecision uint8) legacy.Dec128 {
	if a.IsNaN() {
		return a
	}
	if b.IsNaN() {
		return b
	}
	if b.IsZero() {
		return legacy.NaN(le.DivisionByZero)
	}
	if a.IsZero() {
		return legacy.Zero
	}
	if defaultPrecision > legacy.MaxPrecision {
		return legacy.NaN(le.PrecisionOutOfRange)
	}
	if r, ok := divAttempt(a, b, defaultPrecision); ok {
		return r
	}
	if r, ok := divAttempt(canonicalOperand(a), canonicalOperand(b), defaultPrecision); ok {
		return r
	}
	return legacy.NaN(le.Overflow)
}

func divAttempt(a, b legacy.Dec128, target uint8) (legacy.Dec128, bool) {
	scale := a.Precision()
	factor := b.Precision()
	if scale < target {
		factor += target - scale
		scale = target
	}
	ac, bc := a.Coefficient(), b.Coefficient()
	u, carry := (ju.Uint128{Lo: ac.Lo, Hi: ac.Hi}).MulCarry(ju.Pow10Uint128[factor])
	q, _, st := ju.QuoRem256By128(u, carry, ju.Uint128{Lo: bc.Lo, Hi: bc.Hi})
	if st != state.OK {
		return legacy.Dec128{}, false
	}
	if q.IsZero() {
		return legacy.NewStoredZero(scale, a.StoredNegative() != b.StoredNegative()), true
	}
	return legacy.New(lu.Uint128{Lo: q.Lo, Hi: q.Hi}, scale, a.IsNegative() != b.IsNegative()), true
}
