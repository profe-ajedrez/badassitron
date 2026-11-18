// Package uint128 retains the Badassitron public representation and errors,
// while Jokruger v1.4.0 supplies the arithmetic implementation.
package uint128

import (
	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
)

// These conversions deliberately use public limbs. A Go type alias would
// change method sets and error-return types in the public API.
func toUp(u Uint128) ju.Uint128   { return ju.Uint128{Lo: u.Lo, Hi: u.Hi} }
func fromUp(u ju.Uint128) Uint128 { return Uint128{Lo: u.Lo, Hi: u.Hi} }

// mapping is scoped to unsigned-integer operations. Decimal operations have
// different context-dependent translations and must not use this mapping.
func fromState(s state.State) errors.Error {
	switch s {
	case state.OK, state.Neg:
		return errors.None
	case state.DivisionByZero:
		return errors.DivisionByZero
	case state.Overflow:
		return errors.Overflow
	case state.Underflow:
		return errors.Underflow
	case state.NegativeInUnsignedOp:
		return errors.Negative
	case state.NotEnoughBytes:
		return errors.NotEnoughBytes
	case state.InvalidFormat:
		return errors.InvalidFormat
	case state.PrecisionOutOfRange, state.ScaleOutOfRange:
		return errors.PrecisionOutOfRange
	case state.RescaleToLessPrecision, state.RescaleToLowerScale:
		return errors.RescaleToLessPrecision
	case state.SqrtNegative:
		return errors.SqrtNegative
	default:
		return errors.NotANumber
	}
}

// Jokruger v1.4.0 unsigned-state codes are pinned: OK=0, Overflow=5,
// Underflow=6, DivisionByZero=4. Badassitron codes are None=0,
// Overflow=3, Underflow=4, DivisionByZero=2. The small mappings below
// preserve historical errors without an extra branch on the hot path.
// Add instead uses the imported AddCarry primitive's 0/1 carry.
// The release tests assert these constants and both success/error results.
// These formulas MUST be reviewed when updating the imported dependency.
func (u Uint128) Add(other Uint128) (Uint128, errors.Error) {
	// AddCarry delegates the arithmetic to Jokruger, exposing carry without
	// the upstream state-code conversion required by Add itself.
	r, carry := ju.Uint128(u).AddCarry(ju.Uint128(other))
	// carry is 0 or 1: carry-1 is all ones on success, zero on overflow.
	// Masking both limbs preserves the historical zero result on overflow.
	mask := carry - 1
	return Uint128{Lo: r.Lo & mask, Hi: r.Hi & mask}, errors.Error(uint8(carry) * uint8(errors.Overflow))
}
func (u Uint128) Add64(other uint64) (Uint128, errors.Error) {
	r, st := ju.Uint128(u).Add64(other)
	return Uint128(r), errors.Error((uint8(st) & 1) * uint8(errors.Overflow))
}
func (u Uint128) Sub(other Uint128) (Uint128, errors.Error) {
	r, st := ju.Uint128(u).Sub(ju.Uint128(other))
	return Uint128(r), errors.Error(uint8(st) & uint8(errors.Underflow))
}
func (u Uint128) Sub64(other uint64) (Uint128, errors.Error) {
	r, st := ju.Uint128(u).Sub64(other)
	return Uint128(r), errors.Error(uint8(st) & uint8(errors.Underflow))
}
func (u Uint128) Mul(other Uint128) (Uint128, errors.Error) {
	r, s := ju.Uint128(u).Mul(ju.Uint128(other))
	return Uint128(r), errors.Error((uint8(s) & 1) * uint8(errors.Overflow))
}
func (u Uint128) Mul64(other uint64) (Uint128, errors.Error) {
	r, st := ju.Uint128(u).Mul64(other)
	return Uint128(r), errors.Error((uint8(st) & 1) * uint8(errors.Overflow))
}
func (u Uint128) MulCarry(other Uint128) (Uint128, Uint128) {
	lo, hi := toUp(u).MulCarry(toUp(other))
	return fromUp(lo), fromUp(hi)
}

// Ordinary unsigned quotient methods return only 0 (OK) or 4 (DivisionByZero).
// The shift maps these to Badassitron's 0 (None) and 2 (DivisionByZero).
func quotientError(s state.State) errors.Error { return errors.Error(uint8(s) >> 1) }

func (u Uint128) Div(other Uint128) (Uint128, errors.Error) {
	r, s := ju.Uint128(u).Div(ju.Uint128(other))
	return Uint128(r), quotientError(s)
}
func (u Uint128) Div64(other uint64) (Uint128, errors.Error) {
	r, s := ju.Uint128(u).Div64(other)
	return Uint128(r), quotientError(s)
}
func (u Uint128) Mod(other Uint128) (Uint128, errors.Error) {
	r, s := ju.Uint128(u).Mod(ju.Uint128(other))
	return Uint128(r), quotientError(s)
}
func (u Uint128) Mod64(other uint64) (uint64, errors.Error) {
	r, s := ju.Uint128(u).Mod64(other)
	return r, quotientError(s)
}
func (u Uint128) QuoRem(other Uint128) (Uint128, Uint128, errors.Error) {
	q, r, s := ju.Uint128(u).QuoRem(ju.Uint128(other))
	return Uint128(q), Uint128(r), quotientError(s)
}
func (u Uint128) QuoRem64(other uint64) (Uint128, uint64, errors.Error) {
	q, r, s := ju.Uint128(u).QuoRem64(other)
	return Uint128(q), r, quotientError(s)
}

func (u Uint128) And(other Uint128) Uint128     { return fromUp(toUp(u).And(toUp(other))) }
func (u Uint128) And64(other uint64) Uint128    { return fromUp(toUp(u).And64(other)) }
func (u Uint128) Or(other Uint128) Uint128      { return fromUp(toUp(u).Or(toUp(other))) }
func (u Uint128) Or64(other uint64) Uint128     { return fromUp(toUp(u).Or64(other)) }
func (u Uint128) Xor(other Uint128) Uint128     { return fromUp(toUp(u).Xor(toUp(other))) }
func (u Uint128) Xor64(other uint64) Uint128    { return fromUp(toUp(u).Xor64(other)) }
func (u Uint128) Lsh(n uint) Uint128            { return fromUp(toUp(u).Lsh(n)) }
func (u Uint128) Rsh(n uint) Uint128            { return fromUp(toUp(u).Rsh(n)) }
func (u Uint128) LeadingZeroBitsCount() int     { return toUp(u).LeadingZeroBitsCount() }
func (u Uint128) TrailingZeroBitsCount() int    { return toUp(u).TrailingZeroBitsCount() }
func (u Uint128) NonZeroBitsCount() int         { return toUp(u).NonZeroBitsCount() }
func (u Uint128) RotateBitsLeft(k int) Uint128  { return fromUp(toUp(u).RotateBitsLeft(k)) }
func (u Uint128) RotateBitsRight(k int) Uint128 { return fromUp(toUp(u).RotateBitsRight(k)) }
func (u Uint128) ReverseBits() Uint128          { return fromUp(toUp(u).ReverseBits()) }

// Wide division returns only OK=0, DivisionByZero=4, or Overflow=5.
// (s+1)>>1 maps these to the historical 0, 2, and 3 without a switch.
// Upstream itself checks carry >= divisor and preserves zero-divisor precedence.
func QuoRem256By128(u Uint128, carry Uint128, v Uint128) (Uint128, Uint128, errors.Error) {
	q, r, s := ju.QuoRem256By128(ju.Uint128(u), ju.Uint128(carry), ju.Uint128(v))
	return Uint128(q), Uint128(r), errors.Error((uint8(s) + 1) >> 1)
}
func QuoRem192By64(u Uint128, carry uint64, v uint64) (Uint128, uint64, errors.Error) {
	q, r, s := ju.QuoRem192By64(ju.Uint128(u), carry, v)
	return Uint128(q), r, errors.Error((uint8(s) + 1) >> 1)
}
func SubUnsafe(u Uint128, v Uint128) Uint128 { return fromUp(ju.SubUnsafe(toUp(u), toUp(v))) }
