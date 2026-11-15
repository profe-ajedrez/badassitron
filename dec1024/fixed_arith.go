package dec1024

// Internal fixed-width intermediate arithmetic. A decimal coefficient occupies
// 16 limbs; the worst-case scaled division and square-root operands require
// fewer than 3072 bits. Sixty-four limbs leave headroom for midpoint checks.
// No arbitrary-precision heap objects are involved in these operations.

import (
	"encoding/binary"
	"math/bits"
	"strings"

	"github.com/profe-ajedrez/badassitron/dec1024/uint1024"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/internal/uintwide"
)

const fixedLimbs = 64

type fixedNum [fixedLimbs]uint64

func fixedFrom(coef uint1024.Uint1024) (v fixedNum) {
	b := coef.Bytes()
	for i := 0; i < 16; i++ {
		v[i] = binary.LittleEndian.Uint64(b[i*8 : i*8+8])
	}
	return
}

func (v fixedNum) coefficient() (uint1024.Uint1024, bool) {
	for _, word := range v[16:] {
		if word != 0 {
			return uint1024.Zero, false
		}
	}
	var b [128]byte
	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint64(b[i*8:i*8+8], v[i])
	}
	return uint1024.FromBytes(b), true
}

func (v fixedNum) zero() bool         { return uintwide.IsZero(v[:]) }
func (v fixedNum) cmp(b fixedNum) int { return uintwide.Compare(v[:], b[:]) }
func (v fixedNum) odd() bool          { return v[0]&1 != 0 }

func (v *fixedNum) add(b fixedNum) bool { return uintwide.Add(v[:], v[:], b[:]) == 0 }
func (v *fixedNum) sub(b fixedNum) bool { return uintwide.Sub(v[:], v[:], b[:]) == 0 }
func (v *fixedNum) inc() bool {
	for i := range v {
		v[i]++
		if v[i] != 0 {
			return true
		}
	}
	return false
}

const pow10_19 uint64 = 10_000_000_000_000_000_000

var smallPowers10 = [20]uint64{1, 10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000, 1000000000, 10000000000, 100000000000, 1000000000000, 10000000000000, 100000000000000, 1000000000000000, 10000000000000000, 100000000000000000, 1000000000000000000, pow10_19}

func (v *fixedNum) mul64(mul uint64) bool {
	var carry uint64
	for i, word := range v {
		hi, lo := bits.Mul64(word, mul)
		var overflow uint64
		lo, overflow = bits.Add64(lo, carry, 0)
		v[i] = lo
		carry = hi + overflow
	}
	return carry == 0
}
func (v *fixedNum) times10(n int) bool {
	for n >= 19 {
		if !v.mul64(pow10_19) {
			return false
		}
		n -= 19
	}
	if n != 0 {
		return v.mul64(smallPowers10[n])
	}
	return true
}
func (v fixedNum) div64(divisor uint64) (fixedNum, uint64) {
	var q fixedNum
	rem, _ := uintwide.QuoRem64(q[:], v[:], divisor)
	return q, rem
}
func (v fixedNum) div10() (fixedNum, uint64) { return v.div64(10) }
func (v fixedNum) quoRem(den fixedNum) (fixedNum, fixedNum) {
	var q, r fixedNum
	if !uintwide.QuoRemWide(q[:], r[:], v[:], den[:]) {
		panic("dec1024: internal division by zero")
	}
	return q, r
}

func fixedProduct(coef1, coef2 uint1024.Uint1024) fixedNum {
	return fixedMul(fixedFrom(coef1), fixedFrom(coef2))
}

// fixedFromProduct reconstructs the exact 2048-bit product already computed by
// Uint1024.MulCarry. Never multiply the same operands a second time.
func fixedFromProduct(lo, hi uint1024.Uint1024) (product fixedNum) {
	low, high := lo.Bytes(), hi.Bytes()
	for i := 0; i < 16; i++ {
		product[i] = binary.LittleEndian.Uint64(low[i*8 : i*8+8])
		product[i+16] = binary.LittleEndian.Uint64(high[i*8 : i*8+8])
	}
	return
}

// fixedMul returns the exact product when it fits 4096 bits. The current
// sqrt caller passes operands whose product is provably narrower than this.
func fixedMul(a, b fixedNum) (out fixedNum) {
	la := (uintwide.BitLen(a[:]) + 63) / 64
	lb := (uintwide.BitLen(b[:]) + 63) / 64
	if la == 0 || lb == 0 {
		return
	}
	if la == 1 && a[0] == 1 {
		return b
	}
	if lb == 1 && b[0] == 1 {
		return a
	}
	// The participating limbs are bounded by the actual operand widths.
	// Multiplying a full 4096-bit buffer when only 1024 bits are occupied
	// was responsible for most of the initial square-root regression.
	if la+lb > fixedLimbs {
		panic("dec1024: internal multiplication width exceeded")
	}
	uintwide.MulFull(out[:la+lb], a[:la], b[:lb])
	return
}

// fixedFitExact matches decimalwide.FitExact: remove trailing decimal zeros
// only when doing so is necessary to fit the public coefficient/scale.
func fixedFitExact(coef fixedNum, scale int, negative bool) Dec1024 {
	if coef.zero() {
		if scale > int(MaxPrecision) {
			scale = int(MaxPrecision)
		}
		return New(uint1024.Zero, uint16(scale), negative)
	}
	_, fits := coef.coefficient()
	for scale > 0 && (scale > int(MaxPrecision) || !fits) {
		// A 19-digit chunk can only be removed when at least 19 zeroes
		// must still be removed; otherwise preserve the original scale.
		if scale >= 19 && (scale-int(MaxPrecision) >= 19 || uintwide.BitLen(coef[:]) > 1024+64) {
			q, rem := coef.div64(pow10_19)
			if rem == 0 {
				coef, scale = q, scale-19
				_, fits = coef.coefficient()
				continue
			}
		}
		q, rem := coef.div10()
		if rem != 0 {
			break
		}
		coef, scale = q, scale-1
		_, fits = coef.coefficient()
	}
	if scale > int(MaxPrecision) {
		return NaN(errors.PrecisionOutOfRange)
	}
	if !fits {
		return NaN(errors.Overflow)
	}
	c, _ := coef.coefficient()
	return New(c, uint16(scale), negative)
}

func fixedRound(q, r, den fixedNum, neg bool, mode RoundingMode) (fixedNum, bool) {
	inexact := !r.zero()
	if !inexact {
		return q, false
	}
	up := false
	switch mode {
	case RoundingAwayFromZero:
		up = true
	case RoundingFloor:
		up = neg
	case RoundingCeil:
		up = !neg
	case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
		complement := den
		complement.sub(r)
		comparison := r.cmp(complement)
		up = comparison > 0 || (comparison == 0 &&
			(mode == RoundingHalfAwayFromZero || (mode == RoundingHalfEven && q.odd())))
	}
	if up {
		q.inc()
	}
	return q, true
}

// divSmallInPlace divides by a power of ten fitting in one limb, processing
// only significant limbs. It leaves the quotient in the same fixed buffer.
func (v *fixedNum) divSmallInPlace(div uint64) (rem uint64) {
	n := (uintwide.BitLen(v[:]) + 63) / 64
	for i := n - 1; i >= 0; i-- {
		v[i], rem = bits.Div64(rem, v[i], div)
	}
	return
}

// fixedRoundPower10 rounds an exact coefficient after removing decimal digits.
// Only the most significant discarded chunk determines whether the fractional
// part is above/below half; lower chunks contribute one sticky bit. This avoids
// materializing 10^digits or a large numerator/divisor pair.
func fixedRoundPower10(coef fixedNum, digits int, neg bool, mode RoundingMode) (fixedNum, bool) {
	if digits <= 0 {
		return coef, false
	}
	sticky := false
	for digits > 19 {
		sticky = coef.divSmallInPlace(pow10_19) != 0 || sticky
		digits -= 19
	}
	den := smallPowers10[digits]
	rem := coef.divSmallInPlace(den)
	inexact := sticky || rem != 0
	if !inexact {
		return coef, false
	}
	up := false
	switch mode {
	case RoundingAwayFromZero:
		up = true
	case RoundingFloor:
		up = neg
	case RoundingCeil:
		up = !neg
	case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
		half := den / 2 // 10^digits is even for all digits > 0.
		up = rem > half || (rem == half &&
			(sticky || mode == RoundingHalfAwayFromZero || (mode == RoundingHalfEven && coef.odd())))
	}
	if up {
		coef.inc()
	}
	return coef, true
}

func fixedQuantize(coef fixedNum, scale int, target uint16, neg bool, mode RoundingMode) (Dec1024, bool) {
	if scale <= int(target) {
		if !coef.times10(int(target) - scale) {
			return NaN(errors.Overflow), false
		}
		c, fits := coef.coefficient()
		if !fits {
			return NaN(errors.Overflow), false
		}
		return New(c, target, neg), false
	}
	q, inexact := fixedRoundPower10(coef, scale-int(target), neg, mode)
	c, fits := q.coefficient()
	if !fits {
		return NaN(errors.Overflow), inexact
	}
	return New(c, target, neg && !q.zero()), inexact
}

func fixedAdd(a, b Dec1024) Dec1024 {
	scale := int(max(a.exp, b.exp))
	x, y := fixedFrom(a.coef), fixedFrom(b.coef)
	x.times10(scale - int(a.exp))
	y.times10(scale - int(b.exp))
	if a.neg == b.neg {
		x.add(y)
		return fixedFitExact(x, scale, a.neg)
	}
	switch x.cmp(y) {
	case 0:
		return New(uint1024.Zero, uint16(scale), false)
	case 1:
		x.sub(y)
		return fixedFitExact(x, scale, a.neg)
	default:
		y.sub(x)
		return fixedFitExact(y, scale, b.neg)
	}
}

func fixedCompare(a, b Dec1024) int {
	if a.IsNaN() && b.IsNaN() {
		return 0
	}
	if a.IsNaN() {
		return -1
	}
	if b.IsNaN() {
		return 1
	}
	if a.coef.IsZero() && b.coef.IsZero() {
		return 0
	}
	if a.neg && !b.neg {
		return -1
	}
	if b.neg && !a.neg {
		return 1
	}
	scale := int(max(a.exp, b.exp))
	x, y := fixedFrom(a.coef), fixedFrom(b.coef)
	x.times10(scale - int(a.exp))
	y.times10(scale - int(b.exp))
	cmp := x.cmp(y)
	if a.neg {
		return -cmp
	}
	return cmp
}

func fixedDivRound(a, b Dec1024, target uint16, mode RoundingMode) (Dec1024, bool) {
	x, y := fixedFrom(a.coef), fixedFrom(b.coef)
	shift := int(b.exp) + int(target) - int(a.exp)
	if shift >= 0 {
		x.times10(shift)
	} else {
		y.times10(-shift)
	}
	q, r := x.quoRem(y)
	neg := a.neg != b.neg
	q, inexact := fixedRound(q, r, y, neg, mode)
	c, fits := q.coefficient()
	if !fits {
		return NaN(errors.Overflow), inexact
	}
	return New(c, target, neg && !q.zero()), inexact
}

func fixedQuoRem(a, b Dec1024) (Dec1024, Dec1024) {
	scale := int(max(a.exp, b.exp))
	x, y := fixedFrom(a.coef), fixedFrom(b.coef)
	x.times10(scale - int(a.exp))
	y.times10(scale - int(b.exp))
	q, r := x.quoRem(y)
	coef, ok := q.coefficient()
	if !ok {
		n := NaN(errors.Overflow)
		return n, n
	}
	quotient := New(coef, 0, a.neg != b.neg && !q.zero())
	remainder := fixedFitExact(r, scale, a.neg && !r.zero())
	if remainder.IsNaN() {
		return remainder, remainder
	}
	return quotient, remainder
}

// Mod must remain available even when the integral quotient overflows Dec1024.
func fixedMod(a, b Dec1024) Dec1024 {
	scale := int(max(a.exp, b.exp))
	x, y := fixedFrom(a.coef), fixedFrom(b.coef)
	x.times10(scale - int(a.exp))
	y.times10(scale - int(b.exp))
	_, r := x.quoRem(y)
	return fixedFitExact(r, scale, a.neg && !r.zero())
}

// fixedSqrtFloor uses monotone integer Newton iteration. The starting
// power-of-two bound is above the true square root; each iteration shrinks it.
func fixedSqrtFloor(n fixedNum) (root fixedNum, remainder fixedNum) {
	if n.zero() {
		return root, n
	}
	bit := uint((uintwide.BitLen(n[:]) + 1) / 2)
	root[bit/64] = 1 << (bit % 64)
	for {
		q, _ := n.quoRem(root)
		next := root
		next.add(q)
		uintwide.Rsh(next[:], next[:], 1)
		if next.cmp(root) >= 0 {
			square := fixedMul(root, root)
			remainder = n
			remainder.sub(square)
			return root, remainder
		}
		root = next
	}
}

func fixedSqrt(a Dec1024, target uint16, mode RoundingMode) (Dec1024, bool) {
	if a.IsNaN() {
		return a, false
	}
	if target > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange), false
	}
	if mode > RoundingHalfEven {
		return NaN(errors.InvalidFormat), false
	}
	if a.neg && !a.coef.IsZero() {
		return NaN(errors.SqrtNegative), false
	}
	if a.coef.IsZero() {
		return New(uint1024.Zero, target, false), false
	}
	numerator := fixedFrom(a.coef)
	denominator := fixedNum{1}
	shift := 2*int(target) - int(a.exp)
	if shift >= 0 {
		numerator.times10(shift)
	} else {
		denominator.times10(-shift)
	}
	quotient, rem := numerator.quoRem(denominator)
	root, sqrtRem := fixedSqrtFloor(quotient)
	inexact := !sqrtRem.zero() || !rem.zero()
	if inexact {
		inc := mode == RoundingAwayFromZero || mode == RoundingCeil
		if mode == RoundingHalfTowardZero || mode == RoundingHalfAwayFromZero || mode == RoundingHalfEven {
			square := fixedMul(root, root)
			floor := fixedMul(square, denominator)
			delta := numerator
			delta.sub(floor)
			left := delta
			uintwide.Lsh(left[:], left[:], 2)
			right := root
			uintwide.Lsh(right[:], right[:], 2)
			right[0]++ // (4*k+1)
			right = fixedMul(right, denominator)
			cmp := left.cmp(right)
			inc = cmp > 0 || (cmp == 0 && (mode == RoundingHalfAwayFromZero || (mode == RoundingHalfEven && root.odd())))
		}
		if inc {
			root.inc()
		}
	}
	coef, fits := root.coefficient()
	if !fits {
		return NaN(errors.Overflow), inexact
	}
	return New(coef, target, false), inexact
}

func fixedPow(a Dec1024, power int) Dec1024 {
	if a.IsNaN() {
		return a
	}
	if power == 0 {
		return One
	}
	negative := power < 0
	var magnitude uint
	if negative {
		magnitude = uint(-(power + 1)) + 1
	} else {
		magnitude = uint(power)
	}
	result, factor := One, a
	for magnitude != 0 {
		if magnitude&1 != 0 {
			result = result.Mul(factor)
			if result.IsNaN() {
				return result
			}
		}
		magnitude >>= 1
		if magnitude != 0 {
			factor = factor.Mul(factor)
			if factor.IsNaN() {
				return factor
			}
		}
	}
	if negative {
		return One.DivRound(result, uint16(defaultPrecision.Load()), RoundingTowardZero)
	}
	return result
}

func fixedParse[S string | []byte](s S) Dec1024 {
	if len(s) == 0 {
		return Zero
	}
	neg, start := false, 0
	if s[0] == '-' || s[0] == '+' {
		neg, start = s[0] == '-', 1
	}
	if start == len(s) {
		return NaN(errors.InvalidFormat)
	}
	dot := -1
	for i := start; i < len(s); i++ {
		if s[i] == '.' {
			dot = i
			break
		}
	}
	if dot == start || dot == len(s)-1 {
		return NaN(errors.InvalidFormat)
	}
	scale := 0
	if dot >= 0 {
		scale = len(s) - dot - 1
	}
	if scale > int(MaxPrecision) {
		return NaN(errors.PrecisionOutOfRange)
	}
	for i := start; i < len(s); i++ {
		if i == dot {
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return NaN(errors.InvalidFormat)
		}
	}
	var coef uint1024.Uint1024
	for i := start; i < len(s); i++ {
		if i == dot {
			continue
		}
		v, e := coef.Mul64(10)
		if e != errors.None {
			return NaN(e)
		}
		coef, e = v.Add64(uint64(s[i] - '0'))
		if e != errors.None {
			return NaN(e)
		}
	}
	return New(coef, uint16(scale), neg)
}

func fixedFormat(d Dec1024, withScale bool) string {
	if d.IsNaN() {
		return "NaN"
	}
	digits := d.coef.String()
	scale := int(d.exp)
	var s string
	if scale == 0 {
		s = digits
	} else if scale >= len(digits) {
		s = "0." + strings.Repeat("0", scale-len(digits)) + digits
	} else {
		split := len(digits) - scale
		s = digits[:split] + "." + digits[split:]
	}
	if !withScale && scale > 0 {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	if d.neg && !d.coef.IsZero() {
		return "-" + s
	}
	return s
}
