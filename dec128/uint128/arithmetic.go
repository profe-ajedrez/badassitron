package uint128

import (
	"math/bits"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
)

// Add returns uint128 + other and an error if the result overflows.
func (uint128 Uint128) Add(other Uint128) (Uint128, errors.Error) {
	lo, carry := bits.Add64(uint128.Lo, other.Lo, 0)
	hi, carry := bits.Add64(uint128.Hi, other.Hi, carry)

	if carry != 0 {
		return Zero, errors.Overflow
	}

	return Uint128{lo, hi}, errors.None
}

// Add64 returns uint128 + other and an error if the result overflows.
func (uint128 Uint128) Add64(other uint64) (Uint128, errors.Error) {
	lo, carry := bits.Add64(uint128.Lo, other, 0)
	hi, carry := bits.Add64(uint128.Hi, 0, carry)

	if carry != 0 {
		return Zero, errors.Overflow
	}

	return Uint128{lo, hi}, errors.None
}

// Sub returns uint128 - other and an error if the result underflows.
func (uint128 Uint128) Sub(other Uint128) (Uint128, errors.Error) {
	lo, borrow := bits.Sub64(uint128.Lo, other.Lo, 0)
	hi, borrow := bits.Sub64(uint128.Hi, other.Hi, borrow)

	if borrow != 0 {
		return Zero, errors.Underflow
	}

	return Uint128{lo, hi}, errors.None
}

// Sub64 returns uint128 - other and an error if the result underflows.
func (uint128 Uint128) Sub64(other uint64) (Uint128, errors.Error) {
	lo, borrow := bits.Sub64(uint128.Lo, other, 0)
	hi, borrow := bits.Sub64(uint128.Hi, 0, borrow)

	if borrow != 0 {
		return Zero, errors.Underflow
	}

	return Uint128{lo, hi}, errors.None
}

// Mul returns uint128 * other and an error if the result overflows.
func (uint128 Uint128) Mul(other Uint128) (Uint128, errors.Error) {
	hi, lo := bits.Mul64(uint128.Lo, other.Lo)
	p0, p1 := bits.Mul64(uint128.Hi, other.Lo)
	p2, p3 := bits.Mul64(uint128.Lo, other.Hi)
	hi, c0 := bits.Add64(hi, p1, 0)
	hi, c1 := bits.Add64(hi, p3, c0)

	if (uint128.Hi != 0 && other.Hi != 0) || p0 != 0 || p2 != 0 || c1 != 0 {
		return Zero, errors.Overflow
	}

	return Uint128{lo, hi}, errors.None
}

// MulCarry returns uint128 * other and carry.
func (uint128 Uint128) MulCarry(other Uint128) (Uint128, Uint128) {
	if uint128.Hi == 0 && other.Hi == 0 {
		hi, lo := bits.Mul64(uint128.Lo, other.Lo)
		return Uint128{Lo: lo, Hi: hi}, Zero
	}

	hi, lo := bits.Mul64(uint128.Lo, other.Lo)
	p0, p1 := bits.Mul64(uint128.Hi, other.Lo)
	p2, p3 := bits.Mul64(uint128.Lo, other.Hi)

	// calculate hi + p1 + p3
	// total carry = carry(hi+p1) + carry(hi+p1+p3)
	hi, c0 := bits.Add64(hi, p1, 0)
	hi, c1 := bits.Add64(hi, p3, 0)
	c1 += c0

	// calculate upper part of out carry
	e0, e1 := bits.Mul64(uint128.Hi, other.Hi)
	d, d0 := bits.Add64(p0, p2, 0)
	d, d1 := bits.Add64(d, c1, 0)
	e2, e3 := bits.Add64(d, e1, 0)

	return Uint128{Lo: lo, Hi: hi}, Uint128{Lo: e2, Hi: e0 + d0 + d1 + e3}
}

// Mul64 returns uint128 * other and an error if the result overflows.
func (uint128 Uint128) Mul64(other uint64) (Uint128, errors.Error) {
	hi, lo := bits.Mul64(uint128.Lo, other)
	p0, p1 := bits.Mul64(uint128.Hi, other)
	hi, c0 := bits.Add64(hi, p1, 0)

	if p0 != 0 || c0 != 0 {
		return Zero, errors.Overflow
	}

	return Uint128{lo, hi}, errors.None
}

// Div returns uint128 / other and an error if the divisor is zero.
func (uint128 Uint128) Div(other Uint128) (Uint128, errors.Error) {
	q, _, err := uint128.QuoRem(other)
	return q, err
}

// Div64 returns uint128 / other and an error if the divisor is zero.
func (uint128 Uint128) Div64(other uint64) (Uint128, errors.Error) {
	q, _, err := uint128.QuoRem64(other)
	return q, err
}

// Mod returns uint128 % other and an error if the divisor is zero.
func (uint128 Uint128) Mod(other Uint128) (Uint128, errors.Error) {
	_, r, err := uint128.QuoRem(other)
	return r, err
}

// Mod64 returns uint128 % other and an error if the divisor is zero.
func (uint128 Uint128) Mod64(other uint64) (uint64, errors.Error) {
	_, r, err := uint128.QuoRem64(other)
	return r, err
}

// QuoRem returns uint128 / other and uint128 % other and an error if the divisor is zero.
func (uint128 Uint128) QuoRem(other Uint128) (Uint128, Uint128, errors.Error) {
	if other.IsZero() {
		return Zero, Zero, errors.DivisionByZero
	}

	var q Uint128
	var r Uint128
	var err errors.Error

	if other.Hi == 0 {
		var r64 uint64
		q, r64, err = uint128.QuoRem64(other.Lo)
		if err != errors.None {
			return Zero, Zero, err
		}
		r = FromUint64(r64)
	} else {
		n := uint(bits.LeadingZeros64(other.Hi))
		v1 := other.Lsh(n)
		u1 := uint128.Rsh(1)
		tq, _ := bits.Div64(u1.Hi, u1.Lo, v1.Hi)
		tq >>= 63 - n
		if tq != 0 {
			tq--
		}
		q = FromUint64(tq)
		var m Uint128
		m, err = other.Mul64(tq)
		if err != errors.None {
			return Zero, Zero, err
		}
		r, err = uint128.Sub(m)
		if err != errors.None {
			return Zero, Zero, err
		}
		if r.Compare(other) >= 0 {
			q, err = q.Add64(1)
			if err != errors.None {
				return Zero, Zero, err
			}
			r, err = r.Sub(other)
			if err != errors.None {
				return Zero, Zero, err
			}
		}
	}

	return q, r, errors.None
}

// QuoRem64 returns uint128 / other and uint128 % other and an error if the divisor is zero.
func (uint128 Uint128) QuoRem64(other uint64) (Uint128, uint64, errors.Error) {
	if other == 0 {
		return Zero, 0, errors.DivisionByZero
	}

	var q Uint128
	var r uint64

	if uint128.Hi < other {
		q.Lo, r = bits.Div64(uint128.Hi, uint128.Lo, other)
	} else {
		q.Hi, r = bits.Div64(0, uint128.Hi, other)
		q.Lo, r = bits.Div64(r, uint128.Lo, other)
	}

	return q, r, errors.None
}
