package dec128bridge

import (
	"math/big"
	"math/rand"
	"testing"

	jk "github.com/jokruger/dec128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func TestRoundedOpsDifferential(t *testing.T) {
	r := rand.New(rand.NewSource(2026111707))
	seeds := []string{"0", "-0.0000000000000000000", "1", "-1", "0.0000000000000000001", "0.5", "1.5", "2.5", "1e20", "340282366920938463463374607431768211455", "1234567.89012", "-123.0000000000000000001"}
	check := func(a, b legacy.Dec128, p uint8, mode legacy.RoundingMode) {
		wm, wi := a.MulRoundWithStatus(b, p, mode)
		gm, gi := MulRoundWithStatus(a, b, p, mode)
		if !sameStrict(wm, gm) || wi != gi {
			t.Fatalf("mul a=%s/%d b=%s/%d p=%d mode=%d want=%s/%d %t %v got=%s/%d %t %v", a, a.Precision(), b, b.Precision(), p, mode, wm, wm.Precision(), wi, wm.ErrorDetails(), gm, gm.Precision(), gi, gm.ErrorDetails())
		}
		wd, wdi := a.DivRoundWithStatus(b, p, mode)
		gd, gdi := DivRoundWithStatus(a, b, p, mode)
		if !sameStrict(wd, gd) || wdi != gdi {
			t.Fatalf("div a=%s/%d b=%s/%d p=%d mode=%d want=%s/%d %t %v got=%s/%d %t %v", a, a.Precision(), b, b.Precision(), p, mode, wd, wd.Precision(), wdi, wd.ErrorDetails(), gd, gd.Precision(), gdi, gd.ErrorDetails())
		}
	}
	for _, sa := range seeds {
		for _, sb := range seeds {
			for _, p := range []uint8{0, 1, 2, 9, 18, 19, 20} {
				for m := legacy.RoundingMode(0); m <= 7; m++ {
					check(legacy.FromString(sa), legacy.FromString(sb), p, m)
				}
			}
		}
	}
	for i := 0; i < 30000; i++ {
		a := legacy.New(lu.Uint128{Lo: r.Uint64(), Hi: r.Uint64()}, uint8(r.Intn(20)), r.Intn(2) == 0)
		b := legacy.New(lu.Uint128{Lo: r.Uint64(), Hi: r.Uint64()}, uint8(r.Intn(20)), r.Intn(2) == 0)
		if i%4 == 0 {
			a = legacy.New(lu.Uint128{Lo: r.Uint64() % 10000000}, uint8(r.Intn(20)), r.Intn(2) == 0)
		}
		if i%7 == 0 {
			b = legacy.New(lu.Uint128{Lo: r.Uint64() % 10000000}, uint8(r.Intn(20)), r.Intn(2) == 0)
		}
		if i%13 == 0 {
			b = legacy.Zero
		}
		check(a, b, uint8(r.Intn(22)), legacy.RoundingMode(r.Intn(9)))
	}
}

func TestRoundedOpsIndependentOfGlobals(t *testing.T) {
	a, b := legacy.FromString("123.456789"), legacy.FromString("3.456789")
	m1, mi1 := MulRoundWithStatus(a, b, 8, legacy.RoundingHalfEven)
	d1, di1 := DivRoundWithStatus(a, b, 8, legacy.RoundingHalfEven)
	jk.SetArithmeticRounding(jk.ROUND_UP)
	defer jk.SetArithmeticRounding(jk.ROUND_TOWARD_ZERO)
	m2, mi2 := MulRoundWithStatus(a, b, 8, legacy.RoundingHalfEven)
	d2, di2 := DivRoundWithStatus(a, b, 8, legacy.RoundingHalfEven)
	if !sameStrict(m1, m2) || mi1 != mi2 || !sameStrict(d1, d2) || di1 != di2 {
		t.Fatal("upstream globals changed explicit operations")
	}
}

func TestRoundedOpsExactRational(t *testing.T) {
	rng := rand.New(rand.NewSource(817231))
	checked := 0
	for i := 0; i < 4000; i++ {
		a := legacy.New(lu.Uint128{Lo: rng.Uint64() % 1e12}, uint8(rng.Intn(10)), rng.Intn(2) == 0)
		b := legacy.New(lu.Uint128{Lo: 1 + rng.Uint64()%1e9}, uint8(rng.Intn(10)), rng.Intn(2) == 0)
		p := uint8(rng.Intn(20))
		ar, _ := new(big.Rat).SetString(a.StringFixed())
		br, _ := new(big.Rat).SetString(b.StringFixed())
		for _, tc := range []struct {
			got  legacy.Dec128
			want *big.Rat
		}{
			{func() legacy.Dec128 { v, _ := MulRoundWithStatus(a, b, p, legacy.RoundingTowardZero); return v }(), new(big.Rat).Mul(ar, br)},
			{func() legacy.Dec128 { v, _ := DivRoundWithStatus(a, b, p, legacy.RoundingTowardZero); return v }(), new(big.Rat).Quo(ar, br)},
		} {
			if tc.got.IsNaN() {
				continue
			}
			g, _ := new(big.Rat).SetString(tc.got.StringFixed())
			scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(p)), nil)
			scaled := new(big.Rat).Mul(tc.want, new(big.Rat).SetInt(scale))
			q := new(big.Int).Quo(scaled.Num(), scaled.Denom())
			expected := new(big.Rat).SetFrac(q, scale)
			if g.Cmp(expected) != 0 {
				t.Fatalf("unexpected truncation got=%s expected=%s a=%s b=%s p=%d", g.RatString(), expected.RatString(), a, b, p)
			}
			checked++
		}
	}
	t.Logf("exact rational checks: %d", checked)
}

var roundedSink legacy.Dec128

func BenchmarkMulRoundBridge(b *testing.B) {
	a, c := legacy.FromString("12345.6789"), legacy.FromString("0.00001234567")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		roundedSink, _ = MulRoundWithStatus(a, c, 12, legacy.RoundingHalfEven)
	}
}
func BenchmarkMulRoundLegacy(b *testing.B) {
	a, c := legacy.FromString("12345.6789"), legacy.FromString("0.00001234567")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		roundedSink, _ = a.MulRoundWithStatus(c, 12, legacy.RoundingHalfEven)
	}
}
func BenchmarkDivRoundBridge(b *testing.B) {
	a, c := legacy.FromString("12345.6789"), legacy.FromString("3.00001234567")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		roundedSink, _ = DivRoundWithStatus(a, c, 12, legacy.RoundingHalfEven)
	}
}
func BenchmarkDivRoundLegacy(b *testing.B) {
	a, c := legacy.FromString("12345.6789"), legacy.FromString("3.00001234567")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		roundedSink, _ = a.DivRoundWithStatus(c, 12, legacy.RoundingHalfEven)
	}
}
