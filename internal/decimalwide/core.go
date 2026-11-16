// Package decimalwide implements the shared, exact-intermediate arithmetic of
// the fixed-width Dec256 and Dec512 types. Public values remain fixed-width;
// this first implementation uses math/big for intermediate calculations. It is
// deliberately separate from dec128 so its existing API and performance do not change.
package decimalwide

import (
	"math/big"
	"strings"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
)

type Number struct {
	Coef  *big.Int // nonnegative magnitude; never mutated by public operations
	Scale int
	Neg   bool
	Err   errors.Error
}

func Zero() Number                  { return Number{Coef: new(big.Int)} }
func Invalid(e errors.Error) Number { return Number{Coef: new(big.Int), Err: e} }
func New(coef *big.Int, scale int, neg bool) Number {
	if coef == nil || coef.Sign() < 0 || scale < 0 {
		return Invalid(errors.InvalidFormat)
	}
	return Number{Coef: new(big.Int).Set(coef), Scale: scale, Neg: neg}
}
func (n Number) Ok() bool     { return n.Err == errors.None }
func (n Number) IsZero() bool { return n.Ok() && n.Coef.Sign() == 0 }
func (n Number) Signed() *big.Int {
	z := new(big.Int).Set(n.Coef)
	if n.Neg {
		z.Neg(z)
	}
	return z
}
func pow10(n int) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil) }

func Parse(s string, bits, maxScale int) Number {
	if s == "" {
		return Zero()
	}
	neg := false
	if s[0] == '-' || s[0] == '+' {
		neg = s[0] == '-'
		s = s[1:]
	}
	if s == "" {
		return Invalid(errors.InvalidFormat)
	}
	dot := strings.IndexByte(s, '.')
	if dot == 0 || dot == len(s)-1 {
		return Invalid(errors.InvalidFormat)
	}
	scale := 0
	if dot >= 0 {
		scale = len(s) - dot - 1
	}
	if scale > maxScale {
		return Invalid(errors.PrecisionOutOfRange)
	}
	for _, c := range s {
		if c != '.' && (c < '0' || c > '9') {
			return Invalid(errors.InvalidFormat)
		}
	}
	if dot >= 0 && strings.IndexByte(s[dot+1:], '.') >= 0 {
		return Invalid(errors.InvalidFormat)
	}
	digits := strings.Replace(s, ".", "", 1)
	coef, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return Invalid(errors.InvalidFormat)
	}
	if coef.BitLen() > bits {
		return Invalid(errors.Overflow)
	}
	return Number{Coef: coef, Scale: scale, Neg: neg}
}

// FitExact preserves the numerical value by removing only exact trailing
// decimal zeros if the requested scale or coefficient cannot be represented.
func FitExact(n Number, bits, maxScale int) Number {
	if !n.Ok() {
		return n
	}
	if n.Coef.Sign() == 0 {
		if n.Scale > maxScale {
			n.Scale = maxScale
		}
		return n
	}
	c := new(big.Int).Set(n.Coef)
	ten := big.NewInt(10)
	r := new(big.Int)
	for n.Scale > 0 && (n.Scale > maxScale || c.BitLen() > bits) {
		q := new(big.Int)
		q.QuoRem(c, ten, r)
		if r.Sign() != 0 {
			break
		}
		c = q
		n.Scale--
	}
	if n.Scale > maxScale {
		return Invalid(errors.PrecisionOutOfRange)
	}
	if c.BitLen() > bits {
		return Invalid(errors.Overflow)
	}
	n.Coef = c
	return n
}

func Canonical(n Number) Number {
	if !n.Ok() {
		return n
	}
	if n.Coef.Sign() == 0 {
		return Zero()
	}
	c := new(big.Int).Set(n.Coef)
	r := new(big.Int)
	for n.Scale > 0 {
		q := new(big.Int)
		q.QuoRem(c, big.NewInt(10), r)
		if r.Sign() != 0 {
			break
		}
		c = q
		n.Scale--
	}
	n.Coef = c
	return n
}

func Cmp(a, b Number) int {
	if !a.Ok() && !b.Ok() {
		return 0
	}
	if !a.Ok() {
		return -1
	}
	if !b.Ok() {
		return 1
	}
	if a.Coef.Sign() == 0 && b.Coef.Sign() == 0 {
		return 0
	}
	if a.Neg && !b.Neg {
		return -1
	}
	if b.Neg && !a.Neg {
		return 1
	}
	ac := new(big.Int).Set(a.Coef)
	bc := new(big.Int).Set(b.Coef)
	if a.Scale < b.Scale {
		ac.Mul(ac, pow10(b.Scale-a.Scale))
	}
	if b.Scale < a.Scale {
		bc.Mul(bc, pow10(a.Scale-b.Scale))
	}
	c := ac.Cmp(bc)
	if a.Neg {
		return -c
	}
	return c
}

func Add(a, b Number, bits, maxScale int) Number {
	if !a.Ok() {
		return a
	}
	if !b.Ok() {
		return b
	}
	scale := max(a.Scale, b.Scale)
	ac := a.Signed()
	bc := b.Signed()
	if a.Scale != scale {
		ac.Mul(ac, pow10(scale-a.Scale))
	}
	if b.Scale != scale {
		bc.Mul(bc, pow10(scale-b.Scale))
	}
	ac.Add(ac, bc)
	neg := ac.Sign() < 0
	ac.Abs(ac)
	return FitExact(New(ac, scale, neg), bits, maxScale)
}
func Sub(a, b Number, bits, maxScale int) Number {
	b.Neg = !b.Neg
	return Add(a, b, bits, maxScale)
}
func Mul(a, b Number, bits, maxScale int) Number {
	if !a.Ok() {
		return a
	}
	if !b.Ok() {
		return b
	}
	c := new(big.Int).Mul(a.Coef, b.Coef)
	return FitExact(New(c, a.Scale+b.Scale, a.Neg != b.Neg), bits, maxScale)
}

func roundUp(q, remainder, divisor *big.Int, neg bool, mode dec128.RoundingMode) bool {
	if remainder.Sign() == 0 {
		return false
	}
	switch mode {
	case dec128.RoundingTowardZero:
		return false
	case dec128.RoundingAwayFromZero:
		return true
	case dec128.RoundingFloor:
		return neg
	case dec128.RoundingCeil:
		return !neg
	case dec128.RoundingHalfTowardZero, dec128.RoundingHalfAwayFromZero, dec128.RoundingHalfEven:
		twice := new(big.Int).Lsh(new(big.Int).Set(remainder), 1)
		cmp := twice.Cmp(divisor)
		return cmp > 0 || cmp == 0 && (mode == dec128.RoundingHalfAwayFromZero || mode == dec128.RoundingHalfEven && q.Bit(0) == 1)
	}
	return false
}
func validMode(mode dec128.RoundingMode) bool { return mode <= dec128.RoundingHalfEven }

// Quantize computes an exact decimal rescaling, then rounds once if required.
// An inexact status means information was discarded, whether or not q increased.
func Quantize(n Number, scale, bits, maxScale int, mode dec128.RoundingMode) (Number, bool) {
	if !n.Ok() {
		return n, false
	}
	if scale < 0 || scale > maxScale {
		return Invalid(errors.PrecisionOutOfRange), false
	}
	if !validMode(mode) {
		return Invalid(errors.InvalidFormat), false
	}
	c := new(big.Int).Set(n.Coef)
	if scale >= n.Scale {
		c.Mul(c, pow10(scale-n.Scale))
		if c.BitLen() > bits {
			return Invalid(errors.Overflow), false
		}
		return New(c, scale, n.Neg), false
	}
	den := pow10(n.Scale - scale)
	r := new(big.Int)
	q := new(big.Int)
	q.QuoRem(c, den, r)
	inexact := r.Sign() != 0
	if roundUp(q, r, den, n.Neg, mode) {
		q.Add(q, big.NewInt(1))
	}
	if q.BitLen() > bits {
		return Invalid(errors.Overflow), inexact
	}
	return New(q, scale, n.Neg && q.Sign() != 0), inexact
}

func Div(a, b Number, scale, bits, maxScale int, mode dec128.RoundingMode) (Number, bool) {
	if !a.Ok() {
		return a, false
	}
	if !b.Ok() {
		return b, false
	}
	if b.Coef.Sign() == 0 {
		return Invalid(errors.DivisionByZero), false
	}
	if scale < 0 || scale > maxScale {
		return Invalid(errors.PrecisionOutOfRange), false
	}
	if !validMode(mode) {
		return Invalid(errors.InvalidFormat), false
	}
	numerator := new(big.Int).Set(a.Coef)
	denominator := new(big.Int).Set(b.Coef)
	shift := b.Scale + scale - a.Scale
	if shift >= 0 {
		numerator.Mul(numerator, pow10(shift))
	} else {
		denominator.Mul(denominator, pow10(-shift))
	}
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(numerator, denominator, r)
	neg := a.Neg != b.Neg
	inexact := r.Sign() != 0
	if roundUp(q, r, denominator, neg, mode) {
		q.Add(q, big.NewInt(1))
	}
	if q.BitLen() > bits {
		return Invalid(errors.Overflow), inexact
	}
	return New(q, scale, neg && q.Sign() != 0), inexact
}

func QuoRem(a, b Number, bits, maxScale int) (Number, Number) {
	if !a.Ok() {
		return a, a
	}
	if !b.Ok() {
		return b, b
	}
	if b.Coef.Sign() == 0 {
		n := Invalid(errors.DivisionByZero)
		return n, n
	}
	scale := max(a.Scale, b.Scale)
	ac := new(big.Int).Set(a.Coef)
	bc := new(big.Int).Set(b.Coef)
	if a.Scale < scale {
		ac.Mul(ac, pow10(scale-a.Scale))
	}
	if b.Scale < scale {
		bc.Mul(bc, pow10(scale-b.Scale))
	}
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(ac, bc, r)
	if q.BitLen() > bits {
		n := Invalid(errors.Overflow)
		return n, n
	}
	rq := New(q, 0, a.Neg != b.Neg && q.Sign() != 0)
	rr := FitExact(New(r, scale, a.Neg && r.Sign() != 0), bits, maxScale)
	if !rr.Ok() {
		return rr, rr
	}
	return rq, rr
}
func Mod(a, b Number, bits, maxScale int) Number {
	if !a.Ok() {
		return a
	}
	if !b.Ok() {
		return b
	}
	if b.Coef.Sign() == 0 {
		return Invalid(errors.DivisionByZero)
	}
	scale := max(a.Scale, b.Scale)
	ac := new(big.Int).Set(a.Coef)
	bc := new(big.Int).Set(b.Coef)
	if a.Scale < scale {
		ac.Mul(ac, pow10(scale-a.Scale))
	}
	if b.Scale < scale {
		bc.Mul(bc, pow10(scale-b.Scale))
	}
	r := new(big.Int).Rem(ac, bc)
	return FitExact(New(r, scale, a.Neg && r.Sign() != 0), bits, maxScale)
}

// Sqrt computes sqrt(n) at a requested scale using integer square root and
// an exact integer midpoint comparison; no floating-point approximation.
func Sqrt(n Number, scale, bits, maxScale int, mode dec128.RoundingMode) (Number, bool) {
	if !n.Ok() {
		return n, false
	}
	if scale < 0 || scale > maxScale {
		return Invalid(errors.PrecisionOutOfRange), false
	}
	if !validMode(mode) {
		return Invalid(errors.InvalidFormat), false
	}
	if n.Neg && n.Coef.Sign() != 0 {
		return Invalid(errors.SqrtNegative), false
	}
	if n.Coef.Sign() == 0 {
		return New(big.NewInt(0), scale, false), false
	}
	numerator := new(big.Int).Set(n.Coef)
	denominator := big.NewInt(1)
	shift := 2*scale - n.Scale
	if shift >= 0 {
		numerator.Mul(numerator, pow10(shift))
	} else {
		denominator = pow10(-shift)
	}
	quotient := new(big.Int).Quo(numerator, denominator)
	root := new(big.Int).Sqrt(quotient)
	sq := new(big.Int).Mul(root, root)
	floorScaled := new(big.Int).Mul(sq, denominator)
	inexact := floorScaled.Cmp(numerator) != 0
	// Exact midpoint comparison: 4*N ? (4*k*k+4*k+1)*D.
	midpoint := new(big.Int).Mul(sq, big.NewInt(4))
	midpoint.Add(midpoint, new(big.Int).Mul(root, big.NewInt(4)))
	midpoint.Add(midpoint, big.NewInt(1))
	midpoint.Mul(midpoint, denominator)
	cmp := new(big.Int).Mul(numerator, big.NewInt(4)).Cmp(midpoint)
	inc := false
	if inexact {
		switch mode {
		case dec128.RoundingAwayFromZero, dec128.RoundingCeil:
			inc = true
		case dec128.RoundingHalfTowardZero, dec128.RoundingHalfAwayFromZero, dec128.RoundingHalfEven:
			inc = cmp > 0 || cmp == 0 && (mode == dec128.RoundingHalfAwayFromZero || mode == dec128.RoundingHalfEven && root.Bit(0) == 1)
		}
	}
	if inc {
		root.Add(root, big.NewInt(1))
	}
	if root.BitLen() > bits {
		return Invalid(errors.Overflow), inexact
	}
	return New(root, scale, false), inexact
}

func PowInt(n Number, power int, bits, maxScale, defaultPrecision int) Number {
	if !n.Ok() {
		return n
	}
	if power == 0 {
		return New(big.NewInt(1), 0, false)
	}
	negative := power < 0
	var magnitude uint
	if negative {
		magnitude = uint(-(power + 1)) + 1
	} else {
		magnitude = uint(power)
	}
	// Exponentiation by squaring with early overflow detection. Huge powers
	// return an error instead of allocating unbounded intermediate integers.
	result := New(big.NewInt(1), 0, false)
	factor := n
	for magnitude > 0 {
		if magnitude&1 != 0 {
			result = Mul(result, factor, bits, maxScale)
			if !result.Ok() {
				return result
			}
		}
		magnitude >>= 1
		if magnitude > 0 {
			factor = Mul(factor, factor, bits, maxScale)
			if !factor.Ok() {
				return factor
			}
		}
	}
	if negative {
		r, _ := Div(New(big.NewInt(1), 0, false), result, defaultPrecision, bits, maxScale, dec128.RoundingTowardZero)
		return r
	}
	return result
}

func Format(n Number, fixed bool) string {
	if !n.Ok() {
		return "NaN"
	}
	digits := n.Coef.String()
	var s string
	if n.Scale == 0 {
		s = digits
	} else if n.Scale >= len(digits) {
		s = "0." + strings.Repeat("0", n.Scale-len(digits)) + digits
	} else {
		i := len(digits) - n.Scale
		s = digits[:i] + "." + digits[i:]
	}
	if !fixed && strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	if n.Coef.Sign() == 0 {
		if !fixed {
			return "0"
		}
	}
	if n.Neg && n.Coef.Sign() != 0 {
		return "-" + s
	}
	return s
}
