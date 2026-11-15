package optimus

import (
	"math/rand"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

// Compare the fixed-width route with the exact pre-optimization implementation
// across coefficient tiers, rounding modes, boundary scales and overflows.
func TestNativeRescaleBoundaryDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(20261113))
	coeffs := []uint512.Uint512{uint512.Zero, uint512.One, uint512.Max, uint512.Pow10Uint512[154], uint512.Pow10Uint512[38], uint512.Pow10Uint512[76]}
	for i := 0; i < 300; i++ {
		var data [64]byte
		_, _ = rng.Read(data[:])
		coeffs = append(coeffs, uint512.FromBytes(data))
	}
	scales := []uint8{0, 1, 2, 10, 19, 38, 76, 100, 153, 154, 155, 255}
	modes := []RoundingMode{RoundingTowardZero, RoundingAwayFromZero, RoundingFloor, RoundingCeil, RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven, 7, 255}
	n := 0
	for _, coefficient := range coeffs {
		for _, neg := range []bool{false, true} {
			for _, inScale := range []uint8{0, 1, 2, 38, 76, 153, 154} {
				d := New(coefficient, inScale, neg)
				if d.IsNaN() {
					t.Fatal("unexpected NaN in input")
				}
				for _, outScale := range scales {
					for _, mode := range modes {
						want, wInexact := legacyRescaleOracle(d, outScale, mode)
						got, gInexact := d.RescaleRoundWithStatus(outScale, mode)
						if !rescaleEquivalent(got, want) || gInexact != wInexact {
							t.Fatalf("case=%d coefficient=%s sign=%v inputScale=%d outputScale=%d rounding=%d; old=%s/%v/%v new=%s/%v/%v",
								n, coefficient.String(), neg, inScale, outScale, mode,
								want.StringFixed(), want.ErrorDetails(), wInexact,
								got.StringFixed(), got.ErrorDetails(), gInexact)
						}
						n++
					}
				}
			}
		}
	}
	t.Logf("verified %d boundary/mode/sign/scale combinations", n)
	for _, err := range []errors.Error{errors.Overflow, errors.NotANumber, errors.PrecisionOutOfRange} {
		d := NaN(err)
		got, inexact := d.RescaleRoundWithStatus(2, RoundingHalfEven)
		if !got.IsNaN() || got.ErrorDetails() != d.ErrorDetails() || inexact {
			t.Fatalf("error propagation: %v", err)
		}
	}
}

func rescaleEquivalent(a, b Optimus) bool {
	if a.IsNaN() != b.IsNaN() || a.Precision() != b.Precision() || a.Width() != b.Width() || a.IsNegative() != b.IsNegative() || a.IsZero() != b.IsZero() {
		return false
	}
	if a.IsNaN() {
		return a.ErrorDetails() == b.ErrorDetails()
	}
	return a.Coefficient() == b.Coefficient()
}

func BenchmarkNativeRescale(b *testing.B) {
	cases := []struct {
		name  string
		input string
		scale uint8
		mode  RoundingMode
	}{
		{"same_scale", "10.25", 2, RoundingHalfEven},
		{"round_half_even", "1234.5678", 2, RoundingHalfEven},
		{"round_negative_floor", "-1234.5678", 2, RoundingFloor},
		{"wide_76_digits", "1234567890123456789012345678901234567890123456789012345678901234567890123456.123456789", 2, RoundingHalfEven},
		{"widen_scale", "123.45", 20, RoundingHalfAwayFromZero},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			d := FromString(tc.input)
			if d.IsNaN() {
				b.Fatal(d.ErrorDetails())
			}
			b.ReportAllocs()
			var result Optimus
			for i := 0; i < b.N; i++ {
				result = d.RescaleRound(tc.scale, tc.mode)
			}
			if result.IsNaN() {
				b.Fatal(result.ErrorDetails())
			}
		})
	}
}
