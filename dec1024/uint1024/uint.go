// Package uint1024 provides fixed-width 1024-bit unsigned arithmetic.
// Operations do not use floating-point arithmetic or math/big internally.
package uint1024

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/bits"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec256/uint256"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
	"github.com/profe-ajedrez/badassitron/internal/uintwide"
)

// Uint1024 stores an unsigned 1024-bit integer as little-endian halves.
// Like uint128.Uint128, it is a comparable value type.
type Uint1024 struct {
	Lo uint512.Uint512
	Hi uint512.Uint512
}

const (
	// MaxStrLen is the maximum length of the decimal representation.
	MaxStrLen = 309
	// MaxStrLen64 is the largest number of decimal digits for a uint64.
	MaxStrLen64 = 20
	// MaxSafeStrLen64 is the safe decimal digit count for parsing uint64 without overflowing.
	MaxSafeStrLen64 = 19
	// ZeroStr is the text representation of zero.
	ZeroStr = "0"
)

var (
	// Zero is the predefined zero value as Uint1024.
	Zero = Uint1024{}
	// One is the predefined one value as Uint1024.
	One = FromUint64(1)
	// Max is the predefined largest representable value as Uint1024.
	Max = Uint1024{Lo: uint512.Max, Hi: uint512.Max}
	// Max64 is the predefined largest uint64 value as Uint1024.
	Max64 = FromUint64(math.MaxUint64)
	// Pow10Uint1024 contains precisely representable powers 10^0 through 10^308.
	Pow10Uint1024 = makePowers10()
)

func makePowers10() [309]Uint1024 {
	var out [309]Uint1024
	out[0] = One
	for i := 1; i < len(out); i++ {
		next, e := out[i-1].Mul64(10)
		if e != errors.None {
			panic("internal powers-of-ten overflow")
		}
		out[i] = next
	}
	return out
}

func (u Uint1024) words() [16]uint64 {
	a, b := u.Lo, u.Hi
	return [16]uint64{
		a.Lo.Lo.Lo, a.Lo.Lo.Hi, a.Lo.Hi.Lo, a.Lo.Hi.Hi,
		a.Hi.Lo.Lo, a.Hi.Lo.Hi, a.Hi.Hi.Lo, a.Hi.Hi.Hi,
		b.Lo.Lo.Lo, b.Lo.Lo.Hi, b.Lo.Hi.Lo, b.Lo.Hi.Hi,
		b.Hi.Lo.Lo, b.Hi.Lo.Hi, b.Hi.Hi.Lo, b.Hi.Hi.Hi,
	}
}

func fromWords(w [16]uint64) Uint1024 {
	return Uint1024{
		Lo: uint512.Uint512{
			Lo: uint256.Uint256{Lo: uint128.Uint128{Lo: w[0], Hi: w[1]}, Hi: uint128.Uint128{Lo: w[2], Hi: w[3]}},
			Hi: uint256.Uint256{Lo: uint128.Uint128{Lo: w[4], Hi: w[5]}, Hi: uint128.Uint128{Lo: w[6], Hi: w[7]}},
		},
		Hi: uint512.Uint512{
			Lo: uint256.Uint256{Lo: uint128.Uint128{Lo: w[8], Hi: w[9]}, Hi: uint128.Uint128{Lo: w[10], Hi: w[11]}},
			Hi: uint256.Uint256{Lo: uint128.Uint128{Lo: w[12], Hi: w[13]}, Hi: uint128.Uint128{Lo: w[14], Hi: w[15]}},
		},
	}
}

// FromUint64 creates a wide unsigned integer from a uint64.
func FromUint64(v uint64) Uint1024 { return Uint1024{Lo: uint512.FromUint64(v)} }

// FromUint512 widens a Uint512 without changing its value.
func FromUint512(v uint512.Uint512) Uint1024 { return Uint1024{Lo: v} }

// FromBigInt rejects negative and overflowing integers; it does not mutate v.
func FromBigInt(v *big.Int) (Uint1024, errors.Error) {
	if v == nil {
		return Zero, errors.InvalidFormat
	}
	if v.Sign() < 0 {
		return Zero, errors.Negative
	}
	if v.BitLen() > 1024 {
		return Zero, errors.Overflow
	}
	var b [128]byte
	v.FillBytes(b[:])
	return FromBytesBigEndian(b), errors.None
}

// FromString parses an unsigned base-10 integer, rejecting overflow.
// An empty string maps to zero, like uint128.FromString.
func FromString[S string | []byte](s S) (Uint1024, errors.Error) {
	var v Uint1024
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return Zero, errors.InvalidFormat
		}
		x, e := v.Mul64(10)
		if e != errors.None {
			return Zero, e
		}
		v, e = x.Add64(uint64(s[i] - '0'))
		if e != errors.None {
			return Zero, e
		}
	}
	return v, errors.None
}

// FromBytes parses 128 bytes in little-endian order.
func FromBytes(b [128]byte) Uint1024 {
	var w [16]uint64
	for i := range w {
		w[i] = binary.LittleEndian.Uint64(b[i*8 : i*8+8])
	}
	return fromWords(w)
}

// FromBytesBigEndian parses 128 bytes in big-endian order.
func FromBytesBigEndian(b [128]byte) Uint1024 {
	var w [16]uint64
	for i := range w {
		w[i] = binary.BigEndian.Uint64(b[(16-1-i)*8 : (16-i)*8])
	}
	return fromWords(w)
}

// IsZero reports whether all integer bits are zero.
func (u Uint1024) IsZero() bool { w := u.words(); return uintwide.IsZero(w[:]) }

// Equal reports whether the two unsigned integers are equal.
func (u Uint1024) Equal(v Uint1024) bool { return u == v }

// Compare compares the unsigned values and returns minus one, zero, or one.
func (u Uint1024) Compare(v Uint1024) int {
	a, b := u.words(), v.words()
	return uintwide.Compare(a[:], b[:])
}

// BitLen returns the number of bits required to represent the value.
func (u Uint1024) BitLen() int { w := u.words(); return uintwide.BitLen(w[:]) }

// LeadingZeroBitsCount returns the number of zero bits before the highest set bit.
func (u Uint1024) LeadingZeroBitsCount() int { w := u.words(); return uintwide.LeadingZeros(w[:]) }

// TrailingZeroBitsCount returns the number of zero bits after the lowest set bit.
func (u Uint1024) TrailingZeroBitsCount() int { w := u.words(); return uintwide.TrailingZeros(w[:]) }

// NonZeroBitsCount returns the number of set bits in the value.
func (u Uint1024) NonZeroBitsCount() int { w := u.words(); return uintwide.OnesCount(w[:]) }

// Add adds unsigned integers and reports overflow with an error code.
func (u Uint1024) Add(v Uint1024) (Uint1024, errors.Error) {
	a, b := u.words(), v.words()
	var w [16]uint64
	if uintwide.Add(w[:], a[:], b[:]) != 0 {
		return Zero, errors.Overflow
	}
	return fromWords(w), errors.None
}

// Add64 adds a uint64 and reports overflow with an error code.
func (u Uint1024) Add64(v uint64) (Uint1024, errors.Error) { return u.Add(FromUint64(v)) }

// Sub subtracts unsigned integers and reports underflow with an error code.
func (u Uint1024) Sub(v Uint1024) (Uint1024, errors.Error) {
	a, b := u.words(), v.words()
	var w [16]uint64
	if uintwide.Sub(w[:], a[:], b[:]) != 0 {
		return Zero, errors.Underflow
	}
	return fromWords(w), errors.None
}

// Sub64 subtracts a uint64 and reports underflow with an error code.
func (u Uint1024) Sub64(v uint64) (Uint1024, errors.Error) { return u.Sub(FromUint64(v)) }

// MulCarry returns the low 1024 bits and high 1024 bits of the exact product.
func (u Uint1024) MulCarry(v Uint1024) (Uint1024, Uint1024) {
	a, b := u.words(), v.words()
	// Multiply only significant limbs. The product still uses a fixed-size
	// stack-eligible buffer, and the unused high words remain zero.
	na, nb := len(a), len(b)
	for na > 0 && a[na-1] == 0 {
		na--
	}
	for nb > 0 && b[nb-1] == 0 {
		nb--
	}
	if na == 0 || nb == 0 {
		return Zero, Zero
	}
	var full [32]uint64
	uintwide.MulFull(full[:na+nb], a[:na], b[:nb])
	var lo, hi [16]uint64
	copy(lo[:], full[:16])
	copy(hi[:], full[16:])
	return fromWords(lo), fromWords(hi)
}

// Mul multiplies unsigned integers and reports overflow with an error code.
func (u Uint1024) Mul(v Uint1024) (Uint1024, errors.Error) {
	lo, hi := u.MulCarry(v)
	if !hi.IsZero() {
		return Zero, errors.Overflow
	}
	return lo, errors.None
}

// Mul64 multiplies by a uint64 and reports overflow with an error code.
func (u Uint1024) Mul64(v uint64) (Uint1024, errors.Error) { return u.Mul(FromUint64(v)) }

// QuoRem returns the unsigned quotient, remainder, and a division by zero error when needed.
func (u Uint1024) QuoRem(v Uint1024) (Uint1024, Uint1024, errors.Error) {
	a, b := u.words(), v.words()
	var q, r [16]uint64
	if !uintwide.QuoRem(q[:], r[:], a[:], b[:]) {
		return Zero, Zero, errors.DivisionByZero
	}
	return fromWords(q), fromWords(r), errors.None
}

// QuoRem64 returns the quotient, uint64 remainder, and a division by zero error when needed.
func (u Uint1024) QuoRem64(v uint64) (Uint1024, uint64, errors.Error) {
	a := u.words()
	var q [16]uint64
	r, ok := uintwide.QuoRem64(q[:], a[:], v)
	if !ok {
		return Zero, 0, errors.DivisionByZero
	}
	return fromWords(q), r, errors.None
}

// Div returns the integer quotient, or a division by zero error.
func (u Uint1024) Div(v Uint1024) (Uint1024, errors.Error) { q, _, e := u.QuoRem(v); return q, e }

// Div64 returns the integer quotient after division by a uint64.
func (u Uint1024) Div64(v uint64) (Uint1024, errors.Error) { q, _, e := u.QuoRem64(v); return q, e }

// Mod returns the unsigned remainder, or a division by zero error.
func (u Uint1024) Mod(v Uint1024) (Uint1024, errors.Error) { _, r, e := u.QuoRem(v); return r, e }

// Mod64 returns the uint64 remainder after division by a uint64.
func (u Uint1024) Mod64(v uint64) (uint64, errors.Error) { _, r, e := u.QuoRem64(v); return r, e }

// Lsh shifts bits left by the requested count, discarding bits beyond the width.
func (u Uint1024) Lsh(shift uint) Uint1024 {
	a := u.words()
	var w [16]uint64
	uintwide.Lsh(w[:], a[:], shift)
	return fromWords(w)
}

// Rsh shifts bits right by the requested count.
func (u Uint1024) Rsh(shift uint) Uint1024 {
	a := u.words()
	var w [16]uint64
	uintwide.Rsh(w[:], a[:], shift)
	return fromWords(w)
}

// And returns the bitwise AND of two unsigned integers.
func (u Uint1024) And(v Uint1024) Uint1024 {
	a, b := u.words(), v.words()
	for i := range a {
		a[i] &= b[i]
	}
	return fromWords(a)
}

// And64 returns the bitwise AND with a uint64.
func (u Uint1024) And64(v uint64) Uint1024 { return u.And(FromUint64(v)) }

// Or returns the bitwise OR of two unsigned integers.
func (u Uint1024) Or(v Uint1024) Uint1024 {
	a, b := u.words(), v.words()
	for i := range a {
		a[i] |= b[i]
	}
	return fromWords(a)
}

// Or64 returns the bitwise OR with a uint64.
func (u Uint1024) Or64(v uint64) Uint1024 { return u.Or(FromUint64(v)) }

// Xor returns the bitwise XOR of two unsigned integers.
func (u Uint1024) Xor(v Uint1024) Uint1024 {
	a, b := u.words(), v.words()
	for i := range a {
		a[i] ^= b[i]
	}
	return fromWords(a)
}

// Xor64 returns the bitwise XOR with a uint64.
func (u Uint1024) Xor64(v uint64) Uint1024 { return u.Xor(FromUint64(v)) }

// RotateBitsLeft rotates all bits to the left by the requested count.
func (u Uint1024) RotateBitsLeft(k int) Uint1024 {
	s := uint(k) & (1024 - 1)
	return u.Lsh(s).Or(u.Rsh(1024 - s))
}

// RotateBitsRight rotates all bits to the right by the requested count.
func (u Uint1024) RotateBitsRight(k int) Uint1024 { return u.RotateBitsLeft(-k) }

// ReverseBits returns the value with its bit order reversed.
func (u Uint1024) ReverseBits() Uint1024 {
	a := u.words()
	var w [16]uint64
	for i, word := range a {
		w[16-1-i] = bits.Reverse64(word)
	}
	return fromWords(w)
}

// ReverseBytes returns the value with its byte order reversed.
func (u Uint1024) ReverseBytes() Uint1024 {
	a := u.words()
	var w [16]uint64
	for i, word := range a {
		w[16-1-i] = bits.ReverseBytes64(word)
	}
	return fromWords(w)
}

// Uint64 returns the value as uint64, or an error when it does not fit.
func (u Uint1024) Uint64() (uint64, errors.Error) {
	a := u.words()
	for _, word := range a[1:] {
		if word != 0 {
			return 0, errors.Overflow
		}
	}
	return a[0], errors.None
}

// Uint512 returns the value as Uint512, or an error when it does not fit.
func (u Uint1024) Uint512() (uint512.Uint512, errors.Error) {
	if !u.Hi.IsZero() {
		return uint512.Zero, errors.Overflow
	}
	return u.Lo, errors.None
}

// BigInt returns a new big.Int holding the unsigned value.
func (u Uint1024) BigInt() *big.Int {
	b := u.BytesBigEndian()
	return new(big.Int).SetBytes(b[:])
}

// PutBytes writes the fixed width value into a destination in little endian order.
func (u Uint1024) PutBytes(buf []byte) errors.Error {
	if len(buf) < 128 {
		return errors.NotEnoughBytes
	}
	a := u.words()
	for i, w := range a {
		binary.LittleEndian.PutUint64(buf[i*8:i*8+8], w)
	}
	return errors.None
}

// PutBytesBigEndian writes the fixed width value into a destination in big endian order.
func (u Uint1024) PutBytesBigEndian(buf []byte) errors.Error {
	if len(buf) < 128 {
		return errors.NotEnoughBytes
	}
	a := u.words()
	for i, w := range a {
		binary.BigEndian.PutUint64(buf[(16-1-i)*8:(16-i)*8], w)
	}
	return errors.None
}

// Bytes returns a fixed width byte representation in little endian order.
func (u Uint1024) Bytes() [128]byte { var b [128]byte; _ = u.PutBytes(b[:]); return b }

// BytesBigEndian returns a fixed width byte representation in big endian order.
func (u Uint1024) BytesBigEndian() [128]byte {
	var b [128]byte
	_ = u.PutBytesBigEndian(b[:])
	return b
}

// AppendBytes appends the fixed width bytes in little endian order.
func (u Uint1024) AppendBytes(buf []byte) []byte { b := u.Bytes(); return append(buf, b[:]...) }

// AppendBytesBigEndian appends the fixed width bytes in big endian order.
func (u Uint1024) AppendBytesBigEndian(buf []byte) []byte {
	b := u.BytesBigEndian()
	return append(buf, b[:]...)
}

// String returns the decimal representation of the unsigned integer.
func (u Uint1024) String() string {
	var buf [MaxStrLen]byte
	return string(u.StringToBuf(buf[:]))
}

// StringToBuf writes the decimal digits starting at the end of buf.
// The caller must provide enough space for the result (at most MaxStrLen).
func (u Uint1024) StringToBuf(buf []byte) []byte {
	if u.IsZero() {
		buf[len(buf)-1] = '0'
		return buf[len(buf)-1:]
	}
	pos := len(buf)
	for !u.IsZero() {
		next, rem, _ := u.QuoRem64(10000000000000000000)
		digits := 19
		if next.IsZero() {
			digits = 0
			for r := rem; r > 0; r /= 10 {
				digits++
			}
		}
		for i := 0; i < digits; i++ {
			pos--
			buf[pos] = byte('0' + rem%10)
			rem /= 10
		}
		u = next
	}
	return buf[pos:]
}

// MarshalText writes the decimal representation as text.
func (u Uint1024) MarshalText() ([]byte, error) { return []byte(u.String()), nil }

// UnmarshalText parses an unsigned decimal value from text.
func (u *Uint1024) UnmarshalText(b []byte) error {
	v, e := FromString(b)
	if e != errors.None {
		return e.Value()
	}
	*u = v
	return nil
}

// Scan implements fmt.Scanner, as in uint128.Uint128.
func (u *Uint1024) Scan(state fmt.ScanState, _ rune) error {
	token, e := state.Token(true, nil)
	if e != nil {
		return e
	}
	if len(token) > 0 && token[0] == '-' {
		return errors.Negative.Value()
	}
	v, ec := FromString(token)
	if ec != errors.None {
		return ec.Value()
	}
	*u = v
	return nil
}
