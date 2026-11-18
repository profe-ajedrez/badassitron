package dec128bridge

import (
	"math/big"
	"math/rand"
	"testing"

	legacy "github.com/profe-ajedrez/badassitron/dec128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// independentQuantize uses arbitrary-precision rational arithmetic; it does
// not call any Badassitron or Jokruger rounding implementation.
func independentQuantize(value *big.Rat, precision uint8, mode legacy.RoundingMode) (*big.Rat, bool) {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision)), nil)
	scaled := new(big.Rat).Mul(value, new(big.Rat).SetInt(scale))
	q, r := new(big.Int).QuoRem(scaled.Num(), scaled.Denom(), new(big.Int))
	inexact := r.Sign() != 0
	if inexact {
		negative := scaled.Sign() < 0
		inc := false
		switch mode {
		case legacy.RoundingAwayFromZero:
			inc = true
		case legacy.RoundingFloor:
			inc = negative
		case legacy.RoundingCeil:
			inc = !negative
		case legacy.RoundingHalfTowardZero, legacy.RoundingHalfAwayFromZero, legacy.RoundingHalfEven:
			twice := new(big.Int).Mul(new(big.Int).Abs(r), big.NewInt(2))
			cmp := twice.Cmp(scaled.Denom())
			inc = cmp > 0 || (cmp == 0 && (mode == legacy.RoundingHalfAwayFromZero || (mode == legacy.RoundingHalfEven && q.Bit(0) == 1)))
		}
		if inc {
			if negative {
				q.Sub(q, big.NewInt(1))
			} else {
				q.Add(q, big.NewInt(1))
			}
		}
	}
	return new(big.Rat).SetFrac(q, scale), inexact
}

func TestDivRoundAllModesAgainstBigRat(t *testing.T) {
	rng := rand.New(rand.NewSource(2026111722))
	checked := 0
	for i := 0; i < 8000; i++ {
		a := legacy.New(lu.Uint128{Lo: rng.Uint64() % 1000000000000}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		b := legacy.New(lu.Uint128{Lo: 1 + rng.Uint64()%1000000000000}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		if i%11 == 0 {
			a = legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, 19, rng.Intn(2) == 0)
		}
		if i%13 == 0 {
			b = legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, 0, rng.Intn(2) == 0)
		}
		precision := uint8(rng.Intn(20))
		ar, _ := new(big.Rat).SetString(a.StringFixed())
		br, _ := new(big.Rat).SetString(b.StringFixed())
		ratio := new(big.Rat).Quo(ar, br)
		for mode := legacy.RoundingMode(0); mode <= legacy.RoundingHalfEven; mode++ {
			got, inexact := DivRoundWithStatus(a, b, precision, mode)
			if got.IsNaN() {
				continue
			} // 128-bit overflow is covered by the differential tests.
			expected, wantInexact := independentQuantize(ratio, precision, mode)
			actual, ok := new(big.Rat).SetString(got.StringFixed())
			if !ok || actual.Cmp(expected) != 0 || inexact != wantInexact {
				t.Fatalf("case=%d mode=%d a=%s b=%s scale=%d got=%s status=%v want=%s status=%v", i, mode, a.StringFixed(), b.StringFixed(), precision, got.StringFixed(), inexact, expected.RatString(), wantInexact)
			}
			checked++
		}
	}
	if checked < 30000 {
		t.Fatalf("insufficient independent checks: %d", checked)
	}
	t.Logf("validated %d division/rounding results against math/big", checked)
}

func TestSignedZeroIsNotLosslesslyRepresentableUpstream(t *testing.T) {
	old := legacy.New(lu.Zero, 3, true)
	positive := legacy.New(lu.Zero, 3, false)
	if old != positive && old.StringFixed() != positive.StringFixed() {
		t.Fatal("historical zero formatting changed")
	}
	// Jokruger intentionally normalizes zero's sign. This test documents
	// why the future public wrapper must own any signed-zero metadata.
	if ToUpstream(old) != ToUpstream(positive) {
		t.Fatal("upstream unexpectedly retains the historical zero sign")
	}
	if old == positive {
		t.Fatal("historical representation unexpectedly lost signed zero")
	}
}
