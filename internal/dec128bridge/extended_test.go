package dec128bridge

import (
	"math/rand"
	"testing"

	jk "github.com/jokruger/dec128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func TestMulStrictBoundaryDifferential(t *testing.T) {
	cases := []string{
		"0", "-0.0000000000000000000", "1", "-1", "1.0", "2.0000000000000000000",
		"0.0000000000000000001", "0.0000000000000000002", "0.0000000000000000005",
		"0.0000000000000000010", "9999999999999999999", "100000000000000000000",
		"340282366920938463463374607431768211455", "-340282366920938463463374607431768211455",
		"34028236692093846346337460743176821145.5", "10.0000000000000000000",
		"1.2345678901234567890", "0.1234567890123456789", "2.5", "4.0", "20.00",
	}
	for _, sa := range cases {
		for _, sb := range cases {
			a, b := legacy.FromString(sa), legacy.FromString(sb)
			assertMulEqual(t, sa+" * "+sb, a, b)
		}
	}
	rng := rand.New(rand.NewSource(20261117))
	for i := 0; i < 30000; i++ {
		a := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		b := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		if i%2 == 0 {
			a = legacy.New(lu.Uint128{Lo: rng.Uint64() % 10000000}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		}
		if i%3 == 0 {
			b = legacy.New(lu.Uint128{Lo: rng.Uint64() % 10000000}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		}
		if i%5 == 0 {
			a = a.Canonical()
		}
		if i%7 == 0 {
			b = b.Canonical()
		}
		assertMulEqual(t, "random", a, b)
	}
}

func assertMulEqual(t *testing.T, context string, a, b legacy.Dec128) {
	t.Helper()
	got, want := MulStrict(a, b), a.Mul(b)
	if got.IsNaN() != want.IsNaN() || got.StringFixed() != want.StringFixed() || got.Precision() != want.Precision() || got.IsNegative() != want.IsNegative() || got.ErrorDetails() != want.ErrorDetails() {
		t.Fatalf("%s: a=%s(scale=%d) b=%s(scale=%d), got=%s(scale=%d,err=%v), want=%s(scale=%d,err=%v)", context, a.StringFixed(), a.Precision(), b.StringFixed(), b.Precision(), got.StringFixed(), got.Precision(), got.ErrorDetails(), want.StringFixed(), want.Precision(), want.ErrorDetails())
	}
}

func TestMulStrictDoesNotUseUpstreamGlobals(t *testing.T) {
	a, b := legacy.FromString("1.0000000000000000000"), legacy.FromString("2")
	want := MulStrict(a, b)
	jk.SetLossPolicy(jk.LossRound)
	jk.SetArithmeticRounding(jk.ROUND_UP)
	defer jk.SetArithmeticRounding(jk.ROUND_TOWARD_ZERO)
	got := MulStrict(a, b)
	if got.StringFixed() != want.StringFixed() || got.Precision() != want.Precision() {
		t.Fatalf("global dependency: got %v expected %v", got, want)
	}
}
