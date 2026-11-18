package dec128bridge

import (
	jk "github.com/jokruger/dec128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
	"math/big"
	"math/rand"
	"testing"
)

func sameStrict(a, b legacy.Dec128) bool {
	if a.IsNaN() || b.IsNaN() {
		return a.IsNaN() && b.IsNaN() && a.ErrorDetails() == b.ErrorDetails()
	}
	return a.Coefficient() == b.Coefficient() && a.Precision() == b.Precision() && a.IsNegative() == b.IsNegative()
}
func TestAddSubStrictDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(20261117))
	seeds := []string{"0", "1", "-1", "1.0000000000000000000", "0.0000000000000000001", "340282366920938463463374607431768211455", "34028236692093846346337460743176821145.5", "-99999999999999999999999", "1e20", "1234567890.1234567890"}
	check := func(a, b legacy.Dec128) {
		if want, got := a.Add(b), AddStrict(a, b); !sameStrict(want, got) {
			t.Fatalf("Add a=%v(%d) b=%v(%d): want=%v(%d/%v) got=%v(%d/%v)", a, a.Precision(), b, b.Precision(), want, want.Precision(), want.ErrorDetails(), got, got.Precision(), got.ErrorDetails())
		}
		if want, got := a.Sub(b), SubStrict(a, b); !sameStrict(want, got) {
			t.Fatalf("Sub a=%v(%d) b=%v(%d): want=%v(%d/%v) got=%v(%d/%v)", a, a.Precision(), b, b.Precision(), want, want.Precision(), want.ErrorDetails(), got, got.Precision(), got.ErrorDetails())
		}
	}
	for _, sa := range seeds {
		for _, sb := range seeds {
			check(legacy.FromString(sa), legacy.FromString(sb))
		}
	}
	for i := 0; i < 40000; i++ {
		a := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		b := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		check(a, b)
	}
}

// Independently verify all representable bridge outputs against rational
// arithmetic; the historical implementation is not the only oracle.
func TestAddSubStrictBigRat(t *testing.T) {
	rng := rand.New(rand.NewSource(8012026))
	for i := 0; i < 12000; i++ {
		a := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		b := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		ar, ok := new(big.Rat).SetString(a.StringFixed())
		if !ok {
			t.Fatal("bad a")
		}
		br, ok := new(big.Rat).SetString(b.StringFixed())
		if !ok {
			t.Fatal("bad b")
		}
		for _, tc := range []struct {
			name   string
			result legacy.Dec128
			exact  *big.Rat
		}{
			{"Add", AddStrict(a, b), new(big.Rat).Add(ar, br)},
			{"Sub", SubStrict(a, b), new(big.Rat).Sub(ar, br)},
		} {
			if tc.result.IsNaN() {
				continue
			}
			got, ok := new(big.Rat).SetString(tc.result.StringFixed())
			if !ok || got.Cmp(tc.exact) != 0 {
				t.Fatalf("%s a=%s b=%s result=%s expected=%s", tc.name, a.StringFixed(), b.StringFixed(), tc.result.StringFixed(), tc.exact.RatString())
			}
		}
	}
}

func TestAddSubStrictIndependentOfUpstreamGlobals(t *testing.T) {
	a, b := legacy.FromString("123.4567"), legacy.FromString("0.0000009")
	add, sub := AddStrict(a, b), SubStrict(a, b)
	jk.SetArithmeticRounding(jk.ROUND_UP)
	defer jk.SetArithmeticRounding(jk.ROUND_TOWARD_ZERO)
	if !sameStrict(add, AddStrict(a, b)) || !sameStrict(sub, SubStrict(a, b)) {
		t.Fatal("AddStrict/SubStrict changed under upstream arithmetic rounding setting")
	}
}

var addSubSink legacy.Dec128

func BenchmarkAddStrict(b *testing.B) {
	a, c := legacy.FromString("1234567.890123"), legacy.FromString("987.654321")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		addSubSink = AddStrict(a, c)
	}
}
func BenchmarkAddLegacy(b *testing.B) {
	a, c := legacy.FromString("1234567.890123"), legacy.FromString("987.654321")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		addSubSink = a.Add(c)
	}
}
func BenchmarkSubStrict(b *testing.B) {
	a, c := legacy.FromString("1234567.890123"), legacy.FromString("987.654321")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		addSubSink = SubStrict(a, c)
	}
}
func BenchmarkSubLegacy(b *testing.B) {
	a, c := legacy.FromString("1234567.890123"), legacy.FromString("987.654321")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		addSubSink = a.Sub(c)
	}
}
