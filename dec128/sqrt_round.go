package dec128

import (
	"math/bits"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// SqrtRound returns the square root rounded once to prec decimal places.
// It uses mode without reading or modifying the process-wide default precision.
// Errors (including negative inputs and invalid precision/mode) yield NaN.
func (decimal Dec128) SqrtRound(prec uint8, mode RoundingMode) Dec128 {
	result, _ := decimal.SqrtRoundWithStatus(prec, mode)
	return result
}

// SqrtRoundWithStatus returns SqrtRound and reports whether the mathematically
// exact square root has a nonzero fractional part at the requested scale.
// A true inexact flag does not mean that the result was rounded upward.
// As with other *WithStatus operations, NaN is the error indicator.
func (decimal Dec128) SqrtRoundWithStatus(prec uint8, mode RoundingMode) (Dec128, bool) {
	if decimal.errorCode() != errors.None {
		return decimal, false
	}
	if prec > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange), false
	}
	if mode > RoundingHalfEven {
		return NaN(errors.InvalidFormat), false
	}
	if decimal.IsZero() {
		return New(uint128.Zero, prec, false), false
	}
	if decimal.negative() {
		return NaN(errors.SqrtNegative), false
	}

	root, ok := decimal.sqrtFloorAt(prec)
	if !ok {
		return NaN(errors.Overflow), false
	}

	// The exact coefficient to round is sqrt(coef*10^(2*prec-exp)).
	// Its floor fits Uint128, but the squared radicand may need 256 bits.
	diff := 2*int(prec) - int(decimal.exp)
	var inexact bool
	var halfComparison int // compare exact sqrt to floor+0.5
	if diff >= 0 {
		radicandLow, radicandHigh := decimal.coef.MulCarry(Pow10Uint128[diff])
		squaredLow, squaredHigh := root.MulCarry(root)
		remainderLow, remainderHigh := sub256Nonnegative(radicandLow, radicandHigh, squaredLow, squaredHigh)
		inexact = !remainderLow.IsZero() || !remainderHigh.IsZero()

		// sqrt(N) >= k+0.5 iff N-k^2 >= k+0.25. Since N and
		// k are integers, equality is impossible and we round up exactly
		// when N-k^2 > k.
		if compare256(remainderLow, remainderHigh, root, uint128.Zero) > 0 {
			halfComparison = 1
		} else {
			halfComparison = -1
		}
	} else {
		denominator := Pow10Uint128[-diff]
		squared, err := root.Mul(root)
		if err != errors.None {
			return NaN(err), false
		}
		floorNumerator, err := squared.Mul(denominator)
		if err != errors.None {
			return NaN(err), false
		}
		remainder, err := decimal.coef.Sub(floorNumerator)
		if err != errors.None {
			return NaN(err), false
		}
		inexact = !remainder.IsZero()

		// For n/d with d=10^(-diff), compare sqrt(n/d) to k+0.5
		// exactly: 4*(n-k^2*d) compared with (4*k+1)*d.
		leftLow, leftHigh := remainder.MulCarry(uint128.FromUint64(4))
		fourRoot, err := root.Mul64(4)
		if err != errors.None {
			return NaN(err), false
		}
		fourRootPlusOne, err := fourRoot.Add64(1)
		if err != errors.None {
			return NaN(err), false
		}
		right, err := fourRootPlusOne.Mul(denominator)
		if err != errors.None {
			return NaN(err), false
		}
		halfComparison = compare256(leftLow, leftHigh, right, uint128.Zero)
	}

	increment := false
	if inexact {
		switch mode {
		case RoundingAwayFromZero, RoundingCeil:
			increment = true
		case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
			increment = halfComparison > 0 || halfComparison == 0 &&
				(mode == RoundingHalfAwayFromZero || mode == RoundingHalfEven && root.Lo&1 != 0)
		}
	}
	if increment {
		var err errors.Error
		root, err = root.Add64(1)
		if err != errors.None {
			return NaN(err), inexact
		}
	}
	return New(root, prec, false), inexact
}

// sub256Nonnegative subtracts (bLow,bHigh) from (aLow,aHigh). The caller
// guarantees a>=b, as floor(sqrt(N))^2 <= N.
func sub256Nonnegative(aLow, aHigh, bLow, bHigh uint128.Uint128) (uint128.Uint128, uint128.Uint128) {
	lo, borrow := bits.Sub64(aLow.Lo, bLow.Lo, 0)
	hi, borrow := bits.Sub64(aLow.Hi, bLow.Hi, borrow)
	carryLo, borrow := bits.Sub64(aHigh.Lo, bHigh.Lo, borrow)
	carryHi, _ := bits.Sub64(aHigh.Hi, bHigh.Hi, borrow)
	return uint128.Uint128{Lo: lo, Hi: hi}, uint128.Uint128{Lo: carryLo, Hi: carryHi}
}
