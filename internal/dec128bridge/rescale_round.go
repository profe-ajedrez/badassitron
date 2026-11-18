package dec128bridge

import (
	jk "github.com/jokruger/dec128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
)

// RescaleRoundStrict delegates quantization to Jokruger with a per-call mode.
// It preserves the historical validation order and translates the error in
// this operation, rather than casting numerically incompatible enum values.
// It preserves the stored sign of zero, including the scale-zero state.
func RescaleRoundStrict(v legacy.Dec128, scale uint8, mode legacy.RoundingMode) (legacy.Dec128, bool) {
	if v.IsNaN() {
		return v, false
	}
	if scale > legacy.MaxPrecision {
		return legacy.NaN(le.PrecisionOutOfRange), false
	}
	upstreamMode, ok := toRoundingMode(mode)
	if !ok {
		return legacy.NaN(le.InvalidFormat), false
	}
	out, inexact := ToUpstream(v).RescaleRoundInexact(scale, upstreamMode)
	if out.IsNaN() {
		return legacy.NaN(fromState(out.ErrorDetails())), false
	}
	if out.IsZero() {
		// Exact stored zero retains its sign, including after changing scale.
		// Quantizing a nonzero negative magnitude to zero instead uses the
		// legacy New constructor, which normalizes the resulting zero.
		if v.IsZero() {
			return legacyFiniteZero(scale, v.StoredNegative()), inexact
		}
		return legacyFiniteZero(scale, false), inexact
	}
	return FromUpstream(out), inexact
}

func toRoundingMode(m legacy.RoundingMode) (jk.RoundingMode, bool) {
	switch m {
	case legacy.RoundingTowardZero:
		return jk.ROUND_TOWARD_ZERO, true
	case legacy.RoundingAwayFromZero:
		return jk.ROUND_AWAY_FROM_ZERO, true
	case legacy.RoundingFloor:
		return jk.ROUND_DOWN, true
	case legacy.RoundingCeil:
		return jk.ROUND_UP, true
	case legacy.RoundingHalfTowardZero:
		return jk.ROUND_HALF_TOWARD_ZERO, true
	case legacy.RoundingHalfAwayFromZero:
		return jk.ROUND_HALF_AWAY_FROM_ZERO, true
	case legacy.RoundingHalfEven:
		return jk.ROUND_BANK, true
	default:
		return 0, false
	}
}
