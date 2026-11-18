package dec128bridge

import (
	jk "github.com/jokruger/dec128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	"math/rand"
	"testing"
)

func TestFiniteRoundTrip(t *testing.T) {
	seeds := []string{"0", "-0.00", "12.3400", "-1.235", "0.0000000000000000001", "340282366920938463463374607431768211455", "5."}
	for _, s := range seeds {
		x := legacy.FromString(s)
		if x.IsNaN() {
			continue
		}
		y := FromUpstream(ToUpstream(x))
		if x.StringFixed() != y.StringFixed() || x.Precision() != y.Precision() || x.IsNegative() != y.IsNegative() {
			t.Fatalf("roundtrip %q: %q/%d vs %q/%d", s, x.StringFixed(), x.Precision(), y.StringFixed(), y.Precision())
		}
	}
}

func TestMulStrictDifferentialSmall(t *testing.T) {
	rng := rand.New(rand.NewSource(20261117))
	for i := 0; i < 3000; i++ {
		a := legacy.FromInt64(int64(rng.Intn(2000000) - 1000000))
		b := legacy.FromInt64(int64(rng.Intn(2000000) - 1000000))
		// Stored scale alignment is significant to the historical API.
		a = a.Rescale(uint8(rng.Intn(8)))
		b = b.Rescale(uint8(rng.Intn(8)))
		want, got := a.Mul(b), MulStrict(a, b)
		if got.IsNaN() != want.IsNaN() || got.StringFixed() != want.StringFixed() || got.Precision() != want.Precision() {
			t.Fatalf("%d: %s * %s: want %s [%d], got %s [%d]", i, a.StringFixed(), b.StringFixed(), want.StringFixed(), want.Precision(), got.StringFixed(), got.Precision())
		}
	}
}

func TestUpstreamGlobalIndependentStrict(t *testing.T) {
	// Importing upstream is not a copy. The strict operation passes the mode
	// explicitly and does not rely on the mutable upstream arithmetic globals.
	a, b := legacy.FromString("1.25"), legacy.FromString("2.0")
	baseline := MulStrict(a, b)
	jk.SetArithmeticRounding(jk.ROUND_UP)
	defer jk.SetArithmeticRounding(jk.ROUND_TOWARD_ZERO)
	got := MulStrict(a, b)
	if baseline.StringFixed() != got.StringFixed() || baseline.Precision() != got.Precision() {
		t.Fatalf("global affected strict Mul: %v vs %v", baseline, got)
	}
}

func BenchmarkMulStrict(b *testing.B) {
	a, c := legacy.FromString("1234567.890123"), legacy.FromString("987.654321")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bridgeSink = MulStrict(a, c)
	}
}
func BenchmarkMulLegacy(b *testing.B) {
	a, c := legacy.FromString("1234567.890123"), legacy.FromString("987.654321")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bridgeSink = a.Mul(c)
	}
}
func BenchmarkMulUpstreamRound(b *testing.B) {
	a, c := jk.FromString("1234567.890123"), jk.FromString("987.654321")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		jkSink = a.MulRound(c, 12, jk.ROUND_NAN)
	}
}

var bridgeSink legacy.Dec128
var jkSink jk.Dec128
