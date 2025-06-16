// Package dec128 provides 128-bit fixed-point decimal type, operations and constants.
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

// Dec128 represents a 128-bit fixed-point decimal number.
type Dec128 struct {
	coef uint128.Uint128
	err  errors.Error
	exp  uint8
	neg  bool
}

// New creates a new Dec128 from a uint64 coefficient, uint8 exponent, and negative flag.
// In case of errors it returns NaN with the error.
func New(coef uint128.Uint128, exp uint8, neg bool) Dec128 {
	if exp > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange)
	}

	if coef.IsZero() && exp == 0 {
		return Zero
	}

	return Dec128{coef: coef, exp: exp, neg: neg}
}

// NaN returns a Dec128 with the given error.
func NaN(reason errors.Error) Dec128 {
	if reason == errors.None {
		return Dec128{err: errors.NotANumber}
	}
	return Dec128{err: reason}
}

// IsZero returns true if the Dec128 is zero.
// If the Dec128 is NaN, it returns false.
func (decimal Dec128) IsZero() bool {
	return decimal.err == errors.None && decimal.coef.IsZero()
}

// IsNegative returns true if the Dec128 is negative and false otherwise.
// If the Dec128 is NaN, it returns false.
func (decimal Dec128) IsNegative() bool {
	return decimal.neg && decimal.err == errors.None && !decimal.coef.IsZero()
}

// IsPositive returns true if the Dec128 is positive and false otherwise.
// If the Dec128 is NaN, it returns false.
func (decimal Dec128) IsPositive() bool {
	return !decimal.neg && decimal.err == errors.None && !decimal.coef.IsZero()
}

// IsNaN returns true if the Dec128 is NaN.
func (decimal Dec128) IsNaN() bool {
	return decimal.err != errors.None
}

// ErrorDetails returns the error details of the Dec128.
// If the Dec128 is not NaN, it returns nil.
func (decimal Dec128) ErrorDetails() error {
	return decimal.err.Value()
}

// Sign returns -1 if the Dec128 is negative, 0 if it is zero, and 1 if it is positive.
func (decimal Dec128) Sign() int {
	if decimal.err != errors.None || decimal.IsZero() {
		return 0
	}

	if decimal.IsNegative() {
		return -1
	}

	return 1
}

// Precision returns the precision of the Dec128.
func (decimal Dec128) Precision() uint8 {
	return decimal.exp
}

// Rescale returns a new Dec128 with the given precision.
// If the Dec128 is NaN, it returns itdecimal.
// In case of errors it returns NaN with the error.
func (decimal Dec128) Rescale(prec uint8) Dec128 {
	if decimal.err != errors.None {
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
		return Dec128{coef: coef, exp: prec, neg: decimal.neg}
	}

	// scale down
	diff := decimal.exp - prec
	coef, err := decimal.coef.Div64(Pow10Uint64[diff])
	if err != errors.None {
		return NaN(err)
	}

	return Dec128{coef: coef, exp: prec, neg: decimal.neg}
}

// Equal returns true if the Dec128 is equal to the other Dec128.
func (decimal Dec128) Equal(other Dec128) bool {
	if decimal.err != errors.None && other.err != errors.None {
		return true
	}

	if decimal.err != errors.None || other.err != errors.None {
		return false
	}

	if decimal.neg != other.neg {
		return false
	}

	if decimal.exp == other.exp {
		return decimal.coef.Equal(other.coef)
	}

	prec := max(decimal.exp, other.exp)
	a := decimal.Rescale(prec)
	b := other.Rescale(prec)
	if !a.IsNaN() && !b.IsNaN() {
		return a.coef.Equal(b.coef)
	}

	return false
}

// Compare returns -1 if the Dec128 is less than the other Dec128, 0 if they are equal, and 1 if the Dec128 is greater than the other Dec128.
// NaN is considered less than any valid Dec128.
func (decimal Dec128) Compare(other Dec128) int {
	if decimal.err != errors.None && other.err != errors.None {
		return 0
	}

	if decimal.err != errors.None {
		return -1
	}

	if other.err != errors.None {
		return 1
	}

	if decimal.neg && !other.neg {
		return -1
	}

	if !decimal.neg && other.neg {
		return 1
	}

	if decimal.exp == other.exp {
		if decimal.neg {
			return -decimal.coef.Compare(other.coef)
		}
		return decimal.coef.Compare(other.coef)
	}

	prec := max(decimal.exp, other.exp)
	a := decimal.Rescale(prec)
	if a.IsNaN() {
		return 1
	}
	b := other.Rescale(prec)
	if b.IsNaN() {
		return -1
	}

	if a.neg {
		return -a.coef.Compare(b.coef)
	}

	return a.coef.Compare(b.coef)
}

// Canonical returns a new Dec128 with the canonical representation.
// If the Dec128 is NaN, it returns itdecimal.
func (decimal Dec128) Canonical() Dec128 {
	if decimal.err != errors.None {
		return Dec128{err: decimal.err}
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

	return Dec128{coef: coef, exp: exp, neg: decimal.neg}
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
	if decimal.err != errors.None {
		return decimal
	}
	return Dec128{coef: decimal.coef, exp: decimal.exp, neg: decimal.neg}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (decimal Dec128) MarshalText() ([]byte, error) {
	if decimal.err != errors.None {
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
	if decimal.err != errors.None {
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
	if decimal.neg {
		return 0, errors.Negative.Value()
	}

	d := decimal.Rescale(exp)

	if d.err != errors.None {
		return 0, d.err.Value()
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
	if decimal.neg {
		return uint128.Zero, errors.Negative.Value()
	}

	d := decimal.Rescale(exp)

	if d.err != errors.None {
		return uint128.Zero, d.err.Value()
	}

	return d.coef, nil
}

// String returns the string representation of the Dec128 with the trailing zeros removed.
// If the Dec128 is zero, the string "0" is returned.
// If the Dec128 is NaN, the string "NaN" is returned.
func (decimal Dec128) String() string {
	if decimal.err != errors.None {
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
	if decimal.err != errors.None {
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
	if t.err != errors.None {
		return 0, t.err.Value()
	}
	if t.coef.Hi != 0 {
		return 0, errors.Overflow.Value()
	}
	if t.coef.Lo > math.MaxInt {
		return 0, errors.Overflow.Value()
	}

	if t.neg {
		return -int(t.coef.Lo), nil
	}

	return int(t.coef.Lo), nil
}

// Int64 returns the integer part of the Dec128 as int64.
func (decimal Dec128) Int64() (int64, error) {
	t := decimal.Rescale(0)
	if t.err != errors.None {
		return 0, t.err.Value()
	}
	if t.coef.Hi != 0 {
		return 0, errors.Overflow.Value()
	}
	if t.coef.Lo > math.MaxInt64 {
		return 0, errors.Overflow.Value()
	}

	if t.neg {
		return -int64(t.coef.Lo), nil
	}

	return int64(t.coef.Lo), nil
}

// InexactFloat64 returns the float64 representation of the decimal.
// The result may not be 100% accurate due to the limitation of float64 (less decimal precision).
func (decimal Dec128) InexactFloat64() (float64, error) {
	if decimal.err != errors.None {
		return 0, decimal.err.Value()
	}
	return strconv.ParseFloat(decimal.String(), 64)
}
