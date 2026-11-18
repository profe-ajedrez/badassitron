package dec128bridge

import (
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
	"math/big"
	"math/rand"
	"testing"
)

func TestDivAtPrecisionDifferential(t *testing.T) {
	// The historical default is 19; keep it fixed for this sequential test.
	legacy.SetDefaultPrecision(19)
	defer legacy.SetDefaultPrecision(19)
	samples := []string{"0", "-0.0000000000000000000", "1", "-1", "2", "3", "0.001", "0.0000000000000000001", "1.0000000000000000000", "1e20", "340282366920938463463374607431768211455", "12345.67890"}
	count := 0
	for _, sa := range samples {
		for _, sb := range samples {
			a, b := legacy.FromString(sa), legacy.FromString(sb)
			want, got := a.Div(b), DivAtPrecision(a, b, 19)
			if !sameStrict(want, got) {
				t.Fatalf("boundary a=%s b=%s: want=%s %d %v got=%s %d %v", sa, sb, want, want.Precision(), want.ErrorDetails(), got, got.Precision(), got.ErrorDetails())
			}
			count++
		}
	}
	rng := rand.New(rand.NewSource(120241117))
	for i := 0; i < 25000; i++ {
		a := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		b := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		if i%3 == 0 {
			a = legacy.New(lu.Uint128{Lo: rng.Uint64() % 10000000}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		}
		if i%7 == 0 {
			b = legacy.New(lu.Uint128{Lo: rng.Uint64() % 10000000}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		}
		if i%11 == 0 {
			b = legacy.Zero
		}
		want, got := a.Div(b), DivAtPrecision(a, b, 19)
		if !sameStrict(want, got) {
			t.Fatalf("random %d: a=%s/%d b=%s/%d want=%s/%d/%v got=%s/%d/%v", i, a, a.Precision(), b, b.Precision(), want, want.Precision(), want.ErrorDetails(), got, got.Precision(), got.ErrorDetails())
		}
		count++
	}
	t.Logf("compared %d division pairs", count)
}

func TestDivAtPrecisionExactRationalTruncation(t *testing.T) {
	rng := rand.New(rand.NewSource(7351))
	checked := 0
	for i := 0; i < 4500; i++ {
		a := legacy.New(lu.Uint128{Lo: rng.Uint64() % 10000000000000}, uint8(rng.Intn(14)), rng.Intn(2) == 0)
		b := legacy.New(lu.Uint128{Lo: 1 + rng.Uint64()%100000000}, uint8(rng.Intn(10)), rng.Intn(2) == 0)
		got := DivAtPrecision(a, b, 19)
		if got.IsNaN() {
			continue
		}
		ar, _ := new(big.Rat).SetString(a.StringFixed())
		br, _ := new(big.Rat).SetString(b.StringFixed())
		rational := new(big.Rat).Quo(ar, br)
		pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(got.Precision())), nil)
		scaled := new(big.Rat).Mul(rational, new(big.Rat).SetInt(pow))
		q := new(big.Int).Quo(scaled.Num(), scaled.Denom())
		expected := new(big.Int).SetUint64(got.Coefficient().Hi)
		expected.Lsh(expected, 64)
		expected.Add(expected, new(big.Int).SetUint64(got.Coefficient().Lo))
		if got.IsNegative() {
			expected.Neg(expected)
		}
		if expected.Cmp(q) != 0 {
			t.Fatalf("truncation mismatch %s / %s: got %s target %s", a, b, got, q)
		}
		checked++
	}
	t.Logf("independent rational checks: %d", checked)
}

var divBridgeSink legacy.Dec128

func BenchmarkDivAtPrecision(b *testing.B) {
	a, c := legacy.FromString("1234567.890123"), legacy.FromString("987.654321")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		divBridgeSink = DivAtPrecision(a, c, 19)
	}
}
func BenchmarkDivLegacyForBridge(b *testing.B) {
	a, c := legacy.FromString("1234567.890123"), legacy.FromString("987.654321")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		divBridgeSink = a.Div(c)
	}
}
