package optimus

// Fixed-width temporaries for MulRoundWithStatus. The exact product of two
// Optimus coefficients occupies at most 16 uint64 limbs (1024 bits).
// Upscaling only needs 8 significant limbs: anything larger overflows the
// public 512-bit coefficient. No 4096-bit intermediate is needed.

import (
	"math/bits"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec256/uint256"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
	"github.com/profe-ajedrez/badassitron/internal/uintwide"
)

const fixed01Pow19 = uint64(10_000_000_000_000_000_000)

// fixed01Power10 provides the short decimal divisor/multiplier on the stack.
// A computed local avoids introducing mutable package-wide power tables.
func fixed01Power10(n int) uint64 {
	p := uint64(1)
	for n > 0 {
		p *= 10
		n--
	}
	return p
}

// fixed01Words reads the immutable coefficient directly, without converting
// it to math/big, a string, or a Dec1024 value.
func fixed01Words(d Optimus) (w [8]uint64, n int) {
	if d.wide == nil {
		w[0], w[1] = d.small.Lo, d.small.Hi
		if w[1] != 0 {
			return w, 2
		}
		if w[0] != 0 {
			return w, 1
		}
		return w, 0
	}
	c := *d.wide
	w = [8]uint64{
		c.Lo.Lo.Lo, c.Lo.Lo.Hi, c.Lo.Hi.Lo, c.Lo.Hi.Hi,
		c.Hi.Lo.Lo, c.Hi.Lo.Hi, c.Hi.Hi.Lo, c.Hi.Hi.Hi,
	}
	n = 8
	for n > 0 && w[n-1] == 0 {
		n--
	}
	return
}

func fixed01Trim(v *[16]uint64, n int) int {
	for n > 0 && v[n-1] == 0 {
		n--
	}
	return n
}

// fixed01DivChunk divides in place over significant limbs only.
func fixed01DivChunk(v *[16]uint64, n int, div uint64) (int, uint64) {
	var rem uint64
	for i := n - 1; i >= 0; i-- {
		v[i], rem = bits.Div64(rem, v[i], div)
	}
	return fixed01Trim(v, n), rem
}

// fixed01MulChunk grows a coefficient by a power of ten. The result's final
// limit is 512 bits, so exceeding 8 limbs is an immediate overflow.
func fixed01MulChunk(v *[16]uint64, n int, mul uint64) (int, bool) {
	var carry uint64
	for i := 0; i < n; i++ {
		hi, lo := bits.Mul64(v[i], mul)
		var c uint64
		v[i], c = bits.Add64(lo, carry, 0)
		carry = hi + c
	}
	if carry != 0 {
		if n == 8 {
			return n, false
		}
		v[n] = carry
		n++
	}
	return n, true
}

func fixed01Pack(v *[16]uint64, n int, scale uint8, neg bool) Optimus {
	if n <= 2 {
		return NewFromUint128(uint128.Uint128{Lo: v[0], Hi: v[1]}, scale, neg)
	}
	c := uint512.Uint512{
		Lo: uint256.Uint256{
			Lo: uint128.Uint128{Lo: v[0], Hi: v[1]},
			Hi: uint128.Uint128{Lo: v[2], Hi: v[3]},
		},
		Hi: uint256.Uint256{
			Lo: uint128.Uint128{Lo: v[4], Hi: v[5]},
			Hi: uint128.Uint128{Lo: v[6], Hi: v[7]},
		},
	}
	return New(c, scale, neg)
}

func (d Optimus) mulRoundFixed01(b Optimus, p uint8, m RoundingMode) (Optimus, bool) {
	left, na := fixed01Words(d)
	right, nb := fixed01Words(b)
	var product [16]uint64
	n := na + nb
	if na != 0 && nb != 0 {
		uintwide.MulFull(product[:n], left[:na], right[:nb])
		n = fixed01Trim(&product, n)
	} else {
		n = 0
	}
	scale := int(d.exp) + int(b.exp)
	target := int(p)
	neg := d.neg != b.neg
	if scale <= target {
		// There is no discarded remainder in this branch.
		if n > 8 {
			return NaN(errors.Overflow), false
		}
		shift := target - scale
		for shift >= 19 {
			var ok bool
			n, ok = fixed01MulChunk(&product, n, fixed01Pow19)
			if !ok {
				return NaN(errors.Overflow), false
			}
			shift -= 19
		}
		if shift != 0 {
			var ok bool
			n, ok = fixed01MulChunk(&product, n, fixed01Power10(shift))
			if !ok {
				return NaN(errors.Overflow), false
			}
		}
		return fixed01Pack(&product, n, p, neg), false
	}

	shift := scale - target
	// For division, every discarded chunk except the most significant one
	// contributes only a sticky bit; that upper chunk decides the half tie.
	sticky := false
	for shift > 19 {
		var rem uint64
		n, rem = fixed01DivChunk(&product, n, fixed01Pow19)
		sticky = sticky || rem != 0
		shift -= 19
	}
	var rem uint64
	den := fixed01Power10(shift)
	n, rem = fixed01DivChunk(&product, n, den)
	inexact := sticky || rem != 0
	if inexact {
		up := false
		switch m {
		case RoundingAwayFromZero:
			up = true
		case RoundingFloor:
			up = neg
		case RoundingCeil:
			up = !neg
		case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
			half := den / 2
			up = rem > half || (rem == half &&
				(sticky || m == RoundingHalfAwayFromZero ||
					(m == RoundingHalfEven && product[0]&1 != 0)))
		}
		if up {
			var carry uint64 = 1
			for i := 0; i < n && carry != 0; i++ {
				product[i], carry = bits.Add64(product[i], 0, carry)
			}
			if carry != 0 {
				if n == len(product) {
					return d.mulRoundLegacy(b, p, m)
				}
				product[n] = carry
				n++
			}
		}
	}
	if n > 8 {
		return NaN(errors.Overflow), inexact
	}
	return fixed01Pack(&product, n, p, neg && n != 0), inexact
}
