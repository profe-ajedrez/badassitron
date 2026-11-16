// Package uintwide contains allocation-free, fixed-width limb algorithms
// shared by uint256 and uint512. Words are little-endian 64-bit limbs.
//
// This package is internal: public decimal and integer APIs do not depend on
// any particular implementation of these primitives.
package uintwide

import "math/bits"

func IsZero(x []uint64) bool {
	for _, w := range x {
		if w != 0 {
			return false
		}
	}
	return true
}

func Compare(a, b []uint64) int {
	for i := len(a) - 1; i >= 0; i-- {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func BitLen(a []uint64) int {
	for i := len(a) - 1; i >= 0; i-- {
		if a[i] != 0 {
			return i*64 + bits.Len64(a[i])
		}
	}
	return 0
}

func Add(dst, a, b []uint64) (carry uint64) {
	for i := range a {
		dst[i], carry = bits.Add64(a[i], b[i], carry)
	}
	return
}

func Sub(dst, a, b []uint64) (borrow uint64) {
	for i := range a {
		dst[i], borrow = bits.Sub64(a[i], b[i], borrow)
	}
	return
}

// MulFull writes the complete 2*n-word product of two n-word integers.
// It does not discard carries or truncate the result.
func MulFull(dst, a, b []uint64) {
	clear(dst)
	for i, aw := range a {
		for j, bw := range b {
			hi, lo := bits.Mul64(aw, bw)
			k := i + j
			var c uint64
			dst[k], c = bits.Add64(dst[k], lo, 0)
			dst[k+1], c = bits.Add64(dst[k+1], hi, c)
			for k += 2; c != 0; k++ {
				// The product of two n-word numbers fits in 2*n words.
				dst[k], c = bits.Add64(dst[k], 0, c)
			}
		}
	}
}

// Lsh and Rsh allow in-place operation. Shifts >= width produce zero.
func Lsh(dst, src []uint64, shift uint) {
	n := len(src)
	word := int(shift / 64)
	if word >= n {
		clear(dst)
		return
	}
	part := shift % 64
	for i := n - 1; i >= 0; i-- {
		var v uint64
		if i >= word {
			v = src[i-word] << part
			if part != 0 && i > word {
				v |= src[i-word-1] >> (64 - part)
			}
		}
		dst[i] = v
	}
}

func Rsh(dst, src []uint64, shift uint) {
	n := len(src)
	word := int(shift / 64)
	if word >= n {
		clear(dst)
		return
	}
	part := shift % 64
	for i := 0; i < n; i++ {
		var v uint64
		if i+word < n {
			v = src[i+word] >> part
			if part != 0 && i+word+1 < n {
				v |= src[i+word+1] << (64 - part)
			}
		}
		dst[i] = v
	}
}

// QuoRem64 divides a fixed-width numerator by one 64-bit limb.
// Returns false for division by zero.
func QuoRem64(q, u []uint64, v uint64) (uint64, bool) {
	if v == 0 {
		return 0, false
	}
	var rem uint64
	for i := len(u) - 1; i >= 0; i-- {
		q[i], rem = bits.Div64(rem, u[i], v)
	}
	return rem, true
}

// QuoRem performs an unsigned, exact, fixed-width long division.
// q and r must be distinct from u and v, and all slices have equal length.
// Returns false for a zero divisor.
func QuoRem(q, r, u, v []uint64) bool {
	if IsZero(v) {
		return false
	}
	clear(q)
	copy(r, u)
	if Compare(u, v) < 0 {
		return true
	}

	// Fast path: single-limb divisor, with no heap allocations.
	if BitLen(v) <= 64 {
		rem, _ := QuoRem64(q, u, v[0])
		clear(r)
		r[0] = rem
		return true
	}

	// Align divisor's highest bit to the numerator and subtract on each
	// quotient bit. Every shifted divisor remains representable at width.
	shift := BitLen(u) - BitLen(v)
	var work [8]uint64 // maximum width is 512 bits
	Lsh(work[:len(u)], v, uint(shift))
	for i := shift; i >= 0; i-- {
		if Compare(r, work[:len(u)]) >= 0 {
			Sub(r, r, work[:len(u)])
			q[i/64] |= 1 << uint(i%64)
		}
		Rsh(work[:len(u)], work[:len(u)], 1)
	}
	return true
}

func LeadingZeros(a []uint64) int {
	return len(a)*64 - BitLen(a)
}
func TrailingZeros(a []uint64) int {
	for i, w := range a {
		if w != 0 {
			return i*64 + bits.TrailingZeros64(w)
		}
	}
	return len(a) * 64
}
func OnesCount(a []uint64) int {
	n := 0
	for _, w := range a {
		n += bits.OnesCount64(w)
	}
	return n
}
