package dec128

import (
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func TestV4RescaleRoundExactNegativeZeroRepresentation(t *testing.T) {
	negativeZero := New(uint128.Zero, 3, true)
	if !negativeZero.neg {
		t.Fatal("test fixture must retain an internal negative-zero sign bit")
	}
	for _, prec := range []uint8{0, 2, 3, 19} {
		for mode := RoundingTowardZero; mode <= RoundingHalfEven; mode++ {
			got, inexact := negativeZero.RescaleRoundWithStatus(prec, mode)
			if got != negativeZero.Rescale(prec) || inexact {
				t.Fatalf("scale %d mode %d: exact signed-zero representation differs from Rescale: %+v, inexact=%t", prec, mode, got, inexact)
			}
		}
	}
}
