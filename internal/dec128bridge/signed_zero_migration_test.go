package dec128bridge

import (
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
	"testing"
)

// Historical zero metadata includes the sign bit even after quantization
// produces a zero coefficient. String-based comparisons cannot detect this.
func TestMigrationRescaleSignedZeroPropagation(t *testing.T) {
	modes := []legacy.RoundingMode{legacy.RoundingTowardZero, legacy.RoundingAwayFromZero, legacy.RoundingFloor, legacy.RoundingCeil, legacy.RoundingHalfTowardZero, legacy.RoundingHalfAwayFromZero, legacy.RoundingHalfEven}
	for inputScale := uint8(0); inputScale <= 19; inputScale++ {
		for _, negative := range []bool{false, true} {
			initial := legacy.New(lu.Zero, inputScale, negative)
			// Rescale can produce a negative zero with scale zero, which New normalizes.
			if negative && inputScale == 0 {
				initial = legacy.New(lu.Zero, 1, true).Rescale(0)
			}
			for outputScale := uint8(0); outputScale <= 19; outputScale++ {
				for _, mode := range modes {
					want, wi := initial.RescaleRoundWithStatus(outputScale, mode)
					got, gi := RescaleRoundStrict(initial, outputScale, mode)
					if wi != gi || want.IsNaN() != got.IsNaN() || want.Precision() != got.Precision() || want.Coefficient() != got.Coefficient() || want.StoredNegative() != got.StoredNegative() {
						t.Fatalf("input=%d/%t output=%d mode=%d want=%#v got=%#v status=%v/%v", inputScale, negative, outputScale, mode, want, got, wi, gi)
					}
				}
			}
		}
	}
}

func TestMigrationQuantizedNegativeToZero(t *testing.T) {
	modes := []legacy.RoundingMode{legacy.RoundingTowardZero, legacy.RoundingAwayFromZero, legacy.RoundingFloor, legacy.RoundingCeil, legacy.RoundingHalfTowardZero, legacy.RoundingHalfAwayFromZero, legacy.RoundingHalfEven}
	for _, input := range []string{"-0.0000000000000000001", "-0.001", "-0.5", "0.0000000000000000001", "0.001", "0.5"} {
		original := legacy.FromString(input)
		for p := uint8(0); p <= 19; p++ {
			for _, mode := range modes {
				want, wi := original.RescaleRoundWithStatus(p, mode)
				got, gi := RescaleRoundStrict(original, p, mode)
				if !want.IsNaN() && want.IsZero() && (got.StoredNegative() != want.StoredNegative() || got.Precision() != want.Precision() || gi != wi) {
					t.Fatalf("input=%s p=%d mode=%d want=%#v got=%#v", input, p, mode, want, got)
				}
			}
		}
	}
}

func TestMigrationArithmeticSignedZeroDifferential(t *testing.T) {
	values := []legacy.Dec128{
		legacy.Zero, legacy.New(lu.Zero, 1, true), legacy.New(lu.Zero, 3, true),
		legacy.New(lu.Zero, 19, true), legacy.New(lu.Zero, 3, false),
		legacy.New(lu.Zero, 1, true).Rescale(0),
		legacy.FromString("-0.001"), legacy.FromString("0.001"),
		legacy.FromString("-1.25"), legacy.FromString("1.25"),
	}
	same := func(a, b legacy.Dec128) bool {
		if a.IsNaN() || b.IsNaN() {
			return a.IsNaN() && b.IsNaN() && a.ErrorDetails() == b.ErrorDetails()
		}
		return a.Coefficient() == b.Coefficient() && a.Precision() == b.Precision() && a.StoredNegative() == b.StoredNegative()
	}
	for ai, a := range values {
		for bi, b := range values {
			for _, pair := range []struct {
				label     string
				want, got legacy.Dec128
			}{
				{"add", a.Add(b), AddStrict(a, b)}, {"sub", a.Sub(b), SubStrict(a, b)},
				{"mul", a.Mul(b), MulStrict(a, b)},
				{"div", a.Div(b), DivAtPrecision(a, b, 19)},
			} {
				if !same(pair.want, pair.got) {
					t.Fatalf("a[%d]=%#v b[%d]=%#v %s want=%#v got=%#v", ai, a, bi, b, pair.label, pair.want, pair.got)
				}
			}
			for _, p := range []uint8{0, 2, 19} {
				for _, mode := range []legacy.RoundingMode{legacy.RoundingTowardZero, legacy.RoundingHalfAwayFromZero, legacy.RoundingHalfEven} {
					mw, mi := a.MulRoundWithStatus(b, p, mode)
					mg, mgi := MulRoundWithStatus(a, b, p, mode)
					if mi != mgi || !same(mw, mg) {
						t.Fatalf("mulround a[%d] b[%d] p=%d mode=%d want=%#v got=%#v", ai, bi, p, mode, mw, mg)
					}
					dw, di := a.DivRoundWithStatus(b, p, mode)
					dg, dgi := DivRoundWithStatus(a, b, p, mode)
					if di != dgi || !same(dw, dg) {
						t.Fatalf("divround a[%d] b[%d] p=%d mode=%d want=%#v got=%#v", ai, bi, p, mode, dw, dg)
					}
				}
			}
		}
	}
}
