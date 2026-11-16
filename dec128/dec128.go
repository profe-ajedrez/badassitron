// Package dec128 provides fixed-width base-10 decimal arithmetic with a
// 128-bit unsigned coefficient, a sign, and a scale up to 19 fractional places.
// Precision returns the stored scale, not significant-digit precision.
//
// Construct exact values from decimal text or integers. Ordinary Add, Sub,
// and Mul preserve exact results where representable, otherwise return NaN.
// Division and square root use truncation; use explicit DivRound, MulRound,
// SqrtRound, or RescaleRound when a particular rounding policy is required.
// Check IsNaN after arithmetic; WithStatus methods report discarded nonzero
// digits, not an error.
//
// For database/sql, use ScanDestination(&d) instead of &d: the historical
// Dec128.Scan method has a value receiver and does not update the caller.
// See SQL_SCANNING.md and the repository's examples/06_dec128 directory.
package dec128

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// Dec128 is a fixed-width signed decimal value with a 128-bit coefficient.
// It is a value type. Its private representation can change between versions;
// use the public constructors and methods rather than unsafe field access.
type Dec128 struct {
	coef  uint128.Uint128
	exp   uint8
	state uint16 // bit 0: negative sign; bits 1-8: full errors.Error code
}

// New constructs a decimal from a uint128 coefficient, fractional scale,
// and negative-sign flag. A scale above MaxPrecision returns NaN.
func New(coef uint128.Uint128, exp uint8, neg bool) Dec128 {
	if exp > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange)
	}

	if coef.IsZero() && exp == 0 {
		return Zero
	}

	return Dec128{coef: coef, exp: exp, state: signState(neg)}
}

// NaN returns a Dec128 with the given error.
func NaN(reason errors.Error) Dec128 {
	if reason == errors.None {
		return Dec128{state: errorState(errors.NotANumber)}
	}
	return Dec128{state: errorState(reason)}
}

// IsZero returns true if the Dec128 is zero.
// If the Dec128 is NaN, it returns false.
func (decimal Dec128) IsZero() bool {
	return decimal.errorCode() == errors.None && decimal.coef.IsZero()
}

// IsNegative returns true if the Dec128 is negative and false otherwise.
// If the Dec128 is NaN, it returns false.
func (decimal Dec128) IsNegative() bool {
	return decimal.negative() && decimal.errorCode() == errors.None && !decimal.coef.IsZero()
}

// IsPositive returns true if the Dec128 is positive and false otherwise.
// If the Dec128 is NaN, it returns false.
func (decimal Dec128) IsPositive() bool {
	return !decimal.negative() && decimal.errorCode() == errors.None && !decimal.coef.IsZero()
}

// IsNaN returns true if the Dec128 is NaN.
func (decimal Dec128) IsNaN() bool {
	return decimal.errorCode() != errors.None
}

// ErrorDetails returns the error details of the Dec128.
// If the Dec128 is not NaN, it returns nil.
func (decimal Dec128) ErrorDetails() error {
	return decimal.errorCode().Value()
}

// Sign returns -1 if the Dec128 is negative, 0 if it is zero, and 1 if it is positive.
func (decimal Dec128) Sign() int {
	if decimal.errorCode() != errors.None || decimal.IsZero() {
		return 0
	}

	if decimal.IsNegative() {
		return -1
	}

	return 1
}

// Precision returns the stored fractional decimal scale (0 through 19),
// not the number of significant digits.
func (decimal Dec128) Precision() uint8 {
	return decimal.exp
}

// Rescale changes the stored fractional scale without rounding upward.
// It truncates digits toward zero when reducing the scale, and may return
// NaN if increasing the scale overflows or the requested scale is invalid.
// An existing NaN is propagated. Use RescaleRound for explicit rounding.
func (decimal Dec128) Rescale(prec uint8) Dec128 {
	if decimal.errorCode() != errors.None {
		return decimal
	}

	if decimal.exp == prec {
		return decimal
	}

	if prec > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange)
	}

	if prec > decimal.exp {
		// scale up
		diff := prec - decimal.exp
		coef, err := decimal.coef.Mul64(Pow10Uint64[diff])
		if err != errors.None {
			return NaN(err)
		}
		return Dec128{coef: coef, exp: prec, state: signState(decimal.negative())}
	}

	// scale down
	diff := decimal.exp - prec
	coef, err := decimal.coef.Div64(Pow10Uint64[diff])
	if err != errors.None {
		return NaN(err)
	}

	return Dec128{coef: coef, exp: prec, state: signState(decimal.negative())}
}

// Equal reports numerical equality, regardless of scale or the sign of zero.
// As before, two NaN values compare equal.
func (decimal Dec128) Equal(other Dec128) bool {
	if decimal.errorCode() != errors.None || other.errorCode() != errors.None {
		return decimal.errorCode() != errors.None && other.errorCode() != errors.None
	}
	return decimal.Compare(other) == 0
}

// compareMagnitude compares unsigned decimal magnitudes without truncating or
// overflowing the 128-bit coefficients. Scale alignment can require up to
// 192 bits, so use MulCarry's complete 256-bit product instead of Rescale.
func compareMagnitude(a, b Dec128) int {
	if a.exp == b.exp {
		return a.coef.Compare(b.coef)
	}
	if a.exp < b.exp {
		lo, hi := a.coef.MulCarry(Pow10Uint128[b.exp-a.exp])
		if !hi.IsZero() {
			return 1
		}
		return lo.Compare(b.coef)
	}
	lo, hi := b.coef.MulCarry(Pow10Uint128[a.exp-b.exp])
	if !hi.IsZero() {
		return -1
	}
	return a.coef.Compare(lo)
}

// Compare returns -1 if decimal is less than other, 0 if they have the
// same numerical value, and 1 if decimal is greater than other.
// As before, NaN sorts before any valid Dec128 and two NaNs compare equal.
func (decimal Dec128) Compare(other Dec128) int {
	// Both operands fit the regular-state ABI path and share scale/sign.
	// Comparing magnitudes directly also handles two same-sign zeros.
	if decimal.exp == other.exp && decimal.state == other.state && decimal.state < 2 {
		cmp := decimal.coef.Compare(other.coef)
		if decimal.state == 1 {
			return -cmp
		}
		return cmp
	}

	if decimal.errorCode() != errors.None && other.errorCode() != errors.None {
		return 0
	}
	if decimal.errorCode() != errors.None {
		return -1
	}
	if other.errorCode() != errors.None {
		return 1
	}

	// Zero has no mathematical sign. Constructors can retain negative zero
	// at non-zero scales, but it must still compare equal to ordinary zero.
	if decimal.coef.IsZero() {
		if other.coef.IsZero() {
			return 0
		}
		if other.negative() {
			return 1
		}
		return -1
	}
	if other.coef.IsZero() {
		if decimal.negative() {
			return -1
		}
		return 1
	}

	if decimal.negative() != other.negative() {
		if decimal.negative() {
			return -1
		}
		return 1
	}
	// The general sign and signed-zero cases are handled above. Bypass
	// compareMagnitude for a matching scale without adding an extra
	// condition to the earlier fast exits.
	if decimal.exp == other.exp {
		cmp := decimal.coef.Compare(other.coef)
		if decimal.negative() {
			return -cmp
		}
		return cmp
	}
	cmp := compareMagnitude(decimal, other)
	if decimal.negative() {
		return -cmp
	}
	return cmp
}

// Canonical removes exact trailing fractional zeros while preserving the
// numeric value. It normalizes zero and propagates invalid values.
func (decimal Dec128) Canonical() Dec128 {
	if decimal.errorCode() != errors.None {
		return Dec128{state: errorState(decimal.errorCode())}
	}

	if decimal.IsZero() {
		return Zero
	}

	if decimal.exp == 0 {
		return decimal
	}

	coef := decimal.coef
	exp := decimal.exp
	for {
		t, r, err := coef.QuoRem64(10)
		if err != errors.None || r != 0 {
			break
		}
		coef = t
		exp--
		if exp == 0 {
			break
		}
	}

	return Dec128{coef: coef, exp: exp, state: signState(decimal.negative())}
}

// Exponent returns the exponent of the Dec128.
func (decimal Dec128) Exponent() uint8 {
	return decimal.exp
}

// Coefficient returns the coefficient of the Dec128.
func (decimal Dec128) Coefficient() uint128.Uint128 {
	return decimal.coef
}

// LessThan returns true if the Dec128 is less than the other Dec128.
func (decimal Dec128) LessThan(other Dec128) bool {
	return decimal.Compare(other) < 0
}

// LessThanOrEqual returns true if the Dec128 is less than or equal to the other Dec128.
func (decimal Dec128) LessThanOrEqual(other Dec128) bool {
	return decimal.Compare(other) <= 0
}

// GreaterThan returns true if the Dec128 is greater than the other Dec128.
func (decimal Dec128) GreaterThan(other Dec128) bool {
	return decimal.Compare(other) > 0
}

// GreaterThanOrEqual returns true if the Dec128 is greater than or equal to the other Dec128.
func (decimal Dec128) GreaterThanOrEqual(other Dec128) bool {
	return decimal.Compare(other) >= 0
}

// Copy returns a copy of the Dec128.
func (decimal Dec128) Copy() Dec128 {
	if decimal.errorCode() != errors.None {
		return decimal
	}
	return Dec128{coef: decimal.coef, exp: decimal.exp, state: signState(decimal.negative())}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (decimal Dec128) MarshalText() ([]byte, error) {
	if decimal.errorCode() != errors.None {
		return NaNStrBytes, nil
	}

	if decimal.IsZero() {
		return ZeroStrBytes, nil
	}

	buf := [MaxStrLen]byte{}
	sb, trim := decimal.appendString(buf[:0])
	if trim {
		return trimTrailingZeros(sb), nil
	}

	return sb, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (decimal *Dec128) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*decimal = Zero
		return nil
	}

	t := FromString(data[:])
	if t.IsNaN() {
		return t.ErrorDetails()
	}
	*decimal = t

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (decimal Dec128) MarshalJSON() ([]byte, error) {
	if decimal.errorCode() != errors.None {
		return NaNJsonStrBytes, nil
	}

	if decimal.IsZero() {
		return ZeroJsonStrBytes, nil
	}

	buf := [MaxStrLen + 2]byte{}
	buf[0] = '"'
	sb, trim := decimal.appendString(buf[:1])
	if trim {
		sb = trimTrailingZeros(sb)
	}
	return append(sb, '"'), nil
}

var nullValue = []byte("null")

// UnmarshalJSON implements the json.Unmarshaler interface.
func (decimal *Dec128) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}

	if len(data) == 0 || bytes.Equal(data, nullValue) {
		*decimal = Zero
		return nil
	}

	t := FromString(data[:])
	if t.IsNaN() {
		return t.ErrorDetails()
	}
	*decimal = t

	return nil
}

// Scan implements the sql.Scanner interface.
//
// Deprecated: this historical value receiver cannot populate the caller's
// Dec128. For database/sql use ScanDestination(&value), or use ScanInto
// for direct conversion. Its signature is retained for API compatibility.
func (decimal Dec128) Scan(src any) error {
	var err error
	switch v := src.(type) {
	case string:
		decimal = FromString(v)
		if decimal.IsNaN() {
			err = decimal.ErrorDetails()
		}
	case int:
		decimal = FromInt(v)
	case int64:
		decimal = FromInt64(v)
	case float32:
		decimal = FromString(fmt.Sprintf("%f", v))
	case float64:
		decimal = FromFloat64(v)
	case nil:
		decimal = Zero
	default:
		err = fmt.Errorf("can't scan %T to Dec128: %T is not supported", src, src)
	}

	return err
}

// Value implements the driver.Valuer interface.
func (decimal Dec128) Value() (driver.Value, error) {
	return decimal.String(), nil
}

// EncodeToUint64 returns the Dec128 encoded as uint64 coefficient with requested exponent.
// Negative and too large values are not allowed.
func (decimal Dec128) EncodeToUint64(exp uint8) (uint64, error) {
	if decimal.negative() {
		return 0, errors.Negative.Value()
	}

	d := decimal.Rescale(exp)

	if d.errorCode() != errors.None {
		return 0, d.errorCode().Value()
	}

	i, err := d.coef.Uint64()
	if err != errors.None {
		return 0, err.Value()
	}

	return i, nil
}

// EncodeToUint128 returns the Dec128 encoded as uint128 coefficient with requested exponent.
// Negative values are not allowed.
func (decimal Dec128) EncodeToUint128(exp uint8) (uint128.Uint128, error) {
	if decimal.negative() {
		return uint128.Zero, errors.Negative.Value()
	}

	d := decimal.Rescale(exp)

	if d.errorCode() != errors.None {
		return uint128.Zero, d.errorCode().Value()
	}

	return d.coef, nil
}

// String returns the string representation of the Dec128 with the trailing zeros removed.
// If the Dec128 is zero, the string "0" is returned.
// If the Dec128 is NaN, the string "NaN" is returned.
func (decimal Dec128) String() string {
	if decimal.errorCode() != errors.None {
		return NaNStr
	}

	if decimal.IsZero() {
		return ZeroStr
	}

	buf := [MaxStrLen]byte{}
	sb, trim := decimal.appendString(buf[:0])
	if trim {
		return string(trimTrailingZeros(sb))
	}

	return string(sb)
}

// StringFixed returns the string representation of the Dec128 with the trailing zeros preserved.
// If the Dec128 is NaN, the string "NaN" is returned.
func (decimal Dec128) StringFixed() string {
	if decimal.errorCode() != errors.None {
		return NaNStr
	}

	if decimal.IsZero() {
		return zeroStrs[decimal.exp]
	}

	buf := [MaxStrLen]byte{}
	sb, _ := decimal.appendString(buf[:0])

	return string(sb)
}

// Int returns the integer part of the Dec128 as int.
func (decimal Dec128) Int() (int, error) {
	t := decimal.Rescale(0)
	if t.errorCode() != errors.None {
		return 0, t.errorCode().Value()
	}
	if t.coef.Hi != 0 {
		return 0, errors.Overflow.Value()
	}
	if t.coef.Lo > math.MaxInt {
		return 0, errors.Overflow.Value()
	}

	if t.negative() {
		return -int(t.coef.Lo), nil
	}

	return int(t.coef.Lo), nil
}

// Int64 returns the integer part of the Dec128 as int64.
func (decimal Dec128) Int64() (int64, error) {
	t := decimal.Rescale(0)
	if t.errorCode() != errors.None {
		return 0, t.errorCode().Value()
	}
	if t.coef.Hi != 0 {
		return 0, errors.Overflow.Value()
	}
	if t.coef.Lo > math.MaxInt64 {
		return 0, errors.Overflow.Value()
	}

	if t.negative() {
		return -int64(t.coef.Lo), nil
	}

	return int64(t.coef.Lo), nil
}

// InexactFloat64 returns the float64 representation of the decimal.
// The result may not be 100% accurate due to the limitation of float64 (less decimal precision).
func (decimal Dec128) InexactFloat64() (float64, error) {
	if decimal.errorCode() != errors.None {
		return 0, decimal.errorCode().Value()
	}
	return strconv.ParseFloat(decimal.String(), 64)
}
