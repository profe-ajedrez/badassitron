package uintwide

import "math/bits"

// QuoRemWide divides up to 4096-bit values without heap allocation.
// It implements normalized limb-wise division (Knuth algorithm D), avoiding
// the quadratic per-bit shifts of the original QuoRem hot path.
// q, r, u and v must each have the same length (at most 64) and must not alias.
func QuoRemWide(q, r, u, v []uint64) bool {
	if len(u) != len(v) || len(q) != len(u) || len(r) != len(u) || len(u) > 64 {
		panic("uintwide: invalid wide division length")
	}
	if IsZero(v) {
		return false
	}
	clear(q)
	if Compare(u, v) < 0 {
		copy(r, u)
		return true
	}
	n := (BitLen(v) + 63) / 64
	if n == 1 {
		rem, _ := QuoRem64(q, u, v[0])
		clear(r)
		r[0] = rem
		return true
	}
	uLen := (BitLen(u) + 63) / 64
	m := uLen - n
	shift := uint(bits.LeadingZeros64(v[n-1]))
	var un [65]uint64
	var vn [64]uint64
	if shift == 0 {
		copy(vn[:], v[:n])
		copy(un[:], u[:uLen])
	} else {
		// Normalize divisor and dividend, storing carry in the extra limb.
		var carry uint64
		for i := 0; i < n; i++ {
			vn[i] = v[i]<<shift | carry
			carry = v[i] >> (64 - shift)
		}
		carry = 0
		for i := 0; i < uLen; i++ {
			un[i] = u[i]<<shift | carry
			carry = u[i] >> (64 - shift)
		}
		un[uLen] = carry
	}
	hiDiv := vn[n-1]
	for j := m; j >= 0; j-- {
		top, next := un[j+n], un[j+n-1]
		var qhat, rhat uint64
		overflowRem := false
		if top == hiDiv {
			qhat = ^uint64(0)
			var carry uint64
			rhat, carry = bits.Add64(next, hiDiv, 0)
			overflowRem = carry != 0
		} else {
			qhat, rhat = bits.Div64(top, next, hiDiv)
		}
		if !overflowRem {
			for {
				ph, pl := bits.Mul64(qhat, vn[n-2])
				if ph < rhat || ph == rhat && pl <= un[j+n-2] {
					break
				}
				qhat--
				var carry uint64
				rhat, carry = bits.Add64(rhat, hiDiv, 0)
				if carry != 0 {
					break
				}
			}
		}
		var borrow, carry uint64
		for i := 0; i < n; i++ {
			productHi, productLo := bits.Mul64(qhat, vn[i])
			var overflow uint64
			productLo, overflow = bits.Add64(productLo, carry, 0)
			carry = productHi + overflow
			un[j+i], borrow = bits.Sub64(un[j+i], productLo, borrow)
		}
		un[j+n], borrow = bits.Sub64(un[j+n], carry, borrow)
		if borrow != 0 {
			qhat--
			var carry uint64
			for i := 0; i < n; i++ {
				un[j+i], carry = bits.Add64(un[j+i], vn[i], carry)
			}
			un[j+n], _ = bits.Add64(un[j+n], 0, carry)
		}
		q[j] = qhat
	}
	clear(r)
	if shift == 0 {
		copy(r, un[:n])
	} else {
		for i := 0; i < n; i++ {
			r[i] = un[i]>>shift | un[i+1]<<(64-shift)
		}
	}
	return true
}
