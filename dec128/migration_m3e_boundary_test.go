package dec128

import (
	"math/rand"
	"testing"

	ju "github.com/jokruger/dec128/uint128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// The decimal engine never needs to visit the historical public Uint128
// implementation to construct an internally computed coefficient. This
// invariant must survive both uint128 build-tag choices.
func TestMigrationM3EInternalCoefficientBoundary(t *testing.T) {
	rng := rand.New(rand.NewSource(20261117))
	for i := 0; i < 4000; i++ {
		coef := ju.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		if i%17 == 0 {
			coef = ju.Zero
		}
		scale := uint8(i % (int(MaxPrecision) + 2))
		negative := i%2 == 0
		want := New(lu.Uint128(coef), scale, negative)
		got := fromParts(coef, scale, negative)
		if got != want {
			t.Fatalf("case=%d coefficient=%#v scale=%d negative=%v: internal=%#v public=%#v", i, coef, scale, negative, got, want)
		}
		if scale <= MaxPrecision && !got.IsNaN() && got.Coefficient() != lu.Uint128(coef) {
			t.Fatalf("case=%d lost coefficient", i)
		}
	}
	if New(lu.Zero, 0, true) != Zero || fromParts(ju.Zero, 0, true) != Zero {
		t.Fatal("public New and internal constructor must normalize zero at scale 0")
	}
	if z := NewStoredZero(0, true); !z.IsZero() || !z.StoredNegative() {
		t.Fatal("explicit signed-zero constructor must continue to retain negative zero")
	}
}
