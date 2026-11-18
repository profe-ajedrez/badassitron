// Package dec128 preserves Badassitron decimal semantics on an imported
// Jokruger arithmetic backend. Publication requires separate provenance approval.
package dec128

import (
	"database/sql/driver"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"sync/atomic"

	jk "github.com/jokruger/dec128"
	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

const MaxPrecision uint8 = 19

type RoundingMode uint8

const (
	RoundingTowardZero RoundingMode = iota
	RoundingAwayFromZero
	RoundingFloor
	RoundingCeil
	RoundingHalfTowardZero
	RoundingHalfAwayFromZero
	RoundingHalfEven
)

// Dec128 keeps historical metadata independently from upstream's normalized
// finite-zero representation. All coefficient arithmetic uses Jokruger.
type Dec128 struct {
	coef  ju.Uint128
	scale uint8
	state uint16
}

var (
	Zero             Dec128
	One              = FromInt(1)
	NegativeOne      = FromInt(-1)
	defaultPrecision atomic.Uint32
)

func init() { defaultPrecision.Store(uint32(MaxPrecision)) }

func SetDefaultPrecision(precision uint8) {
	if precision > MaxPrecision {
		panic(le.PrecisionOutOfRange.Value())
	}
	defaultPrecision.Store(uint32(precision))
}

// New converts only the public coefficient representation at the API boundary.
// Internal decimal coefficient arithmetic stays on the imported uint128 type.
func New(c lu.Uint128, scale uint8, negative bool) Dec128 {
	return newCoefficient(ju.Uint128(c), scale, negative)
}

func newCoefficient(c ju.Uint128, scale uint8, negative bool) Dec128 {
	if scale > MaxPrecision {
		return NaN(le.PrecisionOutOfRange)
	}
	if c.IsZero() && scale == 0 {
		return Zero
	}
	return Dec128{coef: c, scale: scale, state: signState(negative)}
}

func NewStoredZero(scale uint8, negative bool) Dec128 {
	if scale > MaxPrecision {
		return NaN(le.PrecisionOutOfRange)
	}
	return Dec128{scale: scale, state: signState(negative)}
}

func fromParts(c ju.Uint128, scale uint8, negative bool) Dec128 {
	return newCoefficient(c, scale, negative)
}
func NaN(reason le.Error) Dec128 {
	if reason == le.None {
		reason = le.NotANumber
	}
	return Dec128{state: uint16(reason) << 1}
}
func signState(negative bool) uint16 {
	if negative {
		return 1
	}
	return 0
}
func (d Dec128) negative() bool       { return d.state&1 != 0 }
func (d Dec128) reason() le.Error     { return le.Error(d.state >> 1) }
func (d Dec128) ErrorCode() le.Error  { return d.reason() }
func (d Dec128) ErrorDetails() error  { return d.reason().Value() }
func (d Dec128) IsNaN() bool          { return d.reason() != le.None }
func (d Dec128) IsZero() bool         { return !d.IsNaN() && d.coef.IsZero() }
func (d Dec128) IsNegative() bool     { return !d.IsNaN() && !d.coef.IsZero() && d.negative() }
func (d Dec128) IsPositive() bool     { return !d.IsNaN() && !d.coef.IsZero() && !d.negative() }
func (d Dec128) StoredNegative() bool { return !d.IsNaN() && d.negative() }
func (d Dec128) Sign() int {
	if d.IsNaN() || d.IsZero() {
		return 0
	}
	if d.negative() {
		return -1
	}
	return 1
}
func (d Dec128) Precision() uint8        { return d.scale }
func (d Dec128) Exponent() uint8         { return d.scale }
func (d Dec128) Coefficient() lu.Uint128 { return lu.Uint128{Lo: d.coef.Lo, Hi: d.coef.Hi} }
func (d Dec128) Copy() Dec128            { return d }
func (d Dec128) upstream() jk.Dec128 {
	if d.IsNaN() {
		return jk.NaN(toState(d.reason().Value()))
	}
	return jk.New(d.coef, d.scale, d.negative())
}
func toState(err error) state.State {
	switch err {
	case le.DivisionByZero.Value():
		return state.DivisionByZero
	case le.Overflow.Value():
		return state.Overflow
	case le.Underflow.Value():
		return state.Underflow
	case le.Negative.Value():
		return state.NegativeInUnsignedOp
	case le.NotEnoughBytes.Value():
		return state.NotEnoughBytes
	case le.InvalidFormat.Value():
		return state.InvalidFormat
	case le.PrecisionOutOfRange.Value():
		return state.PrecisionOutOfRange
	case le.RescaleToLessPrecision.Value():
		return state.RescaleToLessPrecision
	case le.SqrtNegative.Value():
		return state.SqrtNegative
	default:
		return state.NaN
	}
}
func fromState(err error) le.Error {
	switch err {
	case state.DivisionByZero.Error():
		return le.DivisionByZero
	case state.Overflow.Error():
		return le.Overflow
	case state.Underflow.Error():
		return le.Underflow
	case state.NegativeInUnsignedOp.Error():
		return le.Negative
	case state.NotEnoughBytes.Error():
		return le.NotEnoughBytes
	case state.InvalidFormat.Error():
		return le.InvalidFormat
	case state.PrecisionOutOfRange.Error(), state.ScaleOutOfRange.Error():
		return le.PrecisionOutOfRange
	case state.RescaleToLessPrecision.Error(), state.RescaleToLowerScale.Error():
		return le.RescaleToLessPrecision
	case state.SqrtNegative.Error():
		return le.SqrtNegative
	default:
		return le.NotANumber
	}
}
func (d Dec128) Canonical() Dec128 {
	if d.IsNaN() {
		return NaN(d.reason())
	}
	if d.IsZero() {
		return Zero
	}
	for d.scale > 0 {
		q, r, st := d.coef.QuoRem64(10)
		if st != state.OK || r != 0 {
			break
		}
		d.coef, d.scale = q, d.scale-1
	}
	return d
}
func (d Dec128) Rescale(scale uint8) Dec128 {
	if d.IsNaN() {
		return d
	}
	if scale == d.scale {
		return d
	}
	if scale > MaxPrecision {
		return NaN(le.PrecisionOutOfRange)
	}
	if scale > d.scale {
		coef, st := d.coef.Mul(ju.Pow10Uint128[scale-d.scale])
		if st != state.OK {
			return NaN(le.Overflow)
		}
		return Dec128{coef: coef, scale: scale, state: signState(d.negative())}
	}
	coef, _, st := d.coef.QuoRem64(ju.Pow10Uint64[d.scale-scale])
	if st != state.OK {
		return NaN(fromState(st.Error()))
	}
	// Raw sign is deliberately preserved even when truncation yields zero.
	return Dec128{coef: coef, scale: scale, state: signState(d.negative())}
}
func (d Dec128) Compare(b Dec128) int {
	if d.IsNaN() {
		if b.IsNaN() {
			return 0
		}
		return -1
	}
	if b.IsNaN() {
		return 1
	}
	if d.IsZero() || b.IsZero() {
		if d.IsZero() && b.IsZero() {
			return 0
		}
		if d.IsZero() {
			if b.negative() {
				return 1
			}
			return -1
		}
		if d.negative() {
			return -1
		}
		return 1
	}
	if d.negative() != b.negative() {
		if d.negative() {
			return -1
		}
		return 1
	}
	var cmp int
	if d.scale == b.scale {
		cmp = d.coef.Compare(b.coef)
	} else if d.scale < b.scale {
		lo, hi := d.coef.MulCarry(ju.Pow10Uint128[b.scale-d.scale])
		if !hi.IsZero() {
			cmp = 1
		} else {
			cmp = lo.Compare(b.coef)
		}
	} else {
		lo, hi := b.coef.MulCarry(ju.Pow10Uint128[d.scale-b.scale])
		if !hi.IsZero() {
			cmp = -1
		} else {
			cmp = d.coef.Compare(lo)
		}
	}
	if d.negative() {
		return -cmp
	}
	return cmp
}
func (d Dec128) Equal(b Dec128) bool              { return d.Compare(b) == 0 }
func (d Dec128) LessThan(b Dec128) bool           { return d.Compare(b) < 0 }
func (d Dec128) LessThanOrEqual(b Dec128) bool    { return d.Compare(b) <= 0 }
func (d Dec128) GreaterThan(b Dec128) bool        { return d.Compare(b) > 0 }
func (d Dec128) GreaterThanOrEqual(b Dec128) bool { return d.Compare(b) >= 0 }
func (d Dec128) Neg() Dec128 {
	if !d.IsNaN() {
		d.state ^= 1
	}
	return d
}
func (d Dec128) Abs() Dec128 {
	if !d.IsNaN() {
		d.state &^= 1
	}
	return d
}

// StringFixed renders the stored scale using Jokruger's imported uint128
// decimal digit writer. Avoiding a temporary upstream Dec128 also avoids
// reading its formatting globals. Only the final string is allocated.
func (d Dec128) StringFixed() string {
	if d.IsNaN() {
		return "NaN"
	}
	if d.IsZero() {
		if d.scale == 0 {
			return "0"
		}
		return "0." + strings.Repeat("0", int(d.scale))
	}
	var digitsBuf [ju.MaxStrLen]byte
	digits := d.coef.StringToBuf(digitsBuf[:])
	var buf [ju.MaxStrLen + 3]byte
	i := 0
	if d.negative() {
		buf[i] = '-'
		i++
	}
	if d.scale == 0 {
		i += copy(buf[i:], digits)
	} else if len(digits) > int(d.scale) {
		whole := len(digits) - int(d.scale)
		i += copy(buf[i:], digits[:whole])
		buf[i] = '.'
		i++
		i += copy(buf[i:], digits[whole:])
	} else {
		buf[i] = '0'
		i++
		buf[i] = '.'
		i++
		for n := int(d.scale) - len(digits); n > 0; n-- {
			buf[i] = '0'
			i++
		}
		i += copy(buf[i:], digits)
	}
	return string(buf[:i])
}
func (d Dec128) String() string {
	if d.IsNaN() {
		return "NaN"
	}
	if d.IsZero() {
		return "0"
	}
	return d.upstream().String()
}
func (d Dec128) MarshalText() ([]byte, error) { return []byte(d.String()), nil }
func (d Dec128) MarshalJSON() ([]byte, error) { return json.Marshal(d.String()) }
func (d Dec128) Value() (driver.Value, error) { return d.String(), nil }
func (d Dec128) Int64() (int64, error) {
	v := d.Rescale(0)
	if v.IsNaN() {
		return 0, v.ErrorDetails()
	}
	if v.coef.Hi != 0 || v.coef.Lo > math.MaxInt64 {
		return 0, le.Overflow.Value()
	}
	if v.negative() {
		return -int64(v.coef.Lo), nil
	}
	return int64(v.coef.Lo), nil
}
func (d Dec128) Int() (int, error) {
	v := d.Rescale(0)
	if v.IsNaN() {
		return 0, v.ErrorDetails()
	}
	if v.coef.Hi != 0 || v.coef.Lo > math.MaxInt {
		return 0, le.Overflow.Value()
	}
	if v.negative() {
		return -int(v.coef.Lo), nil
	}
	return int(v.coef.Lo), nil
}
func (d Dec128) InexactFloat64() (float64, error) { return strconv.ParseFloat(d.String(), 64) }
func FromInt(i int) Dec128                        { return FromInt64(int64(i)) }
func FromInt64(i int64) Dec128 {
	if i == 0 {
		return Zero
	}
	if i < 0 {
		u := uint64(-(i + 1)) + 1
		return fromParts(ju.Uint128{Lo: u}, 0, true)
	}
	return fromParts(ju.Uint128{Lo: uint64(i)}, 0, false)
}
func FromFloat64(f float64) Dec128 {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return NaN(le.NotANumber)
	}
	return FromString(strconv.FormatFloat(f, 'f', -1, 64))
}

func fromUpstream(v jk.Dec128) Dec128 {
	if v.IsNaN() {
		return NaN(fromState(v.ErrorDetails()))
	}
	c := v.Coefficient()
	return fromParts(c, v.Scale(), v.IsNegative())
}
