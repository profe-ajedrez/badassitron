package dec128bridge

import (
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
	"math/rand"
	"testing"
)

func TestRescaleRoundStrictDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(103617))
	modes := []legacy.RoundingMode{legacy.RoundingTowardZero, legacy.RoundingAwayFromZero, legacy.RoundingFloor, legacy.RoundingCeil, legacy.RoundingHalfTowardZero, legacy.RoundingHalfAwayFromZero, legacy.RoundingHalfEven}
	samples := []string{"1.005", "-1.005", "0.0000000000000000001", "-0.0000000000000000001", "1.9999999999999999999", "-1.9999999999999999999", "99999999999999999999.99", "0.0000", "-0.0000", "340282366920938463463374607431768211455"}
	check := func(a legacy.Dec128, scale uint8, mode legacy.RoundingMode) {
		want, was := a.RescaleRoundWithStatus(scale, mode)
		got, is := RescaleRoundStrict(a, scale, mode)
		if was != is || !sameStrict(want, got) {
			t.Fatalf("a=%s scale=%d mode=%d want=%s %v %v got=%s %v %v", a, scale, mode, want, want.ErrorDetails(), was, got, got.ErrorDetails(), is)
		}
	}
	for _, s := range samples {
		for scale := uint8(0); scale <= 19; scale++ {
			for _, mode := range modes {
				check(legacy.FromString(s), scale, mode)
			}
		}
	}
	for i := 0; i < 18000; i++ {
		a := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		check(a, uint8(rng.Intn(20)), modes[rng.Intn(len(modes))])
	}
	t.Log("validated 19400 rescale/mode combinations")
}
func TestRescaleRoundStrictErrors(t *testing.T) {
	a := legacy.FromString("1.23")
	for _, tc := range []struct {
		scale uint8
		mode  legacy.RoundingMode
	}{{20, legacy.RoundingTowardZero}, {1, legacy.RoundingMode(99)}} {
		want, was := a.RescaleRoundWithStatus(tc.scale, tc.mode)
		got, is := RescaleRoundStrict(a, tc.scale, tc.mode)
		if !sameStrict(want, got) || was != is {
			t.Fatalf("error mismatch want=%v got=%v", want.ErrorDetails(), got.ErrorDetails())
		}
	}
}
