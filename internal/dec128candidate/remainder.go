package dec128candidate

import (
	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func magnitudeCompare(a, b Dec128) int {
	return a.Abs().Compare(b.Abs())
}

// QuoRem returns a 128-bit integral quotient and a full-precision remainder.
// A nonrepresentable quotient invalidates both values, unlike Mod.
func (a Dec128) QuoRem(b Dec128) (Dec128, Dec128) {
	if a.IsNaN() {
		return a, a
	}
	if b.IsNaN() {
		return b, b
	}
	if b.IsZero() {
		v := NaN(le.DivisionByZero)
		return v, v
	}
	if a.IsZero() {
		return Zero, Zero
	}
	if q, r, ok := quoRemAttempt(a, b); ok {
		return q, r
	}
	if q, r, ok := quoRemAttempt(a.Canonical(), b.Canonical()); ok {
		return q, r
	}
	v := NaN(le.Overflow)
	return v, v
}
func quoRemAttempt(a, b Dec128) (Dec128, Dec128, bool) {
	if magnitudeCompare(a, b) < 0 {
		return Zero, a, true
	}
	target := a.scale
	if b.scale > target {
		target = b.scale
	}
	numeratorLow, numeratorHigh := a.coef.MulCarry(ju.Pow10Uint128[target-a.scale])
	divisor, st := b.coef.Mul(ju.Pow10Uint128[target-b.scale])
	if st != state.OK {
		return Dec128{}, Dec128{}, false
	}
	quotient, remainder, st := ju.QuoRem256By128(numeratorLow, numeratorHigh, divisor)
	if st != state.OK {
		return Dec128{}, Dec128{}, false
	}
	return Dec128{coef: quotient, state: signState(a.negative() != b.negative())}, Dec128{coef: remainder, scale: target, state: signState(a.negative())}, true
}
func (a Dec128) Mod(b Dec128) Dec128 {
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
	if r, ok := modAttempt(a, b); ok {
		return r
	}
	if r, ok := modAttempt(a.Canonical(), b.Canonical()); ok {
		return r
	}
	return NaN(le.Overflow)
}
func modAttempt(a, b Dec128) (Dec128, bool) {
	if magnitudeCompare(a, b) < 0 {
		return a, true
	}
	target := a.scale
	if b.scale > target {
		target = b.scale
	}
	low, high := a.coef.MulCarry(ju.Pow10Uint128[target-a.scale])
	divisor, st := b.coef.Mul(ju.Pow10Uint128[target-b.scale])
	if st != state.OK {
		return Dec128{}, false
	}
	if high.Compare(divisor) >= 0 {
		high, st = high.Mod(divisor)
		if st != state.OK {
			return Dec128{}, false
		}
	}
	_, r, st := ju.QuoRem256By128(low, high, divisor)
	if st != state.OK {
		return Dec128{}, false
	}
	return Dec128{coef: r, scale: target, state: signState(a.negative())}, true
}
func (a Dec128) PowInt(n int) Dec128 {
	if a.IsNaN() {
		return a
	}
	if n < 0 {
		return One.Div(a.PowInt(-(n + 1)).Mul(a))
	}
	if n == 0 {
		return One
	}
	if n == 1 {
		return a
	}
	if n&1 == 0 {
		return a.Mul(a).PowInt(n / 2)
	}
	return a.Mul(a).PowInt((n - 1) / 2).Mul(a)
}
func (a Dec128) ModInt(n int) Dec128              { return a.Mod(FromInt(n)) }
func (a Dec128) QuoRemInt(n int) (Dec128, Dec128) { return a.QuoRem(FromInt(n)) }
func (a Dec128) EncodeToUint128(scale uint8) (lu.Uint128, error) {
	if a.StoredNegative() {
		return lu.Zero, le.Negative.Value()
	}
	res := a.Rescale(scale)
	if res.IsNaN() {
		return lu.Zero, res.ErrorDetails()
	}
	return lu.Uint128{Lo: res.coef.Lo, Hi: res.coef.Hi}, nil
}
func (a Dec128) EncodeToUint64(scale uint8) (uint64, error) {
	c, err := a.EncodeToUint128(scale)
	if err != nil {
		return 0, err
	}
	if c.Hi != 0 {
		return 0, le.Overflow.Value()
	}
	return c.Lo, nil
}
