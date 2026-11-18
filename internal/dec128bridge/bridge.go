// Package dec128bridge is an experimental, internal migration adapter.
// It delegates arithmetic to the imported github.com/jokruger/dec128 module.
// It is NOT yet wired into the published badassitron/dec128 API.
package dec128bridge

import (
	jk "github.com/jokruger/dec128"
	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
	"math/bits"
)

// ToUpstream preserves the coefficient, stored scale, and sign of finite values.
// NaN reasons are translated explicitly rather than reusing ordinal codes.
func ToUpstream(v legacy.Dec128) jk.Dec128 {
	if v.IsNaN() {
		return jk.NaN(toState(v.ErrorDetails()))
	}
	c := v.Coefficient()
	return jk.New(ju.Uint128{Lo: c.Lo, Hi: c.Hi}, v.Precision(), v.IsNegative())
}

// FromUpstream preserves finite values without parsing or string allocation.
// New upstream-only errors are intentionally mapped to a historical category
// until the migration contract authorizes public error-code extensions.
func FromUpstream(v jk.Dec128) legacy.Dec128 {
	if v.IsNaN() {
		return legacy.NaN(fromState(v.ErrorDetails()))
	}
	c := v.Coefficient()
	return legacy.New(lu.Uint128{Lo: c.Lo, Hi: c.Hi}, v.Scale(), v.IsNegative())
}

// MulStrict preserves the historical exact-or-NaN multiplication contract.
// Its integer arithmetic comes from the imported Jokruger module. In
// particular, the 256-bit product and divide are not copied into this repo.
// Unlike upstream Mul, it does not read any process-wide rounding setting.
func MulStrict(a, b legacy.Dec128) legacy.Dec128 {
	if a.IsNaN() {
		return a
	}
	if b.IsNaN() {
		return b
	}
	if a.IsZero() || b.IsZero() {
		return legacy.Zero
	}
	// The preexisting 64x64 exact fast path is a small standard-library
	// operation, not copied upstream code. Retain it to avoid adapter cost.
	ac, bc := a.Coefficient(), b.Coefficient()
	scale := a.Precision() + b.Precision()
	if ac.Hi == 0 && bc.Hi == 0 && scale <= jk.MaxScale {
		hi, lo := bits.Mul64(ac.Lo, bc.Lo)
		return legacy.New(lu.Uint128{Lo: lo, Hi: hi}, scale, a.IsNegative() != b.IsNegative())
	}
	if out, ok := mulExact(a, b); ok {
		return out
	}
	// Badassitron historically retries with operand trailing zeros stripped.
	if out, ok := mulExact(canonicalOperand(a), canonicalOperand(b)); ok {
		return out
	}
	return legacy.NaN(le.Overflow)
}

func mulExact(a, b legacy.Dec128) (legacy.Dec128, bool) {
	scale := a.Precision() + b.Precision() // 0..38
	ac, bc := a.Coefficient(), b.Coefficient()
	lo, hi := (ju.Uint128{Lo: ac.Lo, Hi: ac.Hi}).MulCarry(ju.Uint128{Lo: bc.Lo, Hi: bc.Hi})
	neg := a.IsNegative() != b.IsNegative()
	result := func(c ju.Uint128, s uint8) legacy.Dec128 {
		return legacy.New(lu.Uint128{Lo: c.Lo, Hi: c.Hi}, s, neg)
	}

	if hi.IsZero() {
		if scale <= jk.MaxScale {
			return result(lo, scale), true
		}
		// The legacy implementation canonicalizes the entire product when
		// its scale exceeds the maximum, rather than merely limiting to 19.
		for scale > 0 {
			q, r, st := lo.QuoRem64(10)
			if st != state.OK || r != 0 {
				break
			}
			lo = q
			scale--
		}
		if scale <= jk.MaxScale {
			return result(lo, scale), true
		}
		return legacy.Dec128{}, false
	}

	// When the full product does not fit in 128 bits, Badassitron chooses
	// the *largest exact decimal reduction* first, not necessarily the
	// maximum retained scale. Preserve that observable scale contract.
	for drop := scale; drop > 0; drop-- {
		q, r, st := ju.QuoRem256By128(lo, hi, ju.Pow10Uint128[drop])
		if st == state.OK && r.IsZero() {
			return result(q, scale-drop), true
		}
		if st == state.Overflow {
			return legacy.Dec128{}, false
		}
		if scale-drop > jk.MaxScale {
			return legacy.Dec128{}, false
		}
	}
	return legacy.Dec128{}, false
}

func toState(err error) state.State {
	switch err {
	case nil:
		return state.NaN
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
	// The prototype has no public extension for newer errors.
	default:
		return le.NotANumber
	}
}

// canonicalOperand uses imported unsigned primitives for the bridge fallback.
// It must not depend on any arithmetic or normalization method on the public
// historical decimal type. Exact fractional zeros are removed in the same
// order as the historical Canonical implementation.
func canonicalOperand(v legacy.Dec128) legacy.Dec128 {
	if v.IsNaN() {
		return v
	}
	if v.IsZero() {
		return legacy.Zero
	}
	c := v.Coefficient()
	u := ju.Uint128{Lo: c.Lo, Hi: c.Hi}
	scale := v.Precision()
	for scale > 0 {
		q, rem, status := u.QuoRem64(10)
		if status != state.OK || rem != 0 {
			break
		}
		u, scale = q, scale-1
	}
	return legacy.New(lu.Uint128{Lo: u.Lo, Hi: u.Hi}, scale, v.IsNegative())
}
