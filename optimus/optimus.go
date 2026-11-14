// Package optimus provides an adaptive fixed-point decimal with the same
// ordinary arithmetic API as dec128. Coefficients use the smallest of 64,
// 128, 256 or 512 bits. Operations never use binary floating-point arithmetic.
//
// Ordinary arithmetic preserves exact results whenever possible; otherwise it
// rounds half-even to the highest representable scale (up to 154). Division
// and square root use defaultPrecision for non-terminating/irrational results.
package optimus

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec256"
	"github.com/profe-ajedrez/badassitron/dec256/uint256"
	"github.com/profe-ajedrez/badassitron/dec512"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
	"github.com/profe-ajedrez/badassitron/internal/decimalwide"
)

// MaxPrecision is the largest supported decimal scale for this type.
const MaxPrecision uint8 = 154

// MaxScale is an alias for MaxPrecision.
const MaxScale = MaxPrecision

// MaxStrLen bounds the length of a decimal text representation.
const MaxStrLen = uint512.MaxStrLen + int(MaxPrecision) + 3

// The same seven rounding modes are shared across the decimal family.
type RoundingMode = dec128.RoundingMode

const (
	// RoundingTowardZero rounds toward zero.
	RoundingTowardZero = dec128.RoundingTowardZero
	// RoundingAwayFromZero rounds away from zero.
	RoundingAwayFromZero = dec128.RoundingAwayFromZero
	// RoundingFloor rounds toward negative infinity.
	RoundingFloor = dec128.RoundingFloor
	// RoundingCeil rounds toward positive infinity.
	RoundingCeil = dec128.RoundingCeil
	// RoundingHalfTowardZero rounds to the nearest value, with ties toward zero.
	RoundingHalfTowardZero = dec128.RoundingHalfTowardZero
	// RoundingHalfAwayFromZero rounds to the nearest value, with ties away from zero.
	RoundingHalfAwayFromZero = dec128.RoundingHalfAwayFromZero
	// RoundingHalfEven rounds to the nearest value, with ties to even.
	RoundingHalfEven = dec128.RoundingHalfEven
)

// Optimus is an immutable-by-convention value. The coefficient is inline for
// 64/128-bit values and indirect only for 256/512-bit values. The pointed-to
// word array is never modified after construction, including by copies.
// This representation is intentionally non-comparable: use Equal/Compare.
type Optimus struct {
	small uint128.Uint128
	wide  *uint512.Uint512
	err   errors.Error
	exp   uint8
	neg   bool
}

var (
	// Zero is the predefined decimal value zero.
	Zero = Optimus{}
	// One is the predefined decimal value one.
	One = FromInt(1)
	// NegativeOne is the predefined decimal value minus one.
	NegativeOne = FromInt(-1)
	// Decimal0 is the predefined decimal value 0.
	Decimal0 = Zero
	// Decimal1 is the predefined decimal value 1.
	Decimal1 = One
	// Decimal2 is the predefined decimal value 2.
	Decimal2 = FromInt(2)
	// Decimal3 is the predefined decimal value 3.
	Decimal3 = FromInt(3)
	// Decimal4 is the predefined decimal value 4.
	Decimal4 = FromInt(4)
	// Decimal5 is the predefined decimal value 5.
	Decimal5 = FromInt(5)
	// Decimal6 is the predefined decimal value 6.
	Decimal6 = FromInt(6)
	// Decimal7 is the predefined decimal value 7.
	Decimal7 = FromInt(7)
	// Decimal8 is the predefined decimal value 8.
	Decimal8 = FromInt(8)
	// Decimal9 is the predefined decimal value 9.
	Decimal9 = FromInt(9)
	// Decimal10 is the predefined decimal value 10.
	Decimal10 = FromInt(10)
	// Decimal100 is the predefined decimal value 100.
	Decimal100 = FromInt(100)
	// Decimal365 is the predefined decimal value 365.
	Decimal365 = FromInt(365)
	// Decimal366 is the predefined decimal value 366.
	Decimal366 = FromInt(366)
	// Decimal1000 is the predefined decimal value 1000.
	Decimal1000 = FromInt(1000)
	// Ten is a predefined decimal value equal to ten.
	Ten = Decimal10
	// Hundred is a predefined decimal value equal to one hundred.
	Hundred = Decimal100
	// ZeroStr is the decimal text for zero.
	ZeroStr = "0"
	// ZeroStrBytes contains the mutable bytes of the text zero. Treat this exported variable as read only.
	ZeroStrBytes = []byte("0")
	// ZeroJsonStrBytes contains the mutable JSON string bytes for zero. Treat this exported variable as read only.
	ZeroJsonStrBytes = []byte(`"0"`)
	// NaNStr is the decimal text for an invalid value.
	NaNStr = "NaN"
	// NaNStrBytes contains the mutable bytes of the text NaN. Treat this exported variable as read only.
	NaNStrBytes = []byte("NaN")
	// NaNJsonStrBytes contains the mutable JSON string bytes for NaN. Treat this exported variable as read only.
	NaNJsonStrBytes  = []byte(`"NaN"`)
	defaultPrecision atomic.Uint32
)

func init() { defaultPrecision.Store(19) }

// SetDefaultPrecision sets the process wide working scale for operations that need a default. The setting starts at 19, is atomic, and panics if p exceeds MaxPrecision.
func SetDefaultPrecision(p uint8) {
	if p > MaxPrecision {
		panic(errors.PrecisionOutOfRange.Value())
	}
	defaultPrecision.Store(uint32(p))
}

// NaN returns an invalid decimal with the requested error reason. A missing reason becomes NotANumber.
func NaN(reason errors.Error) Optimus {
	if reason == errors.None {
		reason = errors.NotANumber
	}
	return Optimus{err: reason}
}

// New uses a 512-bit coefficient because it must accept every Optimus tier.
// Unlike automatic arithmetic, explicit construction rejects an invalid scale.
func New(c uint512.Uint512, exp uint8, neg bool) Optimus {
	if exp > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange)
	}
	if c.IsZero() {
		return Optimus{exp: exp}
	}
	if c.Hi.IsZero() {
		return NewFromUint256(c.Lo, exp, neg)
	}
	v := c
	return Optimus{wide: &v, exp: exp, neg: neg}
}

// NewFromUint128 constructs a decimal from a 128 bit coefficient, scale, and sign. An invalid scale returns NaN.
func NewFromUint128(c uint128.Uint128, exp uint8, neg bool) Optimus {
	if exp > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange)
	}
	if c.IsZero() {
		return Optimus{exp: exp}
	}
	return Optimus{small: c, exp: exp, neg: neg}
}

// NewFromUint256 constructs a decimal from a 256 bit coefficient, scale, and sign. An invalid scale returns NaN.
func NewFromUint256(c uint256.Uint256, exp uint8, neg bool) Optimus {
	if exp > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange)
	}
	if c.Hi.IsZero() {
		return NewFromUint128(c.Lo, exp, neg)
	}
	v := uint512.FromUint256(c)
	return Optimus{wide: &v, exp: exp, neg: neg}
}

// DecodeFromUint64 constructs a nonnegative decimal from an unsigned coefficient and scale.
func DecodeFromUint64(c uint64, exp uint8) Optimus {
	return NewFromUint128(uint128.FromUint64(c), exp, false)
}

// DecodeFromUint128 constructs a nonnegative decimal from an unsigned coefficient and scale.
func DecodeFromUint128(c uint128.Uint128, exp uint8) Optimus { return NewFromUint128(c, exp, false) }

// DecodeFromUint256 constructs a nonnegative decimal from an unsigned coefficient and scale.
func DecodeFromUint256(c uint256.Uint256, exp uint8) Optimus { return NewFromUint256(c, exp, false) }

// DecodeFromUint512 constructs a nonnegative decimal from an unsigned coefficient and scale.
func DecodeFromUint512(c uint512.Uint512, exp uint8) Optimus { return New(c, exp, false) }

// FromInt creates a decimal value from a signed int.
func FromInt(i int) Optimus { return FromInt64(int64(i)) }

// NewFromInt is an alias for FromInt.
func NewFromInt(i int) Optimus { return FromInt(i) }

// FromInt64 creates a decimal value from a signed 64 bit integer.
func FromInt64(i int64) Optimus {
	neg := i < 0
	if neg {
		return NewFromUint128(uint128.FromUint64(uint64(-(i+1))+1), 0, true)
	}
	return DecodeFromUint64(uint64(i), 0)
}

// FromFloat64 converts a finite float64 to decimal through its shortest fixed point text. This cannot restore digits already lost in binary floating point.
func FromFloat64(v float64) Optimus {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return NaN(errors.NotANumber)
	}
	return FromString(strconv.FormatFloat(v, 'f', -1, 64))
}

// FromString parses an ordinary fixed-point decimal, with optional sign.
// A valid numeric literal beyond 154 fractional places or 512 coefficient
// bits is approximated half-even when an approximation remains representable.
func FromString[S string | []byte](input S) Optimus {
	s := string(input)
	if s == "" {
		return Zero
	}
	neg := false
	if s[0] == '+' || s[0] == '-' {
		neg = s[0] == '-'
		s = s[1:]
	}
	if s == "" {
		return NaN(errors.InvalidFormat)
	}
	dot := strings.IndexByte(s, '.')
	if dot == 0 || dot == len(s)-1 {
		return NaN(errors.InvalidFormat)
	}
	scale := 0
	if dot >= 0 {
		scale = len(s) - dot - 1
	}
	if len(s) > 10000 {
		return NaN(errors.InvalidFormat)
	} // bound untrusted input
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			if i != dot {
				return NaN(errors.InvalidFormat)
			}
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return NaN(errors.InvalidFormat)
		}
	}
	digits := s
	if dot >= 0 {
		digits = s[:dot] + s[dot+1:]
	}
	coeff, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return NaN(errors.InvalidFormat)
	}
	return auto(decimalwide.Number{Coef: coeff, Scale: scale, Neg: neg})
}

// NewFromString parses decimal text and returns an error instead of NaN when parsing fails.
func NewFromString(s string) (Optimus, error) {
	d := FromString(s)
	if d.IsNaN() {
		return Zero, d.ErrorDetails()
	}
	return d, nil
}

// IsNaN reports whether the value represents an error rather than an ordinary number.
func (d Optimus) IsNaN() bool { return d.err != errors.None }

// ErrorDetails returns the error carried by a NaN value.
func (d Optimus) ErrorDetails() error { return d.err.Value() }

// IsZero reports whether this is an ordinary numeric zero.
func (d Optimus) IsZero() bool { return !d.IsNaN() && d.small.IsZero() && d.wide == nil }

// IsNegative reports whether this is an ordinary negative value.
func (d Optimus) IsNegative() bool { return !d.IsNaN() && !d.IsZero() && d.neg }

// IsPositive reports whether this is an ordinary positive value.
func (d Optimus) IsPositive() bool { return !d.IsNaN() && !d.IsZero() && !d.neg }

// Sign returns minus one for negative values, one for positive values, and zero for zero or NaN.
func (d Optimus) Sign() int {
	if d.IsNaN() || d.IsZero() {
		return 0
	}
	if d.neg {
		return -1
	}
	return 1
}

// Precision returns the stored fractional decimal scale, not the number of significant digits.
func (d Optimus) Precision() uint8 { return d.exp }

// Exponent returns the stored fractional decimal scale, matching Precision.
func (d Optimus) Exponent() uint8 { return d.exp }

// Copy returns the same value. Shared wide Optimus storage remains private and read only.
func (d Optimus) Copy() Optimus { return d }

// Width reports the current coefficient tier, not the nominal maximum.
func (d Optimus) Width() int {
	if d.wide != nil {
		if d.wide.Hi.IsZero() {
			return 256
		}
		return 512
	}
	if d.small.Hi == 0 {
		return 64
	}
	return 128
}

// Coefficient returns the unsigned coefficient without the sign or decimal scale.
func (d Optimus) Coefficient() uint512.Uint512 {
	if d.wide != nil {
		return *d.wide
	}
	return uint512.FromUint256(uint256.FromUint128(d.small))
}
func (d Optimus) coefBig() *big.Int {
	if d.wide != nil {
		return d.wide.BigInt()
	}
	return d.small.BigInt()
}
func (d Optimus) core() decimalwide.Number {
	if d.IsNaN() {
		return decimalwide.Invalid(d.err)
	}
	return decimalwide.Number{Coef: d.coefBig(), Scale: int(d.exp), Neg: d.neg}
}
func pack(n decimalwide.Number) Optimus {
	if n.Err != errors.None {
		return NaN(n.Err)
	}
	if n.Coef == nil || n.Scale < 0 {
		return NaN(errors.InvalidFormat)
	}
	if n.Scale > int(MaxPrecision) {
		return NaN(errors.PrecisionOutOfRange)
	}
	c, e := uint512.FromBigInt(n.Coef)
	if e != errors.None {
		return NaN(e)
	}
	return New(c, uint8(n.Scale), n.Neg)
}

// auto keeps an exact value if representable. When the scale or coefficient
// does not fit, it applies one half-even rounding at the highest possible
// target scale; it never saturates an integer overflow.
func auto(n decimalwide.Number) Optimus {
	if !n.Ok() {
		return NaN(n.Err)
	}
	if n.Coef == nil || n.Coef.Sign() < 0 || n.Scale < 0 {
		return NaN(errors.InvalidFormat)
	}
	if n.Coef.Sign() == 0 {
		return Optimus{exp: uint8(min(n.Scale, int(MaxPrecision)))}
	}
	if n.Scale <= int(MaxPrecision) && n.Coef.BitLen() <= 512 {
		return pack(n)
	}
	c := new(big.Int).Set(n.Coef)
	scale := n.Scale
	ten := big.NewInt(10)
	r := new(big.Int)
	for scale > 0 && (scale > int(MaxPrecision) || c.BitLen() > 512) {
		q := new(big.Int)
		q.QuoRem(c, ten, r)
		if r.Sign() != 0 {
			break
		}
		c = q
		scale--
	}
	if scale <= int(MaxPrecision) && c.BitLen() <= 512 {
		return pack(decimalwide.Number{Coef: c, Scale: scale, Neg: n.Neg})
	}
	for p := min(scale, int(MaxPrecision)); p >= 0; p-- {
		rounded, _ := decimalwide.Quantize(decimalwide.Number{Coef: c, Scale: scale, Neg: n.Neg}, p, 512, int(MaxPrecision), RoundingHalfEven)
		if rounded.Ok() {
			return pack(rounded)
		}
	}
	return NaN(errors.Overflow)
}

// Canonical returns a numerically equivalent value with unnecessary trailing decimal zeros removed.
func (d Optimus) Canonical() Optimus {
	if d.IsNaN() {
		return d
	}
	return pack(decimalwide.Canonical(d.core()))
}

// Compare compares numeric values, ignoring differences in stored scale. NaN sorts before ordinary values.
func (d Optimus) Compare(b Optimus) int {
	if d.IsNaN() {
		if b.IsNaN() {
			return 0
		}
		return -1
	}
	if b.IsNaN() {
		return 1
	}
	if d.IsZero() {
		if b.IsZero() {
			return 0
		}
		if b.neg {
			return 1
		}
		return -1
	}
	if b.IsZero() {
		if d.neg {
			return -1
		}
		return 1
	}
	if d.neg != b.neg {
		if d.neg {
			return -1
		}
		return 1
	}
	if d.wide == nil && b.wide == nil {
		p := max(d.exp, b.exp)
		if int(p-d.exp) <= 38 && int(p-b.exp) <= 38 {
			loA, hiA := d.small.MulCarry(uint128.Pow10Uint128[p-d.exp])
			loB, hiB := b.small.MulCarry(uint128.Pow10Uint128[p-b.exp])
			cmp := (uint256.Uint256{Lo: loA, Hi: hiA}).Compare(uint256.Uint256{Lo: loB, Hi: hiB})
			if d.neg {
				return -cmp
			}
			return cmp
		}
	}
	return decimalwide.Cmp(d.core(), b.core())
}

// Equal reports numeric equality across different representations. Two NaN values compare equal.
func (d Optimus) Equal(b Optimus) bool {
	if d.IsNaN() || b.IsNaN() {
		return d.IsNaN() && b.IsNaN()
	}
	return d.Compare(b) == 0
}

// LessThan reports whether the receiver sorts before b under Compare.
func (d Optimus) LessThan(b Optimus) bool { return d.Compare(b) < 0 }

// LessThanOrEqual reports whether the receiver sorts before or equal to b under Compare.
func (d Optimus) LessThanOrEqual(b Optimus) bool { return d.Compare(b) <= 0 }

// GreaterThan reports whether the receiver sorts after b under Compare.
func (d Optimus) GreaterThan(b Optimus) bool { return d.Compare(b) > 0 }

// GreaterThanOrEqual reports whether the receiver sorts after or equal to b under Compare.
func (d Optimus) GreaterThanOrEqual(b Optimus) bool { return d.Compare(b) >= 0 }

// Add adds two decimals and preserves exactness when representable.
func (d Optimus) Add(b Optimus) Optimus {
	if d.IsNaN() {
		return d
	}
	if b.IsNaN() {
		return b
	}
	if d.wide == nil && b.wide == nil && d.exp == b.exp && d.small.Hi == 0 && b.small.Hi == 0 {
		if d.neg == b.neg {
			lo, carry := bits.Add64(d.small.Lo, b.small.Lo, 0)
			return NewFromUint128(uint128.Uint128{Lo: lo, Hi: carry}, d.exp, d.neg)
		}
		switch {
		case d.small.Lo == b.small.Lo:
			return Optimus{exp: d.exp}
		case d.small.Lo > b.small.Lo:
			return NewFromUint128(uint128.FromUint64(d.small.Lo-b.small.Lo), d.exp, d.neg)
		default:
			return NewFromUint128(uint128.FromUint64(b.small.Lo-d.small.Lo), d.exp, b.neg)
		}
	}
	if d.wide == nil && b.wide == nil {
		p := max(d.exp, b.exp)
		if int(p-d.exp) <= 38 && int(p-b.exp) <= 38 {
			aa, highA := d.small.MulCarry(uint128.Pow10Uint128[p-d.exp])
			bb, highB := b.small.MulCarry(uint128.Pow10Uint128[p-b.exp])
			if highA.IsZero() && highB.IsZero() {
				if d.neg == b.neg {
					sum, e := aa.Add(bb)
					if e == errors.None {
						return NewFromUint128(sum, p, d.neg)
					}
					// The sum is necessarily <=129 bits, still fits Uint256.
					wide, _ := uint256.FromUint128(aa).Add(uint256.FromUint128(bb))
					return NewFromUint256(wide, p, d.neg)
				}
				switch aa.Compare(bb) {
				case 0:
					return Optimus{exp: p}
				case 1:
					c, _ := aa.Sub(bb)
					return NewFromUint128(c, p, d.neg)
				default:
					c, _ := bb.Sub(aa)
					return NewFromUint128(c, p, b.neg)
				}
			}
		}
	}
	// All representable aligned operands use the fixed-width integer engine.
	// This also covers 256/512-bit values without allocating big.Int objects.
	p := max(d.exp, b.exp)
	aa, bb := d.Coefficient(), b.Coefficient()
	var errA, errB errors.Error
	if d.exp < p {
		aa, errA = aa.Mul(uint512.Pow10Uint512[p-d.exp])
	}
	if b.exp < p {
		bb, errB = bb.Mul(uint512.Pow10Uint512[p-b.exp])
	}
	if errA == errors.None && errB == errors.None {
		if d.neg == b.neg {
			v, e := aa.Add(bb)
			if e == errors.None {
				return New(v, p, d.neg)
			}
		} else {
			switch aa.Compare(bb) {
			case 0:
				return Optimus{exp: p}
			case 1:
				v, _ := aa.Sub(bb)
				return New(v, p, d.neg)
			default:
				v, _ := bb.Sub(aa)
				return New(v, p, b.neg)
			}
		}
	}
	a, c := d.core(), b.core()
	scale := max(a.Scale, c.Scale)
	left, right := a.Signed(), c.Signed()
	if a.Scale < scale {
		left.Mul(left, pow10(scale-a.Scale))
	}
	if c.Scale < scale {
		right.Mul(right, pow10(scale-c.Scale))
	}
	left.Add(left, right)
	neg := left.Sign() < 0
	left.Abs(left)
	return auto(decimalwide.Number{Coef: left, Scale: scale, Neg: neg})
}

// AddInt adds a signed integer to the receiver.
func (d Optimus) AddInt(b int) Optimus { return d.Add(FromInt(b)) }

// Sub subtracts b from the receiver.
func (d Optimus) Sub(b Optimus) Optimus { return d.Add(b.Neg()) }

// SubInt subtracts a signed integer from the receiver.
func (d Optimus) SubInt(b int) Optimus { return d.Sub(FromInt(b)) }

// Mul multiplies two decimals and preserves exactness when representable.
func (d Optimus) Mul(b Optimus) Optimus {
	if d.IsNaN() {
		return d
	}
	if b.IsNaN() {
		return b
	}
	if d.IsZero() || b.IsZero() {
		return Zero
	}
	if d.wide == nil && b.wide == nil && d.small.Hi == 0 && b.small.Hi == 0 && int(d.exp)+int(b.exp) <= int(MaxPrecision) {
		hi, lo := bits.Mul64(d.small.Lo, b.small.Lo)
		return NewFromUint128(uint128.Uint128{Lo: lo, Hi: hi}, d.exp+b.exp, d.neg != b.neg)
	}
	if d.wide == nil && b.wide == nil && int(d.exp)+int(b.exp) <= int(MaxPrecision) {
		lo, hi := d.small.MulCarry(b.small)
		if hi.IsZero() {
			return NewFromUint128(lo, d.exp+b.exp, d.neg != b.neg)
		}
		return NewFromUint256(uint256.Uint256{Lo: lo, Hi: hi}, d.exp+b.exp, d.neg != b.neg)
	}
	if int(d.exp)+int(b.exp) <= int(MaxPrecision) {
		lo, hi := d.Coefficient().MulCarry(b.Coefficient())
		if hi.IsZero() {
			return New(lo, d.exp+b.exp, d.neg != b.neg)
		}
		// At scale zero, no decimal reduction can make an overflowing
		// integer coefficient representable. Avoid rebuilding the product.
		if d.exp+b.exp == 0 {
			return NaN(errors.Overflow)
		}
	}
	return d.mulFixed02B(b)
}

// MulInt multiplies the receiver by a signed integer.
func (d Optimus) MulInt(b int) Optimus { return d.Mul(FromInt(b)) }

// divideRational uses the exact reduced rational to detect terminating
// divisions, rather than prematurely truncating a potentially exact answer.
func divideRational(num, den *big.Int, neg bool, preferred uint8) Optimus {
	if den.Sign() == 0 {
		return NaN(errors.DivisionByZero)
	}
	if num.Sign() == 0 {
		return Zero
	}
	g := new(big.Int).GCD(nil, nil, num, den)
	n := new(big.Int).Quo(num, g)
	d := new(big.Int).Quo(den, g)
	factor := new(big.Int).Set(d)
	count2 := factor.TrailingZeroBits()
	factor.Rsh(factor, count2)
	count5 := 0
	rem := new(big.Int)
	five := big.NewInt(5)
	for {
		q := new(big.Int)
		q.QuoRem(factor, five, rem)
		if rem.Sign() != 0 {
			break
		}
		factor = q
		count5++
	}
	if factor.Cmp(big.NewInt(1)) == 0 {
		scale := max(int(count2), count5)
		value := new(big.Int).Mul(n, pow10(scale))
		value.Quo(value, d)
		return auto(decimalwide.Number{Coef: value, Scale: scale, Neg: neg})
	}
	// Non-terminating decimal: round to the configured working precision.
	for p := int(preferred); p >= 0; p-- {
		value, _ := decimalwide.Div(decimalwide.Number{Coef: n, Neg: neg},
			decimalwide.Number{Coef: d}, p, 512, int(MaxPrecision), RoundingHalfEven)
		if value.Ok() {
			return pack(value)
		}
	}
	return NaN(errors.Overflow)
}
func pow10(n int) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil) }

// Div divides by b using the default working precision for a repeating quotient.
func (d Optimus) Div(b Optimus) Optimus {
	if d.IsNaN() {
		return d
	}
	if b.IsNaN() {
		return b
	}
	if b.IsZero() {
		return NaN(errors.DivisionByZero)
	}
	if d.IsZero() {
		return Zero
	}
	if d.wide == nil && b.wide == nil && d.exp == b.exp {
		q, r, e := d.small.QuoRem(b.small)
		if e == errors.None && r.IsZero() {
			return NewFromUint128(q, 0, d.neg != b.neg)
		}
		// Small equal-scale rationals can be classified without big.Int.
		if d.small.Hi == 0 && b.small.Hi == 0 {
			n, den := d.small.Lo, b.small.Lo
			x, y := n, den
			for y != 0 {
				x, y = y, x%y
			}
			reduced := den / x
			twos, fives := 0, 0
			for reduced%2 == 0 {
				reduced /= 2
				twos++
			}
			for reduced%5 == 0 {
				reduced /= 5
				fives++
			}
			preferred := max(int(defaultPrecision.Load()), int(d.exp))
			if reduced == 1 {
				exactScale := max(twos, fives)
				if exactScale <= 38 {
					if value, _, ok := d.nativeDivRound(b, uint8(exactScale), RoundingHalfEven); ok {
						return value
					}
				}
			} else if preferred <= 38 {
				if value, _, ok := d.nativeDivRound(b, uint8(preferred), RoundingHalfEven); ok {
					return value
				}
			}
		}
	}
	a, c := d.core(), b.core()
	n := new(big.Int).Mul(a.Coef, pow10(c.Scale))
	den := new(big.Int).Mul(c.Coef, pow10(a.Scale))
	return divideRational(n, den, a.Neg != c.Neg, uint8(max(int(d.exp), int(defaultPrecision.Load()))))
}

// DivInt divides by a signed integer.
func (d Optimus) DivInt(b int) Optimus { return d.Div(FromInt(b)) }

// smallQuoRem aligns 128-bit coefficients without constructing wide operands.
// It defers to math/big if either scaled coefficient exceeds 128 bits.
func smallQuoRem(a, b Optimus) (q, r uint128.Uint128, scale uint8, ok bool) {
	if a.wide != nil || b.wide != nil {
		return q, r, 0, false
	}
	scale = max(a.exp, b.exp)
	num, den := a.small, b.small
	if a.exp < scale {
		if scale-a.exp > 38 {
			return q, r, scale, false
		}
		low, high := num.MulCarry(uint128.Pow10Uint128[scale-a.exp])
		if !high.IsZero() {
			return q, r, scale, false
		}
		num = low
	}
	if b.exp < scale {
		if scale-b.exp > 38 {
			return q, r, scale, false
		}
		low, high := den.MulCarry(uint128.Pow10Uint128[scale-b.exp])
		if !high.IsZero() {
			return q, r, scale, false
		}
		den = low
	}
	q, r, _ = num.QuoRem(den)
	return q, r, scale, true
}

// wideQuoRem handles equal-scale wide operands only when fixed-width division
// avoids the cost of a large bit-at-a-time quotient. All other cases retain
// the original math/big implementation.
func wideQuoRem(a, b Optimus) (q, r uint512.Uint512, ok bool) {
	if a.exp != b.exp {
		return q, r, false
	}
	// A single-limb divisor is handled by a linear fixed-width division,
	// including when the quotient uses all 512 bits.
	if b.wide == nil && b.small.Hi == 0 {
		qq, rr, _ := a.Coefficient().QuoRem64(b.small.Lo)
		return qq, uint512.FromUint64(rr), true
	}
	if a.wide != nil {
		bBits := b.small.BitLen()
		if b.wide != nil {
			bBits = b.wide.BitLen()
		}
		if a.wide.BitLen()-bBits > 8 {
			return q, r, false
		}
	}
	num, den := a.Coefficient(), b.Coefficient()
	if num.Compare(den) < 0 {
		return uint512.Zero, num, true
	}
	qq, rr, _ := num.QuoRem(den)
	return qq, rr, true
}

// Mod returns the remainder after truncating division by b.
func (d Optimus) Mod(b Optimus) Optimus {
	if d.IsNaN() {
		return d
	}
	if b.IsNaN() {
		return b
	}
	if b.IsZero() {
		return NaN(errors.DivisionByZero)
	}
	if _, rem, scale, ok := smallQuoRem(d, b); ok {
		return NewFromUint128(rem, scale, d.neg)
	}
	if _, rem, ok := wideQuoRem(d, b); ok {
		return New(rem, d.exp, d.neg)
	}
	a, c := d.core(), b.core()
	scale := max(a.Scale, c.Scale)
	// core() provides independent coefficients. Only the operand with a
	// lower scale needs multiplication; a factor of 10^0 is redundant.
	num, den := a.Coef, c.Coef
	if a.Scale < scale {
		num.Mul(num, pow10(scale-a.Scale))
	}
	if c.Scale < scale {
		den.Mul(den, pow10(scale-c.Scale))
	}
	r := new(big.Int).Rem(num, den)
	return auto(decimalwide.Number{Coef: r, Scale: scale, Neg: d.neg})
}

// ModInt returns the remainder after division by a signed integer.
func (d Optimus) ModInt(b int) Optimus { return d.Mod(FromInt(b)) }

// QuoRem returns the integer quotient and signed remainder of truncating division by b.
func (d Optimus) QuoRem(b Optimus) (Optimus, Optimus) {
	if d.IsNaN() {
		return d, d
	}
	if b.IsNaN() {
		return b, b
	}
	if b.IsZero() {
		n := NaN(errors.DivisionByZero)
		return n, n
	}
	if q, r, scale, ok := smallQuoRem(d, b); ok {
		return NewFromUint128(q, 0, d.neg != b.neg), NewFromUint128(r, scale, d.neg)
	}
	if q, r, ok := wideQuoRem(d, b); ok {
		return New(q, 0, d.neg != b.neg), New(r, d.exp, d.neg)
	}
	a, c := d.core(), b.core()
	scale := max(a.Scale, c.Scale)
	// core() provides independent coefficients. Only the operand with a
	// lower scale needs multiplication; a factor of 10^0 is redundant.
	num, den := a.Coef, c.Coef
	if a.Scale < scale {
		num.Mul(num, pow10(scale-a.Scale))
	}
	if c.Scale < scale {
		den.Mul(den, pow10(scale-c.Scale))
	}
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(num, den, r)
	if q.BitLen() > 512 {
		n := NaN(errors.Overflow)
		return n, n
	}
	return pack(decimalwide.Number{Coef: q, Scale: 0, Neg: d.neg != b.neg}), auto(decimalwide.Number{Coef: r, Scale: scale, Neg: d.neg})
}

// QuoRemInt returns the integer quotient and remainder after division by a signed integer.
func (d Optimus) QuoRemInt(b int) (Optimus, Optimus) { return d.QuoRem(FromInt(b)) }

// Abs returns the absolute value without changing the receiver.
func (d Optimus) Abs() Optimus {
	if !d.IsNaN() {
		d.neg = false
	}
	return d
}

// Neg returns the additive inverse without changing the receiver.
func (d Optimus) Neg() Optimus {
	if !d.IsNaN() && !d.IsZero() {
		d.neg = !d.neg
	}
	return d
}

// PowInt raises the value to an integer exponent. Unresolved extreme powers may return a precision error.
func (d Optimus) PowInt(n int) Optimus {
	if d.IsNaN() {
		return d
	}
	if n == 0 {
		return One
	}
	if d.IsZero() {
		if n < 0 {
			return NaN(errors.DivisionByZero)
		}
		return Zero
	}
	magnitude := uint(n)
	if n < 0 {
		magnitude = uint(-(n + 1)) + 1
	}
	if d.Equal(One) {
		return One
	}
	if d.Equal(NegativeOne) {
		if magnitude&1 != 0 {
			return NegativeOne
		}
		return One
	}
	// Large exponents near one must not allocate an enormous exact
	// intermediate. Evaluate them using rigorously enclosing intervals.
	if magnitude > 4096 {
		return d.Abs().powBounded(magnitude, n < 0, d.neg && magnitude&1 != 0)
	}
	coef := new(big.Int).Exp(d.coefBig(), new(big.Int).SetUint64(uint64(magnitude)), nil)
	scale := int(d.exp) * int(magnitude)
	neg := d.neg && magnitude&1 != 0
	if n >= 0 {
		return auto(decimalwide.Number{Coef: coef, Scale: scale, Neg: neg})
	}
	return divideRational(pow10(scale), coef, neg, uint8(defaultPrecision.Load()))
}

// powBounded evaluates large integer exponents with enclosing decimal intervals.
// Each intermediate operation rounds the lower bound down and upper bound up.
// A result is returned only when both bounds select the same half-even value;
// otherwise the guard precision is increased rather than guessing.
func (d Optimus) powBounded(exponent uint, reciprocal bool, negative bool) Optimus {
	maxCoef := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	for _, guard := range []int{192, 256, 384, 640, 1024, 1536, 2048} {
		unit := pow10(guard)
		maxScaled := new(big.Int).Mul(maxCoef, unit)
		num, den := d.coefBig(), pow10(int(d.exp))
		if reciprocal {
			num, den = den, num
		}
		lo, rem := new(big.Int).QuoRem(new(big.Int).Mul(num, unit), den, new(big.Int))
		hi := new(big.Int).Set(lo)
		if rem.Sign() != 0 {
			hi.Add(hi, big.NewInt(1))
		}
		lower, upper := new(big.Int).Set(unit), new(big.Int).Set(unit)
		mulDown := func(a, b *big.Int) *big.Int {
			return new(big.Int).Quo(new(big.Int).Mul(a, b), unit)
		}
		mulUp := func(a, b *big.Int) *big.Int {
			prod := new(big.Int).Mul(a, b)
			q, r := new(big.Int).QuoRem(prod, unit, new(big.Int))
			if r.Sign() != 0 {
				q.Add(q, big.NewInt(1))
			}
			return q
		}
		power := exponent
		for power != 0 {
			if power&1 != 0 {
				lower = mulDown(lower, lo)
				upper = mulUp(upper, hi)
			}
			power >>= 1
			if power != 0 {
				lo = mulDown(lo, lo)
				// All remaining powers are positive. Once even the lower
				// bound of their base exceeds the largest integer, no
				// representable result is possible. Check before squaring
				// again to prevent exponential memory growth.
				if lo.Cmp(maxScaled) > 0 {
					return NaN(errors.Overflow)
				}
				hi = mulUp(hi, hi)
			}
			if lower.Cmp(maxScaled) > 0 {
				return NaN(errors.Overflow)
			}
		}
		if lower.Cmp(maxScaled) > 0 {
			return NaN(errors.Overflow)
		}
		// Select the highest representable scale. If that scale's rounding
		// is ambiguous, increase guard precision rather than returning a
		// coarser value: agreement at a lower scale does not justify losing
		// decimal places that the exact result can retain.
		factor := pow10(int(MaxPrecision))
		for scale := int(MaxPrecision); scale >= 0; scale-- {
			qlo := roundHalfEvenRatio(new(big.Int).Mul(lower, factor), unit)
			if qlo.Cmp(maxCoef) > 0 {
				// Even the lower bound exceeds the coefficient limit at
				// this scale, so only a coarser scale can fit.
				factor.Quo(factor, big.NewInt(10))
				continue
			}
			qhi := roundHalfEvenRatio(new(big.Int).Mul(upper, factor), unit)
			if qlo.Cmp(qhi) != 0 || qhi.Cmp(maxCoef) > 0 {
				// This scale might fit, but the bounds do not yet prove
				// its rounded result. Retry it with tighter bounds.
				break
			}
			return pack(decimalwide.Number{Coef: qlo, Scale: scale, Neg: negative})
		}
		// If the upper bound is strictly below half the minimum unit,
		// zero is the correctly rounded result.
		if new(big.Int).Mul(new(big.Int).Mul(upper, big.NewInt(2)), pow10(int(MaxPrecision))).Cmp(unit) < 0 {
			return Zero
		}
	}
	return NaN(errors.PrecisionOutOfRange)
}

func roundHalfEvenRatio(numerator, denominator *big.Int) *big.Int {
	q, rem := new(big.Int).QuoRem(numerator, denominator, new(big.Int))
	cmp := new(big.Int).Lsh(rem, 1).Cmp(denominator)
	if cmp > 0 || (cmp == 0 && q.Bit(0) != 0) {
		q.Add(q, big.NewInt(1))
	}
	return q
}

// Sqrt returns an exact square root when possible and otherwise uses the default working precision.
func (d Optimus) Sqrt() Optimus {
	if d.IsNaN() {
		return d
	}
	if d.IsNegative() {
		return NaN(errors.SqrtNegative)
	}
	if d.IsZero() {
		return Zero
	}
	a := d.core()
	num := new(big.Int).Set(a.Coef)
	den := pow10(a.Scale)
	g := new(big.Int).GCD(nil, nil, num, den)
	num.Quo(num, g)
	den.Quo(den, g)
	rn := new(big.Int).Sqrt(num)
	rd := new(big.Int).Sqrt(den)
	if new(big.Int).Mul(new(big.Int).Set(rn), rn).Cmp(num) == 0 && new(big.Int).Mul(new(big.Int).Set(rd), rd).Cmp(den) == 0 {
		return divideRational(rn, rd, false, uint8(defaultPrecision.Load()))
	}
	for p := int(max(int(d.exp), int(defaultPrecision.Load()))); p >= 0; p-- {
		result, _ := decimalwide.Sqrt(a, p, 512, int(MaxPrecision), RoundingHalfEven)
		if result.Ok() {
			return pack(result)
		}
	}
	return NaN(errors.Overflow)
}

// Rescale changes the stored decimal scale, truncating toward zero when digits must be removed.
func (d Optimus) Rescale(p uint8) Optimus { return d.RescaleRound(p, RoundingTowardZero) }

// RescaleRound changes the scale using the requested rounding mode.
func (d Optimus) RescaleRound(p uint8, m RoundingMode) Optimus {
	v, _ := d.RescaleRoundWithStatus(p, m)
	return v
}

// RescaleRoundWithStatus changes the scale and reports whether nonzero digits were discarded. An error is carried in NaN, not in the boolean.
func (d Optimus) RescaleRoundWithStatus(p uint8, m RoundingMode) (Optimus, bool) {
	if d.IsNaN() {
		return d, false
	}
	// Validate before the shortcuts: invalid arguments must retain their
	// original NaN and status semantics even when the scale is unchanged.
	if p <= MaxPrecision && m <= RoundingHalfEven {
		if p == d.exp {
			return d, false
		}
		// A short decimal divisor fits uint64 for at most 19 digits. This
		// selects an algorithm, not a limit on precision or decimal scale.
		// Larger reductions and all wide coefficients keep the exact path.
		if d.wide == nil && p < d.exp && d.exp-p <= 19 {
			den := fixed01Power10(int(d.exp - p))
			q, r, err := d.small.QuoRem64(den)
			if err == errors.None {
				inexact := r != 0
				if inexact {
					up := false
					switch m {
					case RoundingAwayFromZero:
						up = true
					case RoundingFloor:
						up = d.neg
					case RoundingCeil:
						up = !d.neg
					case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
						up = r > den/2 || (r == den/2 &&
							(m == RoundingHalfAwayFromZero ||
								(m == RoundingHalfEven && q.Lo&1 != 0)))
					}
					if up {
						if rounded, addErr := q.Add64(1); addErr == errors.None {
							q = rounded
						} else {
							// Keep the original route for unexpected overflow.
							goto rescaleGeneral
						}
					}
				}
				return NewFromUint128(q, p, d.neg && !q.IsZero()), inexact
			}
		}
	}
rescaleGeneral:
	// Dec512 rescaling uses fixed-width arithmetic. It avoids allocating
	// big.Int temporaries for the common case, while preserving all seven
	// rounding modes and the inexact status. Do not take this path for
	// invalid inputs: the original decimalwide implementation remains the
	// authority for error handling and unrepresentable results.
	if p <= MaxPrecision && m <= RoundingHalfEven {
		rescaled, inexact := dec512.New(d.Coefficient(), d.exp, d.neg).RescaleRoundWithStatus(p, m)
		if !rescaled.IsNaN() {
			return New(rescaled.Coefficient(), p, rescaled.IsNegative()), inexact
		}
	}
	// A fixed-width intermediate can overflow even if a later rounding
	// operation would succeed. Keep the existing exact fallback for every
	// such case, and for invalid precision or rounding mode.
	n, inx := decimalwide.Quantize(d.core(), int(p), 512, int(MaxPrecision), m)
	return pack(n), inx
}

// MulRound multiplies and rounds the exact product to the requested scale.
func (d Optimus) MulRound(b Optimus, p uint8, m RoundingMode) Optimus {
	v, _ := d.MulRoundWithStatus(b, p, m)
	return v
}

// MulRoundWithStatus multiplies with explicit rounding and reports whether nonzero digits were discarded. Check NaN for errors.
func (d Optimus) MulRoundWithStatus(b Optimus, p uint8, m RoundingMode) (Optimus, bool) {
	if d.IsNaN() {
		return d, false
	}
	if b.IsNaN() {
		return b, false
	}
	if p > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange), false
	}
	if m > RoundingHalfEven {
		return NaN(errors.InvalidFormat), false
	}
	return d.mulRoundFixed01(b, p, m)
}

// mulRoundLegacy preserves the pre-experiment exact math/big calculation as
// the reference fallback. It is deliberately not used on the ordinary path.
func (d Optimus) mulRoundLegacy(b Optimus, p uint8, m RoundingMode) (Optimus, bool) {
	if d.IsNaN() {
		return d, false
	}
	if b.IsNaN() {
		return b, false
	}
	a, c := d.core(), b.core()
	prod := decimalwide.Number{Coef: new(big.Int).Mul(a.Coef, c.Coef), Scale: a.Scale + c.Scale, Neg: a.Neg != c.Neg}
	r, inx := decimalwide.Quantize(prod, int(p), 512, int(MaxPrecision), m)
	return pack(r), inx
}

// DivRound divides and rounds to the requested scale and rounding mode.
func (d Optimus) DivRound(b Optimus, p uint8, m RoundingMode) Optimus {
	v, _ := d.DivRoundWithStatus(b, p, m)
	return v
}

// DivRoundWithStatus divides with explicit rounding and reports whether nonzero information was discarded. Check NaN for errors.
func (d Optimus) DivRoundWithStatus(b Optimus, p uint8, m RoundingMode) (Optimus, bool) {
	if d.IsNaN() {
		return d, false
	}
	if b.IsNaN() {
		return b, false
	}
	if p > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange), false
	}
	if m > RoundingHalfEven {
		return NaN(errors.InvalidFormat), false
	}
	if b.IsZero() {
		return NaN(errors.DivisionByZero), false
	}
	if d.wide == nil && b.wide == nil {
		if value, inx, ok := d.nativeDivRound(b, p, m); ok {
			return value, inx
		}
	}
	r, inx := decimalwide.Div(d.core(), b.core(), int(p), 512, int(MaxPrecision), m)
	return pack(r), inx
}

// nativeDivRound evaluates the scaled quotient using at most 256 bits.
// It returns ok=false if the fast path cannot represent an intermediate;
// callers then use the exact big-integer fallback rather than truncate.
func (d Optimus) nativeDivRound(b Optimus, p uint8, m RoundingMode) (Optimus, bool, bool) {
	shift := int(p) + int(b.exp) - int(d.exp)
	numerator, denominator := d.small, b.small
	var quotient, remainder uint128.Uint128
	var e errors.Error
	var denHigh uint128.Uint128
	if shift >= 0 {
		if shift > 38 {
			return Zero, false, false
		}
		low, high := numerator.MulCarry(uint128.Pow10Uint128[shift])
		quotient, remainder, e = uint128.QuoRem256By128(low, high, denominator)
		if e != errors.None {
			return Zero, false, false
		}
	} else {
		if -shift > 38 {
			return Zero, false, false
		}
		denLow, high := denominator.MulCarry(uint128.Pow10Uint128[-shift])
		denHigh = high
		denominator = denLow
		if high.IsZero() {
			quotient, remainder, e = numerator.QuoRem(denLow)
			if e != errors.None {
				return Zero, false, false
			}
		} else {
			quotient = uint128.Zero
			remainder = numerator
		}
	}
	inexact := !remainder.IsZero()
	neg := d.neg != b.neg
	if inexact {
		up := false
		switch m {
		case RoundingTowardZero:
		case RoundingAwayFromZero:
			up = true
		case RoundingFloor:
			up = neg
		case RoundingCeil:
			up = !neg
		case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
			var cmp int
			if denHigh.IsZero() {
				complement, _ := denominator.Sub(remainder)
				cmp = remainder.Compare(complement)
			} else {
				den256 := uint256.Uint256{Lo: denominator, Hi: denHigh}
				complement, _ := den256.Sub(uint256.FromUint128(remainder))
				cmp = uint256.FromUint128(remainder).Compare(complement)
			}
			up = cmp > 0 || cmp == 0 && (m == RoundingHalfAwayFromZero || m == RoundingHalfEven && quotient.Lo&1 == 1)
		}
		if up {
			var err errors.Error
			quotient, err = quotient.Add64(1)
			if err != errors.None {
				return Zero, false, false
			}
		}
	}
	return NewFromUint128(quotient, p, neg), inexact, true
}

// SqrtRound computes a square root rounded to the requested scale and mode.
func (d Optimus) SqrtRound(p uint8, m RoundingMode) Optimus {
	v, _ := d.SqrtRoundWithStatus(p, m)
	return v
}

// SqrtRoundWithStatus computes a square root with explicit rounding and reports inexactness separately from errors.
func (d Optimus) SqrtRoundWithStatus(p uint8, m RoundingMode) (Optimus, bool) {
	r, inx := decimalwide.Sqrt(d.core(), int(p), 512, int(MaxPrecision), m)
	return pack(r), inx
}

// Round rounds to the nearest value at scale p, with ties away from zero.
func (d Optimus) Round(p uint8) Optimus { return d.RoundHalfAwayFromZero(p) }

// RoundDown rounds to scale p toward negative infinity.
func (d Optimus) RoundDown(p uint8) Optimus { return d.roundOnly(p, RoundingFloor) }

// RoundUp rounds to scale p toward positive infinity.
func (d Optimus) RoundUp(p uint8) Optimus { return d.roundOnly(p, RoundingCeil) }

// RoundTowardZero rounds to scale p toward zero.
func (d Optimus) RoundTowardZero(p uint8) Optimus { return d.roundOnly(p, RoundingTowardZero) }

// RoundAwayFromZero rounds to scale p away from zero.
func (d Optimus) RoundAwayFromZero(p uint8) Optimus { return d.roundOnly(p, RoundingAwayFromZero) }

// RoundHalfTowardZero rounds to scale p to the nearest value, with ties toward zero.
func (d Optimus) RoundHalfTowardZero(p uint8) Optimus { return d.roundOnly(p, RoundingHalfTowardZero) }

// RoundHalfAwayFromZero rounds to scale p to the nearest value, with ties away from zero.
func (d Optimus) RoundHalfAwayFromZero(p uint8) Optimus {
	return d.roundOnly(p, RoundingHalfAwayFromZero)
}

// RoundBank rounds to scale p to the nearest value, with ties to even.
func (d Optimus) RoundBank(p uint8) Optimus { return d.roundOnly(p, RoundingHalfEven) }

// Trunc discards fractional digits beyond scale p without rounding away from zero.
func (d Optimus) Trunc(p uint8) Optimus { return d.roundOnly(p, RoundingTowardZero) }
func (d Optimus) roundOnly(p uint8, m RoundingMode) Optimus {
	if d.IsNaN() || p >= d.exp {
		return d
	}
	return d.RescaleRound(p, m)
}

// Min returns the smallest value under Compare.
func Min(a Optimus, rest ...Optimus) Optimus {
	for _, v := range rest {
		if v.Compare(a) < 0 {
			a = v
		}
	}
	return a
}

// Max returns the largest value under Compare.
func Max(a Optimus, rest ...Optimus) Optimus {
	for _, v := range rest {
		if v.Compare(a) > 0 {
			a = v
		}
	}
	return a
}

// Sum adds the first value and all remaining values.
func Sum(a Optimus, rest ...Optimus) Optimus {
	for _, v := range rest {
		a = a.Add(v)
	}
	return a
}

// Avg returns the sum divided by the number of values.
func Avg(a Optimus, rest ...Optimus) Optimus { return Sum(a, rest...).DivInt(len(rest) + 1) }

// String returns decimal text without insignificant trailing fractional zeros. NaN is written as NaN.
func (d Optimus) String() string { return d.format(false) }

// StringFixed returns decimal text retaining the stored fractional scale.
func (d Optimus) StringFixed() string { return d.format(true) }
func (d Optimus) format(fixed bool) string {
	if d.IsNaN() {
		return "NaN"
	}
	var digits string
	if d.wide == nil {
		digits = d.small.String()
	} else {
		digits = d.wide.String()
	}
	p := int(d.exp)
	var out string
	if p == 0 {
		out = digits
	} else if p >= len(digits) {
		out = "0." + strings.Repeat("0", p-len(digits)) + digits
	} else {
		out = digits[:len(digits)-p] + "." + digits[len(digits)-p:]
	}
	if !fixed && p > 0 {
		out = strings.TrimRight(strings.TrimRight(out, "0"), ".")
	}
	if d.neg && !d.IsZero() {
		return "-" + out
	}
	return out
}

// MarshalText implements encoding.TextMarshaler using the decimal String representation.
func (d Optimus) MarshalText() ([]byte, error) { return []byte(d.String()), nil }

// UnmarshalText implements encoding.TextUnmarshaler. Invalid text leaves the destination unchanged.
func (d *Optimus) UnmarshalText(src []byte) error {
	if d == nil {
		return fmt.Errorf("optimus: nil destination")
	}
	v := FromString(src)
	if v.IsNaN() {
		return v.ErrorDetails()
	}
	*d = v
	return nil
}

// MarshalJSON encodes the decimal as a JSON string to avoid numeric decoding through float64.
func (d Optimus) MarshalJSON() ([]byte, error) { return []byte(strconv.Quote(d.String())), nil }

// UnmarshalJSON accepts a JSON decimal string or number. JSON null becomes numeric zero.
func (d *Optimus) UnmarshalJSON(src []byte) error {
	if d == nil {
		return fmt.Errorf("optimus: nil destination")
	}
	if bytes.Equal(src, []byte("null")) {
		*d = Zero
		return nil
	}
	if len(src) >= 2 && src[0] == '"' && src[len(src)-1] == '"' {
		s, err := strconv.Unquote(string(src))
		if err != nil {
			return err
		}
		return d.UnmarshalText([]byte(s))
	}
	return d.UnmarshalText(src)
}

// Scan implements sql.Scanner with legacy numeric conversions. NULL becomes zero and floats are accepted.
func (d *Optimus) Scan(src any) error {
	if d == nil {
		return fmt.Errorf("optimus: nil scan destination")
	}
	switch v := src.(type) {
	case nil:
		*d = Zero
		return nil
	case string:
		return d.UnmarshalText([]byte(v))
	case []byte:
		return d.UnmarshalText(v)
	case int:
		*d = FromInt(v)
		return nil
	case int64:
		*d = FromInt64(v)
		return nil
	case float32:
		*d = FromFloat64(float64(v))
	case float64:
		*d = FromFloat64(v)
	default:
		return fmt.Errorf("cannot scan %T to Optimus", src)
	}
	return d.ErrorDetails()
}

// ScanInto scans a database value into dst using the historical Optimus Scan behavior.
func ScanInto(dst *Optimus, src any) error {
	if dst == nil {
		return fmt.Errorf("optimus: nil scan destination")
	}
	return dst.Scan(src)
}

// ScanDestination returns a database scanner for the provided destination.
func ScanDestination(dst *Optimus) sql.Scanner { return dst }

// Value implements driver.Valuer using decimal text. Legacy NaN values are written as NaN text.
func (d Optimus) Value() (driver.Value, error) { return d.String(), nil }

// EncodeToUint64 truncates or extends to scale p, then returns a nonnegative unsigned coefficient if it fits.
func (d Optimus) EncodeToUint64(p uint8) (uint64, error) {
	if d.IsNaN() {
		return 0, d.ErrorDetails()
	}
	if d.IsNegative() {
		return 0, errors.Negative.Value()
	}
	x := d.Rescale(p)
	if x.IsNaN() {
		return 0, x.ErrorDetails()
	}
	u := x.coefBig()
	if !u.IsUint64() {
		return 0, errors.Overflow.Value()
	}
	return u.Uint64(), nil
}

// EncodeToUint128 truncates or extends to scale p, then returns a nonnegative unsigned coefficient if it fits.
func (d Optimus) EncodeToUint128(p uint8) (uint128.Uint128, error) {
	if d.IsNaN() {
		return uint128.Zero, d.ErrorDetails()
	}
	if d.IsNegative() {
		return uint128.Zero, errors.Negative.Value()
	}
	x := d.Rescale(p)
	if x.IsNaN() {
		return uint128.Zero, x.ErrorDetails()
	}
	u, e := uint128.FromBigInt(x.coefBig())
	if e != errors.None {
		return uint128.Zero, e.Value()
	}
	return u, nil
}

// EncodeToUint256 truncates or extends to scale p, then returns a nonnegative unsigned coefficient if it fits.
func (d Optimus) EncodeToUint256(p uint8) (uint256.Uint256, error) {
	if d.IsNaN() {
		return uint256.Zero, d.ErrorDetails()
	}
	if d.IsNegative() {
		return uint256.Zero, errors.Negative.Value()
	}
	x := d.Rescale(p)
	if x.IsNaN() {
		return uint256.Zero, x.ErrorDetails()
	}
	u, e := uint256.FromBigInt(x.coefBig())
	if e != errors.None {
		return uint256.Zero, e.Value()
	}
	return u, nil
}

// EncodeToUint512 truncates or extends to scale p, then returns a nonnegative unsigned coefficient if it fits.
func (d Optimus) EncodeToUint512(p uint8) (uint512.Uint512, error) {
	if d.IsNaN() {
		return uint512.Zero, d.ErrorDetails()
	}
	if d.IsNegative() {
		return uint512.Zero, errors.Negative.Value()
	}
	x := d.Rescale(p)
	if x.IsNaN() {
		return uint512.Zero, x.ErrorDetails()
	}
	return x.Coefficient(), nil
}

// Int64 truncates to an integer and returns an error if the result does not fit int64.
func (d Optimus) Int64() (int64, error) {
	if d.IsNaN() {
		return 0, d.ErrorDetails()
	}
	v := d.Trunc(0).coefBig()
	if d.neg {
		v.Neg(v)
	}
	if !v.IsInt64() {
		return 0, errors.Overflow.Value()
	}
	return v.Int64(), nil
}

// Int truncates to an integer and returns an error if the result does not fit int.
func (d Optimus) Int() (int, error) {
	v, e := d.Int64()
	if e != nil {
		return 0, e
	}
	if int64(int(v)) != v {
		return 0, errors.Overflow.Value()
	}
	return int(v), nil
}

// InexactFloat64 converts the decimal text to float64, which may lose decimal precision.
func (d Optimus) InexactFloat64() (float64, error) {
	if d.IsNaN() {
		return 0, d.ErrorDetails()
	}
	return strconv.ParseFloat(d.String(), 64)
}

// FromDec128 converts a Dec128 value to Optimus without changing an ordinary numeric value.
func FromDec128(d dec128.Dec128) Optimus {
	if d.IsNaN() {
		return NaN(errors.NotANumber)
	}
	return NewFromUint128(d.Coefficient(), d.Precision(), d.IsNegative())
}

// FromDec256 converts a Dec256 value to Optimus without changing an ordinary numeric value.
func FromDec256(d dec256.Dec256) Optimus {
	if d.IsNaN() {
		return NaN(errors.NotANumber)
	}
	return NewFromUint256(d.Coefficient(), d.Precision(), d.IsNegative())
}

// FromDec512 converts a Dec512 value to Optimus without changing an ordinary numeric value.
func FromDec512(d dec512.Dec512) Optimus {
	if d.IsNaN() {
		return NaN(errors.NotANumber)
	}
	return New(d.Coefficient(), d.Precision(), d.IsNegative())
}

// Dec128Exact converts to Dec128 without rounding and returns an error if the value cannot fit.
func (d Optimus) Dec128Exact() (dec128.Dec128, error) {
	if d.IsNaN() {
		return dec128.Zero, d.ErrorDetails()
	}
	n := decimalwide.Canonical(d.core())
	if n.Scale > int(dec128.MaxPrecision) {
		return dec128.Zero, errors.PrecisionOutOfRange.Value()
	}
	c, e := uint128.FromBigInt(n.Coef)
	if e != errors.None {
		return dec128.Zero, e.Value()
	}
	return dec128.New(c, uint8(n.Scale), n.Neg), nil
}

// Dec256Exact converts to Dec256 without rounding and returns an error if the value cannot fit.
func (d Optimus) Dec256Exact() (dec256.Dec256, error) {
	if d.IsNaN() {
		return dec256.Zero, d.ErrorDetails()
	}
	n := decimalwide.Canonical(d.core())
	if n.Scale > int(dec256.MaxPrecision) {
		return dec256.Zero, errors.PrecisionOutOfRange.Value()
	}
	c, e := uint256.FromBigInt(n.Coef)
	if e != errors.None {
		return dec256.Zero, e.Value()
	}
	return dec256.New(c, uint8(n.Scale), n.Neg), nil
}

// Dec512Exact converts to Dec512 without rounding and returns an error if the value cannot fit.
func (d Optimus) Dec512Exact() (dec512.Dec512, error) {
	if d.IsNaN() {
		return dec512.Zero, d.ErrorDetails()
	}
	return dec512.New(d.Coefficient(), d.exp, d.neg), nil
}
