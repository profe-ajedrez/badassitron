package dec128_test

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// Independently verify truncation against arbitrary-precision integer
// division, including 128-bit coefficients and both stored signs. This test
// runs with the default M2 implementation AND the tagged legacy rollback.
func TestM2ERescaleTruncateAgainstBigInt(t *testing.T) {
	rng := rand.New(rand.NewSource(20261117))
	for source := uint8(0); source <= dec128.MaxPrecision; source++ {
		for target := uint8(0); target <= source; target++ {
			divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(source-target)), nil)
			for trial := 0; trial < 24; trial++ {
				coefficient := uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
				negative := rng.Intn(2) == 0
				value := dec128.New(coefficient, source, negative)
				got := value.Rescale(target)
				if got.IsNaN() {
					t.Fatalf("unexpected NaN: source=%d target=%d coefficient=%v error=%v", source, target, coefficient, got.ErrorDetails())
				}
				exact := new(big.Int).SetUint64(coefficient.Hi)
				exact.Lsh(exact, 64)
				exact.Or(exact, new(big.Int).SetUint64(coefficient.Lo))
				exact.Quo(exact, divisor)
				want := uint128.Uint128{Lo: exact.Uint64(), Hi: new(big.Int).Rsh(exact, 64).Uint64()}
				if got.Coefficient() != want || got.Precision() != target || got.StoredNegative() != negative {
					t.Fatalf("source=%d target=%d coeff=%v negative=%t: got=(%v,%d,%t) want=(%v,%d,%t)",
						source, target, coefficient, negative, got.Coefficient(), got.Precision(), got.StoredNegative(), want, target, negative)
				}
			}
		}
	}
}

func TestM2ERescaleTruncateSignedZeroAndErrorPrecedence(t *testing.T) {
	for scale := uint8(0); scale <= dec128.MaxPrecision; scale++ {
		for target := uint8(0); target <= dec128.MaxPrecision; target++ {
			value := dec128.NewStoredZero(scale, true)
			got := value.Rescale(target)
			if got.IsNaN() || !got.IsZero() || !got.StoredNegative() || got.Precision() != target {
				t.Fatalf("negative stored zero at %d -> %d: %v", scale, target, got)
			}
		}
	}
	// A pre-existing NaN is propagated even when the requested scale is
	// invalid. The public error and its ordering are frozen compatibility.
	invalid := dec128.NaN(errors.DivisionByZero)
	if got := invalid.Rescale(255); !got.IsNaN() || got.ErrorDetails() != errors.DivisionByZero.Value() {
		t.Fatalf("lost first error: %v", got)
	}
	if got := dec128.One.Rescale(255); !got.IsNaN() || got.ErrorDetails() != errors.PrecisionOutOfRange.Value() {
		t.Fatalf("invalid target scale: %v", got)
	}
}
