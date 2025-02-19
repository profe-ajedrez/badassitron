package dec128

import "github.com/profe-ajedrez/badassitron/dec128/errors"

// Add returns the sum of the Dec128 and the other Dec128.
// If any of the Dec128 is NaN, the result will be NaN.
// In case of overflow, the result will be NaN.
func (decimal Dec128) Add(other Dec128) Dec128 {
	if decimal.err != errors.None {
		return decimal
	}

	if other.err != errors.None {
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
	if decimal.err != errors.None {
		return decimal
	}

	if other.err != errors.None {
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
	if decimal.err != errors.None {
		return decimal
	}

	if other.err != errors.None {
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
	if decimal.err != errors.None {
		return decimal
	}

	if other.err != errors.None {
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

// Mod returns decimal % other.
// If any of the Dec128 is NaN, the result will be NaN.
// In case of overflow, underflow, or division by zero, the result will be NaN.
func (decimal Dec128) Mod(other Dec128) Dec128 {
	if decimal.err != errors.None {
		return decimal
	}

	if other.err != errors.None {
		return other
	}

	if other.IsZero() {
		return NaN(errors.DivisionByZero)
	}

	if decimal.IsZero() {
		return Zero
	}

	_, r, ok := decimal.tryQuoRem(other)
	if ok {
		return r
	}

	a := decimal.Canonical()
	b := other.Canonical()
	_, r, ok = a.tryQuoRem(b)
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

// QuoRem returns the quotient and remainder of the division of Dec128 by other Dec128.
// If any of the Dec128 is NaN, the result will be NaN.
// In case of overflow, underflow, or division by zero, the result will be NaN.
func (decimal Dec128) QuoRem(other Dec128) (Dec128, Dec128) {
	if decimal.err != errors.None {
		return decimal, decimal
	}

	if other.err != errors.None {
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
	if decimal.err != errors.None {
		return decimal
	}
	return Dec128{coef: decimal.coef, exp: decimal.exp}
}

// Neg returns -d
// If Dec128 is NaN, the result will be NaN.
func (decimal Dec128) Neg() Dec128 {
	if decimal.err != errors.None {
		return decimal
	}
	return Dec128{coef: decimal.coef, exp: decimal.exp, neg: !decimal.neg}
}

// Sqrt returns the square root of the Dec128.
// If Dec128 is NaN, the result will be NaN.
// If Dec128 is negative, the result will be NaN.
// In case of overflow, the result will be NaN.
func (decimal Dec128) Sqrt() Dec128 {
	if decimal.err != errors.None {
		return decimal
	}

	if decimal.IsZero() {
		return Zero
	}

	if decimal.neg {
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
	if decimal.err != errors.None {
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
