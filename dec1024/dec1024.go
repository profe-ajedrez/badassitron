// Package dec1024 provides fixed-width decimal arithmetic with a 1024-bit
// unsigned coefficient and up to 308 fractional decimal places.
// The API follows dec128; widening does not modify the dec128 package.
package dec1024

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"
	"sync/atomic"

	"github.com/profe-ajedrez/badassitron/dec1024/uint1024"
	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec256"
	"github.com/profe-ajedrez/badassitron/dec256/uint256"
	"github.com/profe-ajedrez/badassitron/dec512"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

// MaxPrecision is the largest supported decimal scale for this type.
const MaxPrecision uint16 = 308

// MaxStrLen bounds the length of a decimal text representation.
const MaxStrLen = uint1024.MaxStrLen + int(MaxPrecision) + 3

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

// Dec1024 is a fixed-width signed decimal, storing a nonnegative coefficient
// and base-10 fractional scale separately. It is a comparable value type.
type Dec1024 struct {
	coef uint1024.Uint1024
	err  errors.Error
	exp  uint16
	neg  bool
}

var (
	// Zero is the predefined decimal value zero.
	Zero = Dec1024{}
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
	// Pow10Uint1024 holds powers of ten up to the maximum decimal scale. Treat the exported table as read only.
	Pow10Uint1024    = uint1024.Pow10Uint1024
	defaultPrecision atomic.Uint32
)

func init() { defaultPrecision.Store(19) }

// SetDefaultPrecision sets the process wide working scale for default division and square root. It starts at 19, is atomic, and panics if p exceeds MaxPrecision.
func SetDefaultPrecision(prec uint16) {
	if prec > MaxPrecision {
		panic(errors.PrecisionOutOfRange.Value())
	}
	defaultPrecision.Store(uint32(prec))
}

// New constructs a fixed width decimal from a coefficient, decimal scale, and sign. An invalid scale returns NaN.
func New(coef uint1024.Uint1024, exp uint16, neg bool) Dec1024 {
	if exp > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange)
	}
	if coef.IsZero() && exp == 0 {
		return Zero
	}
	return Dec1024{coef: coef, exp: exp, neg: neg}
}

// NaN returns an invalid decimal with the requested error reason. A missing reason becomes NotANumber.
func NaN(reason errors.Error) Dec1024 {
	if reason == errors.None {
		reason = errors.NotANumber
	}
	return Dec1024{err: reason}
}

// IsNaN reports whether the value represents an error rather than an ordinary number.
func (d Dec1024) IsNaN() bool { return d.err != errors.None }

// ErrorDetails returns the error carried by a NaN value.
func (d Dec1024) ErrorDetails() error { return d.err.Value() }

// IsZero reports whether this is an ordinary numeric zero.
func (d Dec1024) IsZero() bool { return d.err == errors.None && d.coef.IsZero() }

// IsNegative reports whether this is an ordinary negative value.
func (d Dec1024) IsNegative() bool { return d.err == errors.None && !d.coef.IsZero() && d.neg }

// IsPositive reports whether this is an ordinary positive value.
func (d Dec1024) IsPositive() bool { return d.err == errors.None && !d.coef.IsZero() && !d.neg }

// Sign returns minus one for negative values, one for positive values, and zero for zero or NaN.
func (d Dec1024) Sign() int {
	if d.IsNaN() || d.IsZero() {
		return 0
	}
	if d.neg {
		return -1
	}
	return 1
}

// Precision returns the stored fractional decimal scale, not the number of significant digits.
func (d Dec1024) Precision() uint16 { return d.exp }

// Exponent returns the stored fractional decimal scale, matching Precision.
func (d Dec1024) Exponent() uint16 { return d.exp }

// Coefficient returns the unsigned coefficient without the sign or decimal scale.
func (d Dec1024) Coefficient() uint1024.Uint1024 { return d.coef }

// Copy returns an independent value copy.
func (d Dec1024) Copy() Dec1024 { return d }

// Canonical returns a numerically equal value with trailing decimal zeros removed.
func (d Dec1024) Canonical() Dec1024 {
	if d.IsNaN() {
		return d
	}
	if d.coef.IsZero() {
		return Zero
	}
	for d.exp > 0 {
		q, rem, _ := d.coef.QuoRem64(10)
		if rem != 0 {
			break
		}
		d.coef, d.exp = q, d.exp-1
	}
	return d
}

// Compare compares numeric values regardless of trailing zeros or stored scale.
func (d Dec1024) Compare(b Dec1024) int { return fixedCompare(d, b) }

// Equal reports numeric equality across different representations. Two NaN values compare equal.
func (d Dec1024) Equal(b Dec1024) bool {
	if d.IsNaN() || b.IsNaN() {
		return d.IsNaN() && b.IsNaN()
	}
	return d.Compare(b) == 0
}

// LessThan reports whether the receiver sorts before b under Compare.
func (d Dec1024) LessThan(b Dec1024) bool { return d.Compare(b) < 0 }

// LessThanOrEqual reports whether the receiver sorts before or equal to b under Compare.
func (d Dec1024) LessThanOrEqual(b Dec1024) bool { return d.Compare(b) <= 0 }

// GreaterThan reports whether the receiver sorts after b under Compare.
func (d Dec1024) GreaterThan(b Dec1024) bool { return d.Compare(b) > 0 }

// GreaterThanOrEqual reports whether the receiver sorts after or equal to b under Compare.
func (d Dec1024) GreaterThanOrEqual(b Dec1024) bool { return d.Compare(b) >= 0 }

// FromString parses decimal text with a fixed coefficient limit. Invalid or unrepresentable input returns NaN.
func FromString[S string | []byte](s S) Dec1024 {
	return fixedParse(s)
}

// NewFromString parses decimal text and returns an error instead of NaN when parsing fails.
func NewFromString(s string) (Dec1024, error) {
	d := FromString(s)
	if d.IsNaN() {
		return Zero, d.ErrorDetails()
	}
	return d, nil
}

// FromInt creates a decimal value from a signed int.
func FromInt(i int) Dec1024 { return FromInt64(int64(i)) }

// NewFromInt is an alias for FromInt.
func NewFromInt(i int) Dec1024 { return FromInt(i) }

// FromInt64 creates a decimal value from a signed 64 bit integer.
func FromInt64(i int64) Dec1024 {
	var v uint64
	neg := i < 0
	if neg {
		v = uint64(-(i + 1)) + 1
	} else {
		v = uint64(i)
	}
	return New(uint1024.FromUint64(v), 0, neg)
}

// DecodeFromUint64 constructs a nonnegative decimal from an unsigned coefficient and scale.
func DecodeFromUint64(coef uint64, exp uint16) Dec1024 {
	return New(uint1024.FromUint64(coef), exp, false)
}

// DecodeFromUint1024 constructs a nonnegative decimal from an unsigned coefficient and scale.
func DecodeFromUint1024(coef uint1024.Uint1024, exp uint16) Dec1024 { return New(coef, exp, false) }

// FromFloat64 formats a finite binary float as decimal text and parses it. This cannot restore already lost decimal digits.
func FromFloat64(f float64) Dec1024 {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return NaN(errors.NotANumber)
	}
	return FromString(strconv.FormatFloat(f, 'f', -1, 64))
}

// Add adds two decimal values without silently rounding a result that exceeds capacity.
func (d Dec1024) Add(b Dec1024) Dec1024 {
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
		aCoef, e = aCoef.Mul(Pow10Uint1024[p-d.exp])
		if e != errors.None {
			return fixedAdd(d, b)
		}
	}
	if b.exp < p {
		bCoef, e = bCoef.Mul(Pow10Uint1024[p-b.exp])
		if e != errors.None {
			return fixedAdd(d, b)
		}
	}
	if d.neg == b.neg {
		v, e := aCoef.Add(bCoef)
		if e != errors.None {
			return fixedAdd(d, b)
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
func (d Dec1024) AddInt(b int) Dec1024 { return d.Add(FromInt(b)) }

// Sub subtracts b from the receiver.
func (d Dec1024) Sub(b Dec1024) Dec1024 { b.neg = !b.neg; return d.Add(b) }

// SubInt subtracts a signed integer from the receiver.
func (d Dec1024) SubInt(b int) Dec1024 { return d.Sub(FromInt(b)) }

// Mul multiplies two decimal values without silently rounding an overflowing exact product.
func (d Dec1024) Mul(b Dec1024) Dec1024 {
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
	scale := int(d.exp) + int(b.exp)
	if hi.IsZero() && scale <= int(MaxPrecision) {
		return New(lo, uint16(scale), d.neg != b.neg)
	}
	return fixedFitExact(fixedProduct(d.coef, b.coef), scale, d.neg != b.neg)
}

// MulInt multiplies the receiver by a signed integer.
func (d Dec1024) MulInt(b int) Dec1024 { return d.Mul(FromInt(b)) }

// Div divides by b using truncation toward zero at the default working scale.
func (d Dec1024) Div(b Dec1024) Dec1024 {
	p := uint16(max(int(d.exp), int(defaultPrecision.Load())))
	return d.DivRound(b, p, RoundingTowardZero)
}

// DivInt divides by a signed integer.
func (d Dec1024) DivInt(b int) Dec1024 { return d.Div(FromInt(b)) }

// QuoRem returns the integer quotient and signed remainder of truncating division by b.
func (d Dec1024) QuoRem(b Dec1024) (Dec1024, Dec1024) {
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
	return fixedQuoRem(d, b)
}

// QuoRemInt returns the integer quotient and remainder after division by a signed integer.
func (d Dec1024) QuoRemInt(b int) (Dec1024, Dec1024) { return d.QuoRem(FromInt(b)) }

// Mod returns the remainder after truncating division by b.
func (d Dec1024) Mod(b Dec1024) Dec1024 {
	if d.IsNaN() {
		return d
	}
	if b.IsNaN() {
		return b
	}
	if b.IsZero() {
		return NaN(errors.DivisionByZero)
	}
	return fixedMod(d, b)
}

// ModInt returns the remainder after division by a signed integer.
func (d Dec1024) ModInt(b int) Dec1024 { return d.Mod(FromInt(b)) }

// Abs returns the absolute value without changing the receiver.
func (d Dec1024) Abs() Dec1024 {
	if d.IsNaN() {
		return d
	}
	d.neg = false
	return d
}

// Neg returns the additive inverse without changing the receiver.
func (d Dec1024) Neg() Dec1024 {
	if d.IsNaN() {
		return d
	}
	d.neg = !d.neg
	return d
}

// PowInt raises the value to an integer power using the fixed width decimal rules.
func (d Dec1024) PowInt(n int) Dec1024 {
	return fixedPow(d, n)
}

// Sqrt computes the square root using truncation toward zero at the default working scale.
func (d Dec1024) Sqrt() Dec1024 {
	p := uint16(max(int(d.exp), int(defaultPrecision.Load())))
	return d.SqrtRound(p, RoundingTowardZero)
}

// Rescale changes the stored decimal scale, truncating toward zero when digits must be removed.
func (d Dec1024) Rescale(p uint16) Dec1024 { return d.RescaleRound(p, RoundingTowardZero) }

// RescaleRound changes the scale using the requested rounding mode.
func (d Dec1024) RescaleRound(p uint16, m RoundingMode) Dec1024 {
	r, _ := d.RescaleRoundWithStatus(p, m)
	return r
}

// RescaleRoundWithStatus changes the scale and reports whether nonzero digits were discarded. An error is carried in NaN, not in the boolean.
func (d Dec1024) RescaleRoundWithStatus(p uint16, m RoundingMode) (Dec1024, bool) {
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
		coef, e := d.coef.Mul(Pow10Uint1024[p-d.exp])
		if e != errors.None {
			return NaN(e), false
		}
		return New(coef, p, d.neg), false
	}
	den := Pow10Uint1024[d.exp-p]
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
func (d Dec1024) MulRound(b Dec1024, p uint16, m RoundingMode) Dec1024 {
	r, _ := d.MulRoundWithStatus(b, p, m)
	return r
}

// MulRoundWithStatus multiplies with explicit rounding and reports whether nonzero digits were discarded. Check NaN for errors.
func (d Dec1024) MulRoundWithStatus(b Dec1024, p uint16, m RoundingMode) (Dec1024, bool) {
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
	scale := int(d.exp) + int(b.exp)
	if hi.IsZero() && scale <= int(MaxPrecision) && scale <= int(p) {
		return New(lo, uint16(scale), d.neg != b.neg).RescaleRoundWithStatus(p, m)
	}
	return fixedQuantize(fixedFromProduct(lo, hi), scale, p, d.neg != b.neg, m)
}

// DivRound divides and rounds to the requested scale and rounding mode.
func (d Dec1024) DivRound(b Dec1024, p uint16, m RoundingMode) Dec1024 {
	r, _ := d.DivRoundWithStatus(b, p, m)
	return r
}

// DivRoundWithStatus divides with explicit rounding and reports whether nonzero information was discarded. Check NaN for errors.
func (d Dec1024) DivRoundWithStatus(b Dec1024, p uint16, m RoundingMode) (Dec1024, bool) {
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
			numerator, e = numerator.Mul(Pow10Uint1024[shift])
			ok = e == errors.None
		}
	} else {
		if -shift > int(MaxPrecision) {
			ok = false
		} else {
			denominator, e = denominator.Mul(Pow10Uint1024[-shift])
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
	return fixedDivRound(d, b, p, m)
}

// SqrtRound computes the square root rounded to the requested decimal scale.
func (d Dec1024) SqrtRound(p uint16, m RoundingMode) Dec1024 {
	r, _ := d.SqrtRoundWithStatus(p, m)
	return r
}

// SqrtRoundWithStatus computes a square root with explicit rounding and reports inexactness separately from errors.
func (d Dec1024) SqrtRoundWithStatus(p uint16, m RoundingMode) (Dec1024, bool) {
	return fixedSqrt(d, p, m)
}

// Round rounds to the nearest value at scale p, with ties away from zero.
func (d Dec1024) Round(p uint16) Dec1024 { return d.RoundHalfAwayFromZero(p) }

// RoundDown rounds to scale p toward negative infinity.
func (d Dec1024) RoundDown(p uint16) Dec1024 { return d.roundOnly(p, RoundingFloor) }

// RoundUp rounds to scale p toward positive infinity.
func (d Dec1024) RoundUp(p uint16) Dec1024 { return d.roundOnly(p, RoundingCeil) }

// RoundTowardZero rounds to scale p toward zero.
func (d Dec1024) RoundTowardZero(p uint16) Dec1024 { return d.roundOnly(p, RoundingTowardZero) }

// RoundAwayFromZero rounds to scale p away from zero.
func (d Dec1024) RoundAwayFromZero(p uint16) Dec1024 { return d.roundOnly(p, RoundingAwayFromZero) }

// RoundHalfTowardZero rounds to scale p to the nearest value, with ties toward zero.
func (d Dec1024) RoundHalfTowardZero(p uint16) Dec1024 { return d.roundOnly(p, RoundingHalfTowardZero) }

// RoundHalfAwayFromZero rounds to scale p to the nearest value, with ties away from zero.
func (d Dec1024) RoundHalfAwayFromZero(p uint16) Dec1024 {
	return d.roundOnly(p, RoundingHalfAwayFromZero)
}

// RoundBank rounds to scale p to the nearest value, with ties to even.
func (d Dec1024) RoundBank(p uint16) Dec1024 { return d.roundOnly(p, RoundingHalfEven) }

// Trunc discards fractional digits beyond scale p without rounding away from zero.
func (d Dec1024) Trunc(p uint16) Dec1024 { return d.roundOnly(p, RoundingTowardZero) }
func (d Dec1024) roundOnly(p uint16, m RoundingMode) Dec1024 {
	if d.IsNaN() || p >= d.exp {
		return d
	}
	return d.RescaleRound(p, m)
}

// Min returns the smallest value under Compare.
func Min(a Dec1024, rest ...Dec1024) Dec1024 {
	for _, x := range rest {
		if x.Compare(a) < 0 {
			a = x
		}
	}
	return a
}

// Max returns the largest value under Compare.
func Max(a Dec1024, rest ...Dec1024) Dec1024 {
	for _, x := range rest {
		if x.Compare(a) > 0 {
			a = x
		}
	}
	return a
}

// Sum adds the first value and all remaining values.
func Sum(a Dec1024, rest ...Dec1024) Dec1024 {
	for _, x := range rest {
		a = a.Add(x)
	}
	return a
}

// Avg divides the sum by the number of values using the fixed width default division rule.
func Avg(a Dec1024, rest ...Dec1024) Dec1024 {
	n := 1 + len(rest)
	return Sum(a, rest...).DivInt(n)
}

// String returns decimal text without insignificant trailing fractional zeros. NaN is written as NaN.
func (d Dec1024) String() string { return fixedFormat(d, false) }

// StringFixed returns decimal text retaining the stored fractional scale.
func (d Dec1024) StringFixed() string { return fixedFormat(d, true) }

// MarshalText implements encoding.TextMarshaler using the decimal String representation.
func (d Dec1024) MarshalText() ([]byte, error) { return []byte(d.String()), nil }

// UnmarshalText implements encoding.TextUnmarshaler. Invalid text leaves the destination unchanged.
func (d *Dec1024) UnmarshalText(s []byte) error {
	v := FromString(s)
	if v.IsNaN() {
		return v.ErrorDetails()
	}
	*d = v
	return nil
}

// MarshalJSON encodes the decimal as a JSON string to avoid numeric decoding through float64.
func (d Dec1024) MarshalJSON() ([]byte, error) { return []byte(strconv.Quote(d.String())), nil }

// UnmarshalJSON accepts a JSON decimal string or number. JSON null becomes numeric zero.
func (d *Dec1024) UnmarshalJSON(s []byte) error {
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
func (d *Dec1024) Scan(src any) error {
	if d == nil {
		return fmt.Errorf("dec1024: nil scan destination")
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
		return fmt.Errorf("cannot scan %T to Dec1024", src)
	}
	if d.IsNaN() {
		return d.ErrorDetails()
	}
	return nil
}

// Value implements driver.Valuer and returns decimal text.
func (d Dec1024) Value() (driver.Value, error) { return d.String(), nil }

// EncodeToUint64 truncates or extends to scale p, then returns a nonnegative unsigned coefficient if it fits.
func (d Dec1024) EncodeToUint64(p uint16) (uint64, error) {
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

// EncodeToUint1024 rescales and returns the unsigned coefficient at scale p, reporting invalid or negative values.
func (d Dec1024) EncodeToUint1024(p uint16) (uint1024.Uint1024, error) {
	if d.IsNaN() {
		return uint1024.Zero, d.ErrorDetails()
	}
	if d.IsNegative() {
		return uint1024.Zero, errors.Negative.Value()
	}
	x := d.Rescale(p)
	if x.IsNaN() {
		return uint1024.Zero, x.ErrorDetails()
	}
	return x.coef, nil
}

// Int64 truncates to an integer and returns an error if the result does not fit int64.
func (d Dec1024) Int64() (int64, error) {
	if d.IsNaN() {
		return 0, d.ErrorDetails()
	}
	x := d.Trunc(0)
	if x.IsNaN() {
		return 0, x.ErrorDetails()
	}
	value, e := x.coef.Uint64()
	if e != errors.None {
		return 0, errors.Overflow.Value()
	}
	if x.neg {
		if value > uint64(1)<<63 {
			return 0, errors.Overflow.Value()
		}
		if value == uint64(1)<<63 {
			return math.MinInt64, nil
		}
		return -int64(value), nil
	}
	if value > math.MaxInt64 {
		return 0, errors.Overflow.Value()
	}
	return int64(value), nil
}

// Int truncates to an integer and returns an error if the result does not fit int.
func (d Dec1024) Int() (int, error) {
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
func (d Dec1024) InexactFloat64() (float64, error) {
	if d.IsNaN() {
		return 0, d.ErrorDetails()
	}
	return strconv.ParseFloat(d.String(), 64)
}

// roundNative compares the remainder to half its divisor without calculating
// 2*remainder, which could overflow the fixed-width representation.
func roundNative(q, r, den uint1024.Uint1024, neg bool, mode RoundingMode) (uint1024.Uint1024, bool, errors.Error) {
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
		lsb := q.Lo.Lo.Lo.Lo & 1
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
func FromDec128(source dec128.Dec128) Dec1024 {
	if source.IsNaN() {
		return NaN(errors.NotANumber)
	}
	coef := uint1024.FromUint512(uint512.FromUint256(uint256.FromUint128(source.Coefficient())))
	return New(coef, uint16(source.Precision()), source.IsNegative())
}

// Dec128Exact narrows a wide decimal only when no information is lost.
// Trailing zeroes may be removed before narrowing.
func (d Dec1024) Dec128Exact() (dec128.Dec128, error) {
	if d.IsNaN() {
		return dec128.Zero, d.ErrorDetails()
	}
	d = d.Canonical()
	if d.exp > uint16(dec128.MaxPrecision) {
		return dec128.Zero, errors.PrecisionOutOfRange.Value()
	}
	u512, e := d.coef.Uint512()
	if e != errors.None {
		return dec128.Zero, e.Value()
	}
	u256, e := u512.Uint256()
	if e != errors.None {
		return dec128.Zero, e.Value()
	}
	coef, e := u256.Uint128()
	if e != errors.None {
		return dec128.Zero, e.Value()
	}
	return dec128.New(coef, uint8(d.exp), d.neg), nil
}

// ScanInto and ScanDestination match the v4 Dec128 SQL helper naming.
// Unlike historical Dec128.Scan, (*Dec1024).Scan directly updates its receiver.
func ScanInto(dst *Dec1024, source any) error {
	if dst == nil {
		return fmt.Errorf("dec1024: nil scan destination")
	}
	return dst.Scan(source)
}

// ScanDestination returns a database scanner for the provided destination.
func ScanDestination(dst *Dec1024) sql.Scanner { return dst }

// FromDec256 widens a Dec256 without losing fractional precision.
func FromDec256(source dec256.Dec256) Dec1024 {
	if source.IsNaN() {
		return NaN(errors.NotANumber)
	}
	return New(uint1024.FromUint512(uint512.FromUint256(source.Coefficient())), uint16(source.Precision()), source.IsNegative())
}

// Dec256Exact returns a Dec256 iff the value can be represented exactly.
func (d Dec1024) Dec256Exact() (dec256.Dec256, error) {
	if d.IsNaN() {
		return dec256.Zero, d.ErrorDetails()
	}
	d = d.Canonical()
	if d.exp > uint16(dec256.MaxPrecision) {
		return dec256.Zero, errors.PrecisionOutOfRange.Value()
	}
	u512, e := d.coef.Uint512()
	if e != errors.None {
		return dec256.Zero, e.Value()
	}
	coef, e := u512.Uint256()
	if e != errors.None {
		return dec256.Zero, e.Value()
	}
	return dec256.New(coef, uint8(d.exp), d.neg), nil
}

// FromDec512 widens a Dec512 without changing its numeric value or stored scale.
func FromDec512(source dec512.Dec512) Dec1024 {
	if source.IsNaN() {
		return NaN(errors.NotANumber)
	}
	return New(uint1024.FromUint512(source.Coefficient()), uint16(source.Precision()), source.IsNegative())
}

// Dec512Exact narrows only if the value is exactly representable at Dec512 width and scale.
func (d Dec1024) Dec512Exact() (dec512.Dec512, error) {
	if d.IsNaN() {
		return dec512.Zero, d.ErrorDetails()
	}
	d = d.Canonical()
	if d.exp > uint16(dec512.MaxPrecision) {
		return dec512.Zero, errors.PrecisionOutOfRange.Value()
	}
	coef, e := d.coef.Uint512()
	if e != errors.None {
		return dec512.Zero, e.Value()
	}
	return dec512.New(coef, uint8(d.exp), d.neg), nil
}
