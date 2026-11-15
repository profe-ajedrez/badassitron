// Package dec512 provides fixed-width decimal arithmetic with a 512-bit
// unsigned coefficient and up to 154 fractional decimal places.
// The API follows dec128; widening does not modify the dec128 package.
package dec512

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync/atomic"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec256"
	"github.com/profe-ajedrez/badassitron/dec256/uint256"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
	"github.com/profe-ajedrez/badassitron/internal/decimalwide"
)

// MaxPrecision is the largest supported decimal scale for this type.
const MaxPrecision uint8 = 154

// MaxStrLen bounds the length of a decimal text representation.
const MaxStrLen = uint512.MaxStrLen + int(MaxPrecision) + 3

// RoundingMode is shared with dec128 so callers can use the same policy.
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

// Dec512 is a fixed-width signed decimal, storing a nonnegative coefficient
// and base-10 fractional scale separately. It is a comparable value type.
type Dec512 struct {
	coef uint512.Uint512
	err  errors.Error
	exp  uint8
	neg  bool
}

var (
	// Zero is the predefined decimal value zero.
	Zero = Dec512{}
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
	NaNJsonStrBytes = []byte(`"NaN"`)
	// Pow10Uint512 holds precomputed powers of ten for this coefficient width. Treat the exported table as read only.
	Pow10Uint512     = uint512.Pow10Uint512
	defaultPrecision atomic.Uint32
)

func init() { defaultPrecision.Store(19) }

// SetDefaultPrecision sets the process wide working scale for default division and square root. It starts at 19, is atomic, and panics if p exceeds MaxPrecision.
func SetDefaultPrecision(prec uint8) {
	if prec > MaxPrecision {
		panic(errors.PrecisionOutOfRange.Value())
	}
	defaultPrecision.Store(uint32(prec))
}

// New constructs a fixed width decimal from a coefficient, decimal scale, and sign. An invalid scale returns NaN.
func New(coef uint512.Uint512, exp uint8, neg bool) Dec512 {
	if exp > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange)
	}
	if coef.IsZero() && exp == 0 {
		return Zero
	}
	return Dec512{coef: coef, exp: exp, neg: neg}
}

// NaN returns an invalid decimal with the requested error reason. A missing reason becomes NotANumber.
func NaN(reason errors.Error) Dec512 {
	if reason == errors.None {
		reason = errors.NotANumber
	}
	return Dec512{err: reason}
}
func (d Dec512) core() decimalwide.Number {
	if d.err != errors.None {
		return decimalwide.Invalid(d.err)
	}
	return decimalwide.New(d.coef.BigInt(), int(d.exp), d.neg)
}
func wrap(n decimalwide.Number) Dec512 {
	if n.Err != errors.None {
		return NaN(n.Err)
	}
	if n.Scale < 0 || n.Scale > int(MaxPrecision) {
		return NaN(errors.PrecisionOutOfRange)
	}
	if n.Coef == nil {
		return NaN(errors.InvalidFormat)
	}
	v, e := uint512.FromBigInt(n.Coef)
	if e != errors.None {
		return NaN(e)
	}
	return New(v, uint8(n.Scale), n.Neg)
}

// IsNaN reports whether the value represents an error rather than an ordinary number.
func (d Dec512) IsNaN() bool { return d.err != errors.None }

// ErrorDetails returns the error carried by a NaN value.
func (d Dec512) ErrorDetails() error { return d.err.Value() }

// IsZero reports whether this is an ordinary numeric zero.
func (d Dec512) IsZero() bool { return d.err == errors.None && d.coef.IsZero() }

// IsNegative reports whether this is an ordinary negative value.
func (d Dec512) IsNegative() bool { return d.err == errors.None && !d.coef.IsZero() && d.neg }

// IsPositive reports whether this is an ordinary positive value.
func (d Dec512) IsPositive() bool { return d.err == errors.None && !d.coef.IsZero() && !d.neg }

// Sign returns minus one for negative values, one for positive values, and zero for zero or NaN.
func (d Dec512) Sign() int {
	if d.IsNaN() || d.IsZero() {
		return 0
	}
	if d.neg {
		return -1
	}
	return 1
}

// Precision returns the stored fractional decimal scale, not the number of significant digits.
func (d Dec512) Precision() uint8 { return d.exp }

// Exponent returns the stored fractional decimal scale, matching Precision.
func (d Dec512) Exponent() uint8 { return d.exp }

// Coefficient returns the unsigned coefficient without the sign or decimal scale.
func (d Dec512) Coefficient() uint512.Uint512 { return d.coef }

// Copy returns an independent value copy.
func (d Dec512) Copy() Dec512 { return d }

// Canonical returns a numerically equal value with trailing decimal zeros removed.
func (d Dec512) Canonical() Dec512 { return wrap(decimalwide.Canonical(d.core())) }

// Compare compares numeric values regardless of trailing zeros or stored scale.
func (d Dec512) Compare(b Dec512) int { return decimalwide.Cmp(d.core(), b.core()) }

// Equal reports numeric equality across different representations. Two NaN values compare equal.
func (d Dec512) Equal(b Dec512) bool {
	if d.IsNaN() || b.IsNaN() {
		return d.IsNaN() && b.IsNaN()
	}
	return d.Compare(b) == 0
}

// LessThan reports whether the receiver sorts before b under Compare.
func (d Dec512) LessThan(b Dec512) bool { return d.Compare(b) < 0 }

// LessThanOrEqual reports whether the receiver sorts before or equal to b under Compare.
func (d Dec512) LessThanOrEqual(b Dec512) bool { return d.Compare(b) <= 0 }

// GreaterThan reports whether the receiver sorts after b under Compare.
func (d Dec512) GreaterThan(b Dec512) bool { return d.Compare(b) > 0 }

// GreaterThanOrEqual reports whether the receiver sorts after or equal to b under Compare.
func (d Dec512) GreaterThanOrEqual(b Dec512) bool { return d.Compare(b) >= 0 }

// FromString parses decimal text with a fixed coefficient limit. Invalid or unrepresentable input returns NaN.
func FromString[S string | []byte](s S) Dec512 {
	return wrap(decimalwide.Parse(string(s), 512, int(MaxPrecision)))
}

// NewFromString parses decimal text and returns an error instead of NaN when parsing fails.
func NewFromString(s string) (Dec512, error) {
	d := FromString(s)
	if d.IsNaN() {
		return Zero, d.ErrorDetails()
	}
	return d, nil
}

// FromInt creates a decimal value from a signed int.
func FromInt(i int) Dec512 { return FromInt64(int64(i)) }

// NewFromInt is an alias for FromInt.
func NewFromInt(i int) Dec512 { return FromInt(i) }

// FromInt64 creates a decimal value from a signed 64 bit integer.
func FromInt64(i int64) Dec512 {
	var v uint64
	neg := i < 0
	if neg {
		v = uint64(-(i + 1)) + 1
	} else {
		v = uint64(i)
	}
	return New(uint512.FromUint64(v), 0, neg)
}

// DecodeFromUint64 constructs a nonnegative decimal from an unsigned coefficient and scale.
func DecodeFromUint64(coef uint64, exp uint8) Dec512 {
	return New(uint512.FromUint64(coef), exp, false)
}

// DecodeFromUint512 constructs a nonnegative decimal from an unsigned coefficient and scale.
func DecodeFromUint512(coef uint512.Uint512, exp uint8) Dec512 { return New(coef, exp, false) }

// FromFloat64 formats a finite binary float as decimal text and parses it. This cannot restore already lost decimal digits.
func FromFloat64(f float64) Dec512 {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return NaN(errors.NotANumber)
	}
	return FromString(strconv.FormatFloat(f, 'f', -1, 64))
}

// Add adds two decimal values without silently rounding a result that exceeds capacity.
func (d Dec512) Add(b Dec512) Dec512 {
	if d.IsNaN() {
		return d
	}
	if b.IsNaN() {
		return b
	}
	p := max(d.exp, b.exp)
	aCoef, bCoef := d.coef, b.coef
	var e errors.Error
	if d.exp < p {
		aCoef, e = aCoef.Mul(Pow10Uint512[p-d.exp])
		if e != errors.None {
			return wrap(decimalwide.Add(d.core(), b.core(), 512, int(MaxPrecision)))
		}
	}
	if b.exp < p {
		bCoef, e = bCoef.Mul(Pow10Uint512[p-b.exp])
		if e != errors.None {
			return wrap(decimalwide.Add(d.core(), b.core(), 512, int(MaxPrecision)))
		}
	}
	if d.neg == b.neg {
		v, e := aCoef.Add(bCoef)
		if e != errors.None {
			return wrap(decimalwide.Add(d.core(), b.core(), 512, int(MaxPrecision)))
		}
		return New(v, p, d.neg)
	}
	switch aCoef.Compare(bCoef) {
	case 0:
		return Zero
	case 1:
		v, _ := aCoef.Sub(bCoef)
		return New(v, p, d.neg)
	default:
		v, _ := bCoef.Sub(aCoef)
		return New(v, p, b.neg)
	}
}

// AddInt adds a signed integer to the receiver.
func (d Dec512) AddInt(b int) Dec512 { return d.Add(FromInt(b)) }

// Sub subtracts b from the receiver.
func (d Dec512) Sub(b Dec512) Dec512 { b.neg = !b.neg; return d.Add(b) }

// SubInt subtracts a signed integer from the receiver.
func (d Dec512) SubInt(b int) Dec512 { return d.Sub(FromInt(b)) }

// Mul multiplies two decimal values without silently rounding an overflowing exact product.
func (d Dec512) Mul(b Dec512) Dec512 {
	if d.IsNaN() {
		return d
	}
	if b.IsNaN() {
		return b
	}
	if d.IsZero() || b.IsZero() {
		return Zero
	}
	lo, hi := d.coef.MulCarry(b.coef)
	p := int(d.exp) + int(b.exp)
	if hi.IsZero() && p <= int(MaxPrecision) {
		return New(lo, uint8(p), d.neg != b.neg)
	}
	return wrap(decimalwide.Mul(d.core(), b.core(), 512, int(MaxPrecision)))
}

// MulInt multiplies the receiver by a signed integer.
func (d Dec512) MulInt(b int) Dec512 { return d.Mul(FromInt(b)) }

// Div divides by b using truncation toward zero at the default working scale.
func (d Dec512) Div(b Dec512) Dec512 {
	p := max(int(d.exp), int(defaultPrecision.Load()))
	result, _ := decimalwide.Div(d.core(), b.core(), p, 512, int(MaxPrecision), RoundingTowardZero)
	return wrap(result)
}

// DivInt divides by a signed integer.
func (d Dec512) DivInt(b int) Dec512 { return d.Div(FromInt(b)) }

// QuoRem returns the integer quotient and signed remainder of truncating division by b.
func (d Dec512) QuoRem(b Dec512) (Dec512, Dec512) {
	q, r := decimalwide.QuoRem(d.core(), b.core(), 512, int(MaxPrecision))
	return wrap(q), wrap(r)
}

// QuoRemInt returns the integer quotient and remainder after division by a signed integer.
func (d Dec512) QuoRemInt(b int) (Dec512, Dec512) { return d.QuoRem(FromInt(b)) }

// Mod returns the remainder after truncating division by b.
func (d Dec512) Mod(b Dec512) Dec512 {
	return wrap(decimalwide.Mod(d.core(), b.core(), 512, int(MaxPrecision)))
}

// ModInt returns the remainder after division by a signed integer.
func (d Dec512) ModInt(b int) Dec512 { return d.Mod(FromInt(b)) }

// Abs returns the absolute value without changing the receiver.
func (d Dec512) Abs() Dec512 {
	if d.IsNaN() {
		return d
	}
	d.neg = false
	return d
}

// Neg returns the additive inverse without changing the receiver.
func (d Dec512) Neg() Dec512 {
	if d.IsNaN() {
		return d
	}
	d.neg = !d.neg
	return d
}

// PowInt raises the value to an integer power using the fixed width decimal rules.
func (d Dec512) PowInt(n int) Dec512 {
	return wrap(decimalwide.PowInt(d.core(), n, 512, int(MaxPrecision), int(defaultPrecision.Load())))
}

// Sqrt computes the square root using truncation toward zero at the default working scale.
func (d Dec512) Sqrt() Dec512 {
	p := max(int(d.exp), int(defaultPrecision.Load()))
	r, _ := decimalwide.Sqrt(d.core(), p, 512, int(MaxPrecision), RoundingTowardZero)
	return wrap(r)
}

// Rescale changes the stored decimal scale, truncating toward zero when digits must be removed.
func (d Dec512) Rescale(p uint8) Dec512 { return d.RescaleRound(p, RoundingTowardZero) }

// RescaleRound changes the scale using the requested rounding mode.
func (d Dec512) RescaleRound(p uint8, m RoundingMode) Dec512 {
	r, _ := d.RescaleRoundWithStatus(p, m)
	return r
}

// RescaleRoundWithStatus changes the scale and reports whether nonzero digits were discarded. An error is carried in NaN, not in the boolean.
func (d Dec512) RescaleRoundWithStatus(p uint8, m RoundingMode) (Dec512, bool) {
	if d.IsNaN() {
		return d, false
	}
	if p > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange), false
	}
	if m > RoundingHalfEven {
		return NaN(errors.InvalidFormat), false
	}
	if p == d.exp {
		return d, false
	}
	if p > d.exp {
		coef, e := d.coef.Mul(Pow10Uint512[p-d.exp])
		if e != errors.None {
			return NaN(e), false
		}
		return New(coef, p, d.neg), false
	}
	den := Pow10Uint512[d.exp-p]
	q, r, e := d.coef.QuoRem(den)
	if e != errors.None {
		return NaN(e), false
	}
	rounded, inexact, e := roundNative(q, r, den, d.neg, m)
	if e != errors.None {
		return NaN(e), inexact
	}
	return New(rounded, p, d.neg && !rounded.IsZero()), inexact
}

// MulRound multiplies and rounds the exact product to the requested scale.
func (d Dec512) MulRound(b Dec512, p uint8, m RoundingMode) Dec512 {
	r, _ := d.MulRoundWithStatus(b, p, m)
	return r
}

// MulRoundWithStatus multiplies with explicit rounding and reports whether nonzero digits were discarded. Check NaN for errors.
func (d Dec512) MulRoundWithStatus(b Dec512, p uint8, m RoundingMode) (Dec512, bool) {
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
	lo, hi := d.coef.MulCarry(b.coef)
	sumScale := int(d.exp) + int(b.exp)
	if hi.IsZero() && sumScale <= int(MaxPrecision) {
		return New(lo, uint8(sumScale), d.neg != b.neg).RescaleRoundWithStatus(p, m)
	}
	// The product can require up to 1024 bits, even when the rounded
	// result fits. Preserve it exactly until the final quantization.
	a, c := d.core(), b.core()
	product := decimalwide.New(new(big.Int).Mul(a.Coef, c.Coef), a.Scale+c.Scale, a.Neg != c.Neg)
	r, inexact := decimalwide.Quantize(product, int(p), 512, int(MaxPrecision), m)
	return wrap(r), inexact
}

// DivRound divides and rounds to the requested scale and rounding mode.
func (d Dec512) DivRound(b Dec512, p uint8, m RoundingMode) Dec512 {
	r, _ := d.DivRoundWithStatus(b, p, m)
	return r
}

// DivRoundWithStatus divides with explicit rounding and reports whether nonzero information was discarded. Check NaN for errors.
func (d Dec512) DivRoundWithStatus(b Dec512, p uint8, m RoundingMode) (Dec512, bool) {
	if d.IsNaN() {
		return d, false
	}
	if b.IsNaN() {
		return b, false
	}
	if b.IsZero() {
		return NaN(errors.DivisionByZero), false
	}
	if p > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange), false
	}
	if m > RoundingHalfEven {
		return NaN(errors.InvalidFormat), false
	}
	numerator, denominator := d.coef, b.coef
	shift := int(b.exp) + int(p) - int(d.exp)
	ok := true
	var e errors.Error
	if shift >= 0 {
		if shift > int(MaxPrecision) {
			ok = false
		} else {
			numerator, e = numerator.Mul(Pow10Uint512[shift])
			ok = e == errors.None
		}
	} else {
		if -shift > int(MaxPrecision) {
			ok = false
		} else {
			denominator, e = denominator.Mul(Pow10Uint512[-shift])
			ok = e == errors.None
		}
	}
	if ok {
		q, r, e := numerator.QuoRem(denominator)
		if e == errors.None {
			neg := d.neg != b.neg
			result, inexact, err := roundNative(q, r, denominator, neg, m)
			if err != errors.None {
				return NaN(err), inexact
			}
			return New(result, p, neg && !result.IsZero()), inexact
		}
	}
	r, inexact := decimalwide.Div(d.core(), b.core(), int(p), 512, int(MaxPrecision), m)
	return wrap(r), inexact
}

// SqrtRound computes the square root rounded to the requested decimal scale.
func (d Dec512) SqrtRound(p uint8, m RoundingMode) Dec512 {
	r, _ := d.SqrtRoundWithStatus(p, m)
	return r
}

// SqrtRoundWithStatus computes a square root with explicit rounding and reports inexactness separately from errors.
func (d Dec512) SqrtRoundWithStatus(p uint8, m RoundingMode) (Dec512, bool) {
	r, inexact := decimalwide.Sqrt(d.core(), int(p), 512, int(MaxPrecision), m)
	return wrap(r), inexact
}

// Round rounds to the nearest value at scale p, with ties away from zero.
func (d Dec512) Round(p uint8) Dec512 { return d.RoundHalfAwayFromZero(p) }

// RoundDown rounds to scale p toward negative infinity.
func (d Dec512) RoundDown(p uint8) Dec512 { return d.roundOnly(p, RoundingFloor) }

// RoundUp rounds to scale p toward positive infinity.
func (d Dec512) RoundUp(p uint8) Dec512 { return d.roundOnly(p, RoundingCeil) }

// RoundTowardZero rounds to scale p toward zero.
func (d Dec512) RoundTowardZero(p uint8) Dec512 { return d.roundOnly(p, RoundingTowardZero) }

// RoundAwayFromZero rounds to scale p away from zero.
func (d Dec512) RoundAwayFromZero(p uint8) Dec512 { return d.roundOnly(p, RoundingAwayFromZero) }

// RoundHalfTowardZero rounds to scale p to the nearest value, with ties toward zero.
func (d Dec512) RoundHalfTowardZero(p uint8) Dec512 { return d.roundOnly(p, RoundingHalfTowardZero) }

// RoundHalfAwayFromZero rounds to scale p to the nearest value, with ties away from zero.
func (d Dec512) RoundHalfAwayFromZero(p uint8) Dec512 {
	return d.roundOnly(p, RoundingHalfAwayFromZero)
}

// RoundBank rounds to scale p to the nearest value, with ties to even.
func (d Dec512) RoundBank(p uint8) Dec512 { return d.roundOnly(p, RoundingHalfEven) }

// Trunc discards fractional digits beyond scale p without rounding away from zero.
func (d Dec512) Trunc(p uint8) Dec512 { return d.roundOnly(p, RoundingTowardZero) }
func (d Dec512) roundOnly(p uint8, m RoundingMode) Dec512 {
	if d.IsNaN() || p >= d.exp {
		return d
	}
	return d.RescaleRound(p, m)
}

// Min returns the smallest value under Compare.
func Min(a Dec512, rest ...Dec512) Dec512 {
	for _, x := range rest {
		if x.Compare(a) < 0 {
			a = x
		}
	}
	return a
}

// Max returns the largest value under Compare.
func Max(a Dec512, rest ...Dec512) Dec512 {
	for _, x := range rest {
		if x.Compare(a) > 0 {
			a = x
		}
	}
	return a
}

// Sum adds the first value and all remaining values.
func Sum(a Dec512, rest ...Dec512) Dec512 {
	for _, x := range rest {
		a = a.Add(x)
	}
	return a
}

// Avg divides the sum by the number of values using the fixed width default division rule.
func Avg(a Dec512, rest ...Dec512) Dec512 {
	n := 1 + len(rest)
	return Sum(a, rest...).DivInt(n)
}

// String returns decimal text without insignificant trailing fractional zeros. NaN is written as NaN.
func (d Dec512) String() string { return decimalwide.Format(d.core(), false) }

// StringFixed returns decimal text retaining the stored fractional scale.
func (d Dec512) StringFixed() string { return decimalwide.Format(d.core(), true) }

// MarshalText implements encoding.TextMarshaler using the decimal String representation.
func (d Dec512) MarshalText() ([]byte, error) { return []byte(d.String()), nil }

// UnmarshalText implements encoding.TextUnmarshaler. Invalid text leaves the destination unchanged.
func (d *Dec512) UnmarshalText(s []byte) error {
	v := FromString(s)
	if v.IsNaN() {
		return v.ErrorDetails()
	}
	*d = v
	return nil
}

// MarshalJSON encodes the decimal as a JSON string to avoid numeric decoding through float64.
func (d Dec512) MarshalJSON() ([]byte, error) { return []byte(strconv.Quote(d.String())), nil }

// UnmarshalJSON accepts a JSON decimal string or number. JSON null becomes numeric zero.
func (d *Dec512) UnmarshalJSON(s []byte) error {
	if bytes.Equal(s, []byte("null")) {
		*d = Zero
		return nil
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		v, err := strconv.Unquote(string(s))
		if err != nil {
			return err
		}
		return d.UnmarshalText([]byte(v))
	}
	return d.UnmarshalText(s)
}

// Scan implements sql.Scanner with the accepted legacy database value types. NULL becomes zero.
func (d *Dec512) Scan(src any) error {
	if d == nil {
		return fmt.Errorf("dec512: nil scan destination")
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
		return fmt.Errorf("cannot scan %T to Dec512", src)
	}
	if d.IsNaN() {
		return d.ErrorDetails()
	}
	return nil
}

// Value implements driver.Valuer and returns decimal text.
func (d Dec512) Value() (driver.Value, error) { return d.String(), nil }

// EncodeToUint64 truncates or extends to scale p, then returns a nonnegative unsigned coefficient if it fits.
func (d Dec512) EncodeToUint64(p uint8) (uint64, error) {
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
	u, e := x.coef.Uint64()
	return u, e.Value()
}

// EncodeToUint512 rescales and returns the unsigned coefficient at scale p, reporting invalid or negative values.
func (d Dec512) EncodeToUint512(p uint8) (uint512.Uint512, error) {
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
	return x.coef, nil
}

// Int64 truncates to an integer and returns an error if the result does not fit int64.
func (d Dec512) Int64() (int64, error) {
	if d.IsNaN() {
		return 0, d.ErrorDetails()
	}
	x := d.Trunc(0)
	if x.IsNaN() {
		return 0, x.ErrorDetails()
	}
	value := x.coef.BigInt()
	if x.neg {
		value.Neg(value)
	}
	if !value.IsInt64() {
		return 0, errors.Overflow.Value()
	}
	return value.Int64(), nil
}

// Int truncates to an integer and returns an error if the result does not fit int.
func (d Dec512) Int() (int, error) {
	i, e := d.Int64()
	if e != nil {
		return 0, e
	}
	if int64(int(i)) != i {
		return 0, errors.Overflow.Value()
	}
	return int(i), nil
}

// InexactFloat64 converts the decimal text to float64, which may lose decimal precision.
func (d Dec512) InexactFloat64() (float64, error) {
	if d.IsNaN() {
		return 0, d.ErrorDetails()
	}
	return strconv.ParseFloat(d.String(), 64)
}

// roundNative compares the remainder to half its divisor without calculating
// 2*remainder, which could overflow the fixed-width representation.
func roundNative(q, r, den uint512.Uint512, neg bool, mode RoundingMode) (uint512.Uint512, bool, errors.Error) {
	inexact := !r.IsZero()
	if !inexact {
		return q, false, errors.None
	}
	up := false
	switch mode {
	case RoundingTowardZero:
	case RoundingAwayFromZero:
		up = true
	case RoundingFloor:
		up = neg
	case RoundingCeil:
		up = !neg
	case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
		complement, _ := den.Sub(r)
		cmp := r.Compare(complement)
		lsb := q.Lo.Lo.Lo & 1
		up = cmp > 0 || cmp == 0 && (mode == RoundingHalfAwayFromZero || mode == RoundingHalfEven && lsb == 1)
	}
	if up {
		v, e := q.Add64(1)
		return v, true, e
	}
	return q, true, errors.None
}

// FromDec128 widens an existing decimal without losing scale or value.
// NaN is propagated as NotANumber because Dec128 exposes only an error value.
func FromDec128(source dec128.Dec128) Dec512 {
	if source.IsNaN() {
		return NaN(errors.NotANumber)
	}
	coef := uint512.FromUint256(uint256.FromUint128(source.Coefficient()))
	return New(coef, source.Precision(), source.IsNegative())
}

// Dec128Exact narrows a wide decimal only when no information is lost.
// Trailing zeroes may be removed before narrowing.
func (d Dec512) Dec128Exact() (dec128.Dec128, error) {
	if d.IsNaN() {
		return dec128.Zero, d.ErrorDetails()
	}
	d = d.Canonical()
	if d.exp > dec128.MaxPrecision {
		return dec128.Zero, errors.PrecisionOutOfRange.Value()
	}
	bigValue := d.coef.BigInt()
	coef, e := uint128.FromBigInt(new(big.Int).Set(bigValue))
	if e != errors.None {
		return dec128.Zero, e.Value()
	}
	return dec128.New(coef, d.exp, d.neg), nil
}

// ScanInto and ScanDestination match the v4 Dec128 SQL helper naming.
// Unlike historical Dec128.Scan, (*Dec512).Scan directly updates its receiver.
func ScanInto(dst *Dec512, source any) error {
	if dst == nil {
		return fmt.Errorf("dec512: nil scan destination")
	}
	return dst.Scan(source)
}

// ScanDestination returns a database scanner for the provided destination.
func ScanDestination(dst *Dec512) sql.Scanner { return dst }

// FromDec256 widens a Dec256 without losing fractional precision.
func FromDec256(source dec256.Dec256) Dec512 {
	if source.IsNaN() {
		return NaN(errors.NotANumber)
	}
	return New(uint512.FromUint256(source.Coefficient()), source.Precision(), source.IsNegative())
}

// Dec256Exact returns a Dec256 iff the value can be represented exactly.
func (d Dec512) Dec256Exact() (dec256.Dec256, error) {
	if d.IsNaN() {
		return dec256.Zero, d.ErrorDetails()
	}
	d = d.Canonical()
	if d.exp > dec256.MaxPrecision {
		return dec256.Zero, errors.PrecisionOutOfRange.Value()
	}
	coef, e := d.coef.Uint256()
	if e != errors.None {
		return dec256.Zero, e.Value()
	}
	return dec256.New(coef, d.exp, d.neg), nil
}
