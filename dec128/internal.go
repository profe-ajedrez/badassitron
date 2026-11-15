package dec128

import (
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

var (
	// precalculated StringFixed values for 0 Dec128 in all possible prec
	zeroStrs = [...]string{
		"0",                     // 10^0
		"0.0",                   // 10^1
		"0.00",                  // 10^2
		"0.000",                 // 10^3
		"0.0000",                // 10^4
		"0.00000",               // 10^5
		"0.000000",              // 10^6
		"0.0000000",             // 10^7
		"0.00000000",            // 10^8
		"0.000000000",           // 10^9
		"0.0000000000",          // 10^10
		"0.00000000000",         // 10^11
		"0.000000000000",        // 10^12
		"0.0000000000000",       // 10^13
		"0.00000000000000",      // 10^14
		"0.000000000000000",     // 10^15
		"0.0000000000000000",    // 10^16
		"0.00000000000000000",   // 10^17
		"0.000000000000000000",  // 10^18
		"0.0000000000000000000", // 10^19
	}

	// precalculated array of zero characters
	zeros = [...]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
)

func (decimal Dec128) tryAdd(other Dec128) (Dec128, bool) {
	prec := max(decimal.exp, other.exp)

	a := decimal.Rescale(prec)
	if a.IsNaN() {
		return a, false
	}

	b := other.Rescale(prec)
	if b.IsNaN() {
		return b, false
	}

	if a.neg == b.neg {
		coef, err := a.coef.Add(b.coef)
		if err != errors.None {
			return NaN(err), false
		}
		return Dec128{coef: coef, exp: prec, neg: a.neg}, true
	}

	switch a.coef.Compare(b.coef) {
	case 1:
		coef, err := a.coef.Sub(b.coef)
		if err != errors.None {
			return NaN(err), false
		}
		return Dec128{coef: coef, exp: prec, neg: a.neg}, true
	case 0:
		return Zero, true
	default:
		coef, err := b.coef.Sub(a.coef)
		if err != errors.None {
			return NaN(err), false
		}
		return Dec128{coef: coef, exp: prec, neg: b.neg}, true
	}
}

func (decimal Dec128) trySub(other Dec128) (Dec128, bool) {
	prec := max(decimal.exp, other.exp)

	a := decimal.Rescale(prec)
	if a.IsNaN() {
		return a, false
	}

	b := other.Rescale(prec)
	if b.IsNaN() {
		return b, false
	}

	if a.neg != b.neg {
		coef, err := a.coef.Add(b.coef)
		if err != errors.None {
			return NaN(err), false
		}
		return Dec128{coef: coef, exp: prec, neg: a.neg}, true
	}

	switch a.coef.Compare(b.coef) {
	case 1:
		coef, err := a.coef.Sub(b.coef)
		if err != errors.None {
			return NaN(err), false
		}
		return Dec128{coef: coef, exp: prec, neg: a.neg}, true
	case 0:
		return Zero, true
	default:
		coef, err := b.coef.Sub(a.coef)
		if err != errors.None {
			return NaN(err), false
		}
		return Dec128{coef: coef, exp: prec, neg: !a.neg}, true
	}
}

func (decimal Dec128) tryMul(other Dec128) (Dec128, bool) {
	neg := decimal.neg != other.neg
	prec := decimal.exp + other.exp
	rcoef, rcarry := decimal.coef.MulCarry(other.coef)

	if rcarry.IsZero() {
		r := Dec128{coef: rcoef, exp: prec, neg: neg}
		if prec <= MaxPrecision {
			return r, true
		}
		r = r.Canonical()
		return r, r.exp <= MaxPrecision
	}

	i := prec
	for {
		if i == 0 {
			return NaN(errors.Overflow), false
		}
		q, r, err := uint128.QuoRem256By128(rcoef, rcarry, Pow10Uint128[i])
		if err == errors.None && r.IsZero() {
			return Dec128{coef: q, exp: prec - i, neg: neg}, true
		}
		if err == errors.Overflow {
			return NaN(errors.Overflow), false
		}
		i--
		if prec-i > MaxPrecision {
			return NaN(errors.Overflow), false
		}
	}
}

func (decimal Dec128) tryDiv(other Dec128) (Dec128, bool) {
	neg := decimal.neg != other.neg
	factor := other.exp
	prec := decimal.exp
	targetPrecision := uint8(defaultPrecision.Load())
	if prec < targetPrecision {
		factor = factor + targetPrecision - prec
		prec = targetPrecision
	}
	u, c := decimal.coef.MulCarry(Pow10Uint128[factor])
	q, _, err := uint128.QuoRem256By128(u, c, other.coef)
	if err != errors.None {
		return NaN(err), false
	}
	return Dec128{coef: q, exp: prec, neg: neg}, true
}

func (decimal Dec128) tryQuoRem(other Dec128) (Dec128, Dec128, bool) {
	// If the dividend is smaller, integer division is exactly zero. Do not
	// upscale the divisor: it can overflow Uint128 although the remainder is
	// just the original (fully representable) dividend.
	if compareMagnitude(decimal, other) < 0 {
		return Zero, decimal, true
	}

	var factor uint8
	var u uint128.Uint128
	var c uint128.Uint128
	var d uint128.Uint128
	var err errors.Error

	if decimal.exp == other.exp {
		factor = decimal.exp
		u = decimal.coef
		d = other.coef
	} else {
		factor = max(decimal.exp, other.exp)
		u, c = decimal.coef.MulCarry(Pow10Uint128[factor-decimal.exp])
		d, err = other.coef.Mul(Pow10Uint128[factor-other.exp])
		if err != errors.None {
			return NaN(err), NaN(err), false
		}
	}

	q1, r1, err := uint128.QuoRem256By128(u, c, d)
	if err != errors.None {
		return NaN(err), NaN(err), false
	}

	return Dec128{coef: q1, exp: 0, neg: decimal.neg != other.neg}, Dec128{coef: r1, exp: factor, neg: decimal.neg}, true
}

// tryMod computes the remainder independently of the representability of the
// integer quotient. QuoRem correctly fails when its integer quotient does not
// fit in 128 bits; Mod does not need to represent that quotient at all.
func (decimal Dec128) tryMod(other Dec128) (Dec128, bool) {
	if compareMagnitude(decimal, other) < 0 {
		return decimal, true
	}

	factor := max(decimal.exp, other.exp)
	numeratorLow, numeratorHigh := decimal.coef.MulCarry(Pow10Uint128[factor-decimal.exp])
	divisor, err := other.coef.Mul(Pow10Uint128[factor-other.exp])
	if err != errors.None {
		return NaN(err), false
	}

	// Reduce the high 128 bits modulo the divisor before division. This
	// ensures the final division fits in 128 bits even when the original
	// integer quotient would require up to 192 bits.
	if numeratorHigh.Compare(divisor) >= 0 {
		numeratorHigh, err = numeratorHigh.Mod(divisor)
		if err != errors.None {
			return NaN(err), false
		}
	}
	_, remainder, err := uint128.QuoRem256By128(numeratorLow, numeratorHigh, divisor)
	if err != errors.None {
		return NaN(err), false
	}
	return Dec128{coef: remainder, exp: factor, neg: decimal.neg}, true
}

// appendString appends the string representation of the decimal to sb. Returns the new slice and whether the decimal contains a decimal point.
func (decimal Dec128) appendString(sb []byte) ([]byte, bool) {
	buf := [uint128.MaxStrLen]byte{}
	coef := decimal.coef.StringToBuf(buf[:])

	if decimal.neg {
		sb = append(sb, '-')
	}

	prec := int(decimal.exp)
	if prec == 0 {
		return append(sb, coef...), false
	}

	sz := len(coef)
	if prec > sz {
		sb = append(sb, '0', '.')
		sb = append(sb, zeros[:prec-sz]...)
		sb = append(sb, coef...)
	} else if prec == sz {
		sb = append(sb, '0', '.')
		sb = append(sb, coef...)
	} else {
		sb = append(sb, coef[:sz-prec]...)
		sb = append(sb, '.')
		sb = append(sb, coef[sz-prec:]...)
	}

	return sb, true
}

func trimTrailingZeros(sb []byte) []byte {
	i := len(sb)

	for i > 0 && sb[i-1] == '0' {
		i--
	}

	if i > 0 && sb[i-1] == '.' {
		i--
	}

	return sb[:i]
}

// sqrtFloor256 returns floor(sqrt(high*2^128 + low)). Its initial Newton
// estimate is never smaller than the true root. The subtraction-based average
// avoids overflow when x and n/x together would exceed 128 bits.
func sqrtFloor256(low, high uint128.Uint128) (uint128.Uint128, bool) {
	if low.IsZero() && high.IsZero() {
		return uint128.Zero, true
	}

	bitLen := low.BitLen()
	if !high.IsZero() {
		bitLen = 128 + high.BitLen()
	}
	initialBit := uint((bitLen + 1) / 2)
	var x uint128.Uint128
	if initialBit == 128 {
		// sqrt of a 256-bit integer is at most 2^128-1.
		x = uint128.Max
	} else {
		x = uint128.One.Lsh(initialBit)
	}

	for {
		y, _, err := uint128.QuoRem256By128(low, high, x)
		if err != errors.None {
			return uint128.Zero, false
		}
		if y.Compare(x) >= 0 {
			return x, true
		}
		// floor((x+y)/2) = x-ceil((x-y)/2).
		delta, _ := x.Sub(y)
		half := delta.Rsh(1)
		if delta.Lo&1 != 0 {
			half, _ = half.Add64(1)
		}
		x, _ = x.Sub(half)
	}
}

// sqrtFloorAt returns the coefficient of the truncated square root at an
// explicitly supplied precision. For a negative exponent difference,
// floor(sqrt(floor(n/d))) equals floor(sqrt(n/d)). The original fractional
// remainder is retained by callers that need exact rounding decisions.
func (decimal Dec128) sqrtFloorAt(prec uint8) (uint128.Uint128, bool) {
	diff := int(prec)*2 - int(decimal.exp)
	coef := decimal.coef
	if diff < 0 {
		var err errors.Error
		coef, err = coef.Div(Pow10Uint128[-diff])
		if err != errors.None {
			return uint128.Zero, false
		}
		diff = 0
	}
	low, high := coef.MulCarry(Pow10Uint128[diff])
	return sqrtFloor256(low, high)
}

// trySqrt implements the legacy truncated square root using the process-wide
// default precision. Its public rounding contract remains unchanged.
func (decimal Dec128) trySqrt() (Dec128, bool) {
	prec := uint8(defaultPrecision.Load())
	root, ok := decimal.sqrtFloorAt(prec)
	if !ok {
		return NaN(errors.Overflow), false
	}
	return Dec128{coef: root, exp: prec}, true
}
