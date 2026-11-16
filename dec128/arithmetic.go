package dec128

import (
	"math/bits"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// Add returns the sum of the Dec128 and the other Dec128.
// If any of the Dec128 is NaN, the result will be NaN.
// In case of overflow, the result will be NaN.
func (decimal Dec128) Add(other Dec128) Dec128 {
	// The common case needs no scale alignment. Keep the existing slow
	// path for overflow, including its canonicalization fallback.
	if decimal.exp == other.exp && decimal.state == other.state && decimal.state < 2 {
		lo, carry := bits.Add64(decimal.coef.Lo, other.coef.Lo, 0)
		hi, carry := bits.Add64(decimal.coef.Hi, other.coef.Hi, carry)
		if carry == 0 {
			return Dec128{
				coef: uint128.Uint128{Lo: lo, Hi: hi},
				exp:  decimal.exp, state: decimal.state,
			}
		}
	}

	if decimal.errorCode() != errors.None {
		return decimal
	}

	if other.errorCode() != errors.None {
		return other
	}

	r, ok := decimal.tryAdd(other)
	if ok {
		return r
	}

	a := decimal.Canonical()
	b := other.Canonical()
	r, ok = a.tryAdd(b)
	if ok {
		return r
	}

	return NaN(errors.Overflow)
}

// AddInt returns the sum of the Dec128 and the int.
// If Dec128 is NaN, the result will be NaN.
// In case of overflow, the result will be NaN.
func (decimal Dec128) AddInt(other int) Dec128 {
	return decimal.Add(FromInt(other))
}

// Sub returns the difference of the Dec128 and the other Dec128.
// If any of the Dec128 is NaN, the result will be NaN.
// In case of overflow/underflow, the result will be NaN.
func (decimal Dec128) Sub(other Dec128) Dec128 {
	// Subtract full-width coefficients once. If the subtraction borrows,
	// its result is the two's-complement negative magnitude. Negate that
	// magnitude and reverse the decimal sign. Exact cancellation returns Zero.
	if decimal.exp == other.exp && decimal.state == other.state && decimal.state < 2 {
		lo, borrow := bits.Sub64(decimal.coef.Lo, other.coef.Lo, 0)
		hi, borrow := bits.Sub64(decimal.coef.Hi, other.coef.Hi, borrow)
		state := decimal.state
		if borrow != 0 {
			var carry uint64
			lo, carry = bits.Add64(^lo, 1, 0)
			hi, _ = bits.Add64(^hi, 0, carry)
			state ^= 1
		}
		if lo == 0 && hi == 0 {
			return Zero
		}
		return Dec128{
			coef: uint128.Uint128{Lo: lo, Hi: hi},
			exp:  decimal.exp, state: state,
		}
	}

	if decimal.errorCode() != errors.None {
		return decimal
	}

	if other.errorCode() != errors.None {
		return other
	}

	r, ok := decimal.trySub(other)
	if ok {
		return r
	}

	a := decimal.Canonical()
	b := other.Canonical()
	r, ok = a.trySub(b)
	if ok {
		return r
	}

	return NaN(errors.Overflow)
}

// SubInt returns the difference of the Dec128 and the int.
// If Dec128 is NaN, the result will be NaN.
// In case of overflow/underflow, the result will be NaN.
func (decimal Dec128) SubInt(other int) Dec128 {
	return decimal.Sub(FromInt(other))
}

// Mul returns decimal * other.
// If any of the Dec128 is NaN, the result will be NaN.
// In case of overflow, the result will be NaN.
func (decimal Dec128) Mul(other Dec128) Dec128 {
	// For two 64-bit coefficients the complete product fits in 128 bits.
	// The general path remains responsible for excess scale and recovery.
	if decimal.state < 2 && other.state < 2 &&
		decimal.coef.Hi == 0 && other.coef.Hi == 0 &&
		decimal.coef.Lo != 0 && other.coef.Lo != 0 &&
		decimal.exp+other.exp <= MaxPrecision {
		hi, lo := bits.Mul64(decimal.coef.Lo, other.coef.Lo)
		return Dec128{
			coef:  uint128.Uint128{Lo: lo, Hi: hi},
			exp:   decimal.exp + other.exp,
			state: decimal.state ^ other.state,
		}
	}

	if decimal.errorCode() != errors.None {
		return decimal
	}

	if other.errorCode() != errors.None {
		return other
	}

	if decimal.IsZero() || other.IsZero() {
		return Zero
	}

	r, ok := decimal.tryMul(other)
	if ok {
		return r
	}

	a := decimal.Canonical()
	b := other.Canonical()
	r, ok = a.tryMul(b)
	if ok {
		return r
	}

	return NaN(errors.Overflow)
}

// MulInt returns decimal * other.
// If Dec128 is NaN, the result will be NaN.
// In case of overflow, the result will be NaN.
func (decimal Dec128) MulInt(other int) Dec128 {
	return decimal.Mul(FromInt(other))
}

// Div returns decimal / other.
// If any of the Dec128 is NaN, the result will be NaN.
// In case of overflow, underflow, or division by zero, the result will be NaN.
func (decimal Dec128) Div(other Dec128) Dec128 {
	if decimal.errorCode() != errors.None {
		return decimal
	}

	if other.errorCode() != errors.None {
		return other
	}

	if other.IsZero() {
		return NaN(errors.DivisionByZero)
	}

	if decimal.IsZero() {
		return Zero
	}

	r, ok := decimal.tryDiv(other)
	if ok {
		return r
	}

	a := decimal.Canonical()
	b := other.Canonical()
	r, ok = a.tryDiv(b)
	if ok {
		return r
	}

	return NaN(errors.Overflow)
}

// DivInt returns decimal / other.
// If Dec128 is NaN, the result will be NaN.
// In case of overflow, underflow, or division by zero, the result will be NaN.
func (decimal Dec128) DivInt(other int) Dec128 {
	return decimal.Div(FromInt(other))
}

// Mod returns the exact remainder of division truncated toward zero.
// The remainder retains the dividend's sign, including signed zero.
// It can succeed even if the integer quotient would overflow a Dec128.
// NaN operands or division by zero return NaN.
func (decimal Dec128) Mod(other Dec128) Dec128 {
	if decimal.errorCode() != errors.None {
		return decimal
	}

	if other.errorCode() != errors.None {
		return other
	}

	if other.IsZero() {
		return NaN(errors.DivisionByZero)
	}

	if decimal.IsZero() {
		return Zero
	}

	r, ok := decimal.tryMod(other)
	if ok {
		return r
	}

	a := decimal.Canonical()
	b := other.Canonical()
	r, ok = a.tryMod(b)
	if ok {
		return r
	}

	return NaN(errors.Overflow)
}

// ModInt returns decimal % other.
// If Dec128 is NaN, the result will be NaN.
// In case of overflow, underflow, or division by zero, the result will be NaN.
func (decimal Dec128) ModInt(other int) Dec128 {
	return decimal.Mod(FromInt(other))
}

// QuoRem returns the integer quotient truncated toward zero and its exact remainder.
// The quotient is unscaled and must fit in 128 bits; an unrepresentable
// quotient makes both results NaN, even when Mod could return the remainder.
// NaN operands or division by zero also return NaN.
func (decimal Dec128) QuoRem(other Dec128) (Dec128, Dec128) {
	if decimal.errorCode() != errors.None {
		return decimal, decimal
	}

	if other.errorCode() != errors.None {
		return other, other
	}

	if other.IsZero() {
		return NaN(errors.DivisionByZero), NaN(errors.DivisionByZero)
	}

	if decimal.IsZero() {
		return Zero, Zero
	}

	q, r, ok := decimal.tryQuoRem(other)
	if ok {
		return q, r
	}

	a := decimal.Canonical()
	b := other.Canonical()
	q, r, ok = a.tryQuoRem(b)
	if ok {
		return q, r
	}

	return NaN(errors.Overflow), NaN(errors.Overflow)
}

// QuoRemInt returns the quotient and remainder of the division of Dec128 by int.
// If Dec128 is NaN, the result will be NaN.
// In case of overflow, underflow, or division by zero, the result will be NaN.
func (decimal Dec128) QuoRemInt(other int) (Dec128, Dec128) {
	return decimal.QuoRem(FromInt(other))
}

// Abs returns |d|
// If Dec128 is NaN, the result will be NaN.
func (decimal Dec128) Abs() Dec128 {
	if decimal.errorCode() != errors.None {
		return decimal
	}
	return Dec128{coef: decimal.coef, exp: decimal.exp}
}

// Neg returns -d
// If Dec128 is NaN, the result will be NaN.
func (decimal Dec128) Neg() Dec128 {
	if decimal.errorCode() != errors.None {
		return decimal
	}
	return Dec128{coef: decimal.coef, exp: decimal.exp, state: signState(!decimal.negative())}
}

// Sqrt returns the square root of the Dec128.
// If Dec128 is NaN, the result will be NaN.
// If Dec128 is negative, the result will be NaN.
// In case of overflow, the result will be NaN.
func (decimal Dec128) Sqrt() Dec128 {
	if decimal.errorCode() != errors.None {
		return decimal
	}

	if decimal.IsZero() {
		return Zero
	}

	if decimal.negative() {
		return NaN(errors.SqrtNegative)
	}

	if decimal.Equal(One) {
		return One
	}

	r, ok := decimal.trySqrt()
	if ok {
		return r
	}

	a := decimal.Canonical()
	r, ok = a.trySqrt()
	if ok {
		return r
	}

	return NaN(errors.Overflow)
}

// PowInt returns Dec128 raised to the power of n.
func (decimal Dec128) PowInt(n int) Dec128 {
	if decimal.errorCode() != errors.None {
		return decimal
	}

	if n < 0 {
		return One.Div(decimal.PowInt(-n))
	}

	if n == 0 {
		return One
	}

	if n == 1 {
		return decimal
	}

	if (n & 1) == 0 {
		return decimal.Mul(decimal).PowInt(n / 2)
	}

	return decimal.Mul(decimal).PowInt((n - 1) / 2).Mul(decimal)
}
