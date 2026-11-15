package optimus

import (
	"database/sql"
	"encoding/json"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec256"
	"github.com/profe-ajedrez/badassitron/dec512"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

var _ sql.Scanner = (*Optimus)(nil)

func must(t *testing.T, s string) Optimus {
	t.Helper()
	d := FromString(s)
	if d.IsNaN() {
		t.Fatalf("parse %q: %v", s, d.ErrorDetails())
	}
	return d
}
func want(t *testing.T, got Optimus, s string) {
	t.Helper()
	if got.IsNaN() || got.String() != s {
		t.Fatalf("got %s (%v), want %s", got.String(), got.ErrorDetails(), s)
	}
}
func TestTiersAndConversions(t *testing.T) {
	nums := []struct{ bits, wantWidth int }{{0, 64}, {63, 64}, {64, 128}, {127, 128}, {128, 256}, {255, 256}, {256, 512}, {511, 512}}
	for _, tc := range nums {
		c := new(big.Int).Lsh(big.NewInt(1), uint(tc.bits))
		if tc.bits == 0 {
			c.SetUint64(0)
		}
		u, e := uint512.FromBigInt(c)
		if e != 0 {
			t.Fatal(e)
		}
		d := New(u, 0, false)
		if d.Width() != tc.wantWidth {
			t.Fatalf("bits=%d tier=%d expected %d", tc.bits, d.Width(), tc.wantWidth)
		}
		if d.coefBig().Cmp(c) != 0 {
			t.Fatal("coefficient round-trip")
		}
	}
	a := must(t, "12345678901234567890.125")
	if a.Width() != 128 {
		t.Fatalf("width %d", a.Width())
	}
	if _, err := a.Int64(); err == nil {
		t.Fatal("expected int64 overflow")
	}
}

func TestPromotionsCancellationAndLimits(t *testing.T) {
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	u, e := uint512.FromBigInt(max)
	if e != 0 {
		t.Fatal(e)
	}
	d := New(u, 1, false)
	if d.Width() != 512 {
		t.Fatal("expected 512")
	}
	r := d.Add(DecodeFromUint64(1, 1))
	if r.IsNaN() {
		t.Fatal("should round rather than overflow", r.ErrorDetails())
	}
	// Adding 0.1 makes the 512-bit coefficient overflow; rounding to
	// scale 0 must remain available.
	expected := new(big.Int).Quo(max, big.NewInt(10))
	expected.Add(expected, big.NewInt(0)) // max mod 10 == 5, tie-to-even
	if expected.Bit(0) == 1 {
		expected.Add(expected, big.NewInt(1))
	}
	if r.coefBig().Cmp(expected) != 0 || r.Precision() != 0 {
		t.Fatalf("got %s exp %d expected %s", r, r.Precision(), expected)
	}
	bigInteger := New(u, 0, false)
	if !bigInteger.Add(One).IsNaN() {
		t.Fatal("integer overflow must be NaN")
	}
	if !bigInteger.Mul(FromInt(2)).IsNaN() {
		t.Fatal("integer overflow must be NaN")
	}
	if !bigInteger.Sub(bigInteger).Equal(Zero) {
		t.Fatal("cancellation")
	}
	if got := bigInteger.Sub(New(u, 0, false)).Width(); got != 64 {
		t.Fatalf("demotion to zero: %d", got)
	}
	a := must(t, "340282366920938463463374607431768211455")
	b := a.Add(One)
	if b.Width() != 256 {
		t.Fatalf("promotion 128->256: %d", b.Width())
	}
	if !b.Sub(One).Equal(a) {
		t.Fatal("round trip")
	}
}

func TestPrecisionAndRounding(t *testing.T) {
	want(t, must(t, "1.245").RoundBank(2), "1.24")
	want(t, must(t, "1.255").RoundBank(2), "1.26")
	want(t, must(t, "-1.255").RoundBank(2), "-1.26")
	want(t, must(t, "-1.245").RoundBank(2), "-1.24")
	want(t, must(t, "1.245").Round(2), "1.25")
	want(t, must(t, "-1.245").Round(2), "-1.25")
	want(t, must(t, "-1.245").RoundDown(2), "-1.25")
	want(t, must(t, "-1.245").RoundUp(2), "-1.24")
	want(t, must(t, "-1.245").Trunc(2), "-1.24")
	z := must(t, "0."+strings.Repeat("0", 154)+"6")
	want(t, z, "0."+strings.Repeat("0", 153)+"1")
	if z.Precision() != 154 {
		t.Fatalf("precision %d", z.Precision())
	}
	tie := must(t, "0."+strings.Repeat("0", 154)+"5")
	if !tie.Equal(Zero) {
		t.Fatalf("half-even 0 tie %s", tie)
	}
	n := must(t, "-0."+strings.Repeat("0", 154)+"6")
	if !n.Equal(z.Neg()) {
		t.Fatalf("negative %s", n)
	}
	want(t, must(t, "1.000").Canonical(), "1")
	if must(t, "1.000").StringFixed() != "1.000" {
		t.Fatal("fixed scale")
	}
	if !must(t, "0.000").Equal(must(t, "-0.00")) {
		t.Fatal("zero sign")
	}
	v, inx := must(t, "1.245").RescaleRoundWithStatus(2, RoundingHalfEven)
	want(t, v, "1.24")
	if !inx {
		t.Fatal("missing inexact")
	}
	if !must(t, "1.2").Rescale(3).Equal(must(t, "1.200")) {
		t.Fatal("rescale up")
	}
}

func TestDivisionExactAndInexact(t *testing.T) {
	old := uint8(defaultPrecision.Load())
	defer SetDefaultPrecision(old)
	SetDefaultPrecision(19)
	want(t, FromInt(1).Div(FromInt(8)), "0.125")
	want(t, FromInt(1).Div(FromInt(3)), "0.3333333333333333333")
	want(t, FromInt(2).Div(FromInt(3)), "0.6666666666666666667")
	want(t, FromInt(-2).Div(FromInt(3)), "-0.6666666666666666667")
	want(t, FromInt(1).Div(FromInt(1<<20)), "0.00000095367431640625") // 20 digits, above default
	v, inx := FromInt(1).DivRoundWithStatus(FromInt(8), 2, RoundingHalfEven)
	want(t, v, "0.12")
	if !inx {
		t.Fatal("inexact")
	}
	exact, inx := FromInt(1).DivRoundWithStatus(FromInt(8), 3, RoundingHalfEven)
	want(t, exact, "0.125")
	if inx {
		t.Fatal("exact")
	}
	if !FromInt(1).Div(Zero).IsNaN() {
		t.Fatal("divide by zero")
	}
	if !FromInt(-1).Sqrt().IsNaN() {
		t.Fatal("negative root")
	}
	want(t, FromInt(4).Sqrt(), "2")
	want(t, FromString("0.0004").Sqrt(), "0.02")
	want(t, FromString("2.25").SqrtRound(0, RoundingHalfEven), "2")
	want(t, FromString("6.25").SqrtRound(0, RoundingHalfEven), "2")
	want(t, FromString("6.25").SqrtRound(0, RoundingHalfAwayFromZero), "3")
	want(t, FromInt(2).SqrtRound(3, RoundingHalfEven), "1.414")
	want(t, FromInt(2).PowInt(10), "1024")
	want(t, FromInt(2).PowInt(-3), "0.125")
	want(t, FromInt(-2).PowInt(3), "-8")
	want(t, FromInt(-2).PowInt(4), "16")
	want(t, FromInt(1).Mod(FromInt(3)), "1")
	q, r := FromInt(-11).QuoRem(FromInt(3))
	want(t, q, "-3")
	want(t, r, "-2")
}

func TestJSONSQLAndInterop(t *testing.T) {
	initial := must(t, "-123.4500")
	payload, err := json.Marshal(initial)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Optimus
	if err = json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(initial) {
		t.Fatalf("JSON roundtrip %s", decoded)
	}
	if err = decoded.Scan([]byte("456.789")); err != nil {
		t.Fatal(err)
	}
	want(t, decoded, "456.789")
	got, err := decoded.Value()
	if err != nil || got != "456.789" {
		t.Fatalf("SQL value %v, %v", got, err)
	}
	a := dec128.FromString("123.456")
	x := FromDec128(a)
	back, err := x.Dec128Exact()
	if err != nil || !back.Equal(a) {
		t.Fatalf("128 %v %v", back, err)
	}
	b := dec256.FromString("123456789012345678901234567890123456789.25")
	y := FromDec256(b)
	if y.Width() != 256 {
		t.Fatalf("256 tier %d", y.Width())
	}
	back256, err := y.Dec256Exact()
	if err != nil || !back256.Equal(b) {
		t.Fatalf("256 %v %v", back256, err)
	}
	c := dec512.FromString("123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890.25")
	z := FromDec512(c)
	if z.Width() != 512 {
		t.Fatalf("512 tier %d", z.Width())
	}
	back512, err := z.Dec512Exact()
	if err != nil || !back512.Equal(c) {
		t.Fatalf("512 %v %v", back512, err)
	}
	if _, err = z.Dec128Exact(); err == nil {
		t.Fatal("narrowing must fail")
	}
	if _, err = z.Dec256Exact(); err == nil {
		t.Fatal("narrowing must fail")
	}
}

func ratFrom(d Optimus) *big.Rat {
	c := d.coefBig()
	if d.neg {
		c.Neg(c)
	}
	return new(big.Rat).SetFrac(c, pow10(int(d.exp)))
}
func oracleAt(rat *big.Rat, p int) (*big.Int, bool) {
	num := new(big.Int).Abs(rat.Num())
	num.Mul(num, pow10(p))
	den := rat.Denom()
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(num, den, r)
	cmp := new(big.Int).Lsh(r, 1).Cmp(den)
	if cmp > 0 || cmp == 0 && q.Bit(0) == 1 {
		q.Add(q, big.NewInt(1))
	}
	if rat.Sign() < 0 {
		q.Neg(q)
	}
	return q, r.Sign() != 0
}
func oracleAuto(rat *big.Rat) (*big.Rat, bool) {
	for p := int(MaxPrecision); p >= 0; p-- {
		q, _ := oracleAt(rat, p)
		if new(big.Int).Abs(q).BitLen() <= 512 {
			return new(big.Rat).SetFrac(q, pow10(p)), true
		}
	}
	return nil, false
}

// Compare the output of all ordinary arithmetic operations with a separate
// rational reference; this catches premature rounding, scale alignment errors
// and incorrect tier promotions.
func TestRandomArithmeticAgainstRational(t *testing.T) {
	rng := rand.New(rand.NewSource(0xBADA5517))
	for i := 0; i < 4000; i++ {
		bitsA := 1 + rng.Intn(512)
		bitsB := 1 + rng.Intn(512)
		aCoef := new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(bitsA)))
		bCoef := new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(bitsB)))
		ua, ea := uint512.FromBigInt(aCoef)
		ub, eb := uint512.FromBigInt(bCoef)
		if ea != 0 || eb != 0 {
			t.Fatal(ea, eb)
		}
		a := New(ua, uint8(rng.Intn(155)), rng.Intn(2) == 1)
		b := New(ub, uint8(rng.Intn(155)), rng.Intn(2) == 1)
		av, bv := ratFrom(a), ratFrom(b)
		cases := []struct {
			name string
			got  Optimus
			rat  *big.Rat
		}{
			{"add", a.Add(b), new(big.Rat).Add(av, bv)},
			{"sub", a.Sub(b), new(big.Rat).Sub(av, bv)},
			{"mul", a.Mul(b), new(big.Rat).Mul(av, bv)},
		}
		for _, tc := range cases {
			expected, ok := oracleAuto(tc.rat)
			if !ok {
				if !tc.got.IsNaN() {
					t.Fatalf("%d %s expected overflow: a=%s b=%s got=%s", i, tc.name, a, b, tc.got)
				}
				continue
			}
			if tc.got.IsNaN() {
				t.Fatalf("%d %s unexpected NaN: a=%s b=%s %v", i, tc.name, a, b, tc.got.ErrorDetails())
			}
			if ratFrom(tc.got).Cmp(expected) != 0 {
				t.Fatalf("%d %s a=%s b=%s got=%s expected=%s", i, tc.name, a, b, tc.got, expected.FloatString(154))
			}
			if tc.got.Width() > 512 {
				t.Fatalf("invalid width")
			}
		}
	}
}

func TestRandomDivRoundAgainstRational(t *testing.T) {
	rng := rand.New(rand.NewSource(4172))
	for i := 0; i < 3000; i++ {
		a := FromInt64(rng.Int63n(math.MaxInt64) - rng.Int63n(math.MaxInt64))
		b := FromInt64(rng.Int63n(math.MaxInt64) + 1)
		p := uint8(rng.Intn(35))
		expected, _ := oracleAt(new(big.Rat).Quo(ratFrom(a), ratFrom(b)), int(p))
		got, inx := a.DivRoundWithStatus(b, p, RoundingHalfEven)
		if got.IsNaN() || got.coefBig().Cmp(new(big.Int).Abs(expected)) != 0 || got.IsNegative() != (expected.Sign() < 0) {
			t.Fatalf("i=%d a=%s b=%s p=%d got=%s expected=%s", i, a, b, p, got, expected)
		}
		exact := new(big.Rat).Quo(ratFrom(a), ratFrom(b))
		_, rem := oracleAt(exact, int(p))
		if inx != rem {
			t.Fatalf("inexact status wrong")
		}
	}
}

func BenchmarkAddSmall(b *testing.B) {
	a, c := FromInt(1200), FromInt(3400)
	var result Optimus
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = a.Add(c)
	}
	sink = result
}
func BenchmarkMulSmall(b *testing.B) {
	a, c := FromInt(1200), FromInt(3400)
	var result Optimus
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = a.Mul(c)
	}
	sink = result
}
func BenchmarkAddWide(b *testing.B) {
	a := FromString(strings.Repeat("9", 140))
	c := FromInt(17)
	var result Optimus
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = a.Add(c)
	}
	sink = result
}
func BenchmarkMulWide(b *testing.B) {
	a := FromString(strings.Repeat("9", 70))
	c := a
	var result Optimus
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = a.Mul(c)
	}
	sink = result
}
func BenchmarkDivSmall(b *testing.B) {
	a, c := FromInt(1200), FromInt(3)
	var result Optimus
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = a.Div(c)
	}
	sink = result
}
func BenchmarkDivRepeating(b *testing.B) {
	a, c := FromInt(1), FromInt(3)
	var result Optimus
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = a.Div(c)
	}
	sink = result
}

var sink Optimus

func TestFinancialGoldenAndExtremeScale(t *testing.T) {
	old := uint8(defaultPrecision.Load())
	defer SetDefaultPrecision(old)
	SetDefaultPrecision(19)
	// The SUNAT grabadora example: no premature rounding between operations.
	price := must(t, "98")
	unit := price.Div(must(t, "1.18"))
	net := unit.Mul(FromInt(2000)).Mul(must(t, "0.90"))
	want(t, net.RoundBank(2), "149491.53")
	// Underflow at 155 places should round to the nearest 154-scale value.
	tiny := must(t, "0."+strings.Repeat("0", 153)+"1")
	if !tiny.DivInt(10).Equal(Zero) {
		t.Fatal("expected nearest zero beyond 154")
	}
	// Exact division may use more decimal places than the default.
	value := FromInt(1).Div(FromInt(1 << 30))
	want(t, value, "0.000000000931322574615478515625")
	if value.Precision() != 30 {
		t.Fatalf("terminating division scale=%d", value.Precision())
	}
	if !FromInt(1).DivRound(FromInt(0), 2, RoundingHalfEven).IsNaN() {
		t.Fatal("missing zero division error")
	}
	if !FromInt(1).Rescale(255).IsNaN() {
		t.Fatal("invalid rescale")
	}
}

func TestRandomSqrtRoundReference(t *testing.T) {
	rng := rand.New(rand.NewSource(235))
	for i := 0; i < 1200; i++ {
		c := new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(rng.Intn(511)+1)))
		u, e := uint512.FromBigInt(c)
		if e != 0 {
			t.Fatal(e)
		}
		a := New(u, uint8(rng.Intn(155)), false)
		p := uint8(rng.Intn(35))
		got, inx := a.SqrtRoundWithStatus(p, RoundingHalfEven)
		if got.IsNaN() {
			continue
		} // integer coefficient may overflow at requested scale
		// Floor of sqrt(c * 10^(2p) / 10^scale), then exact midpoint comparison.
		num := new(big.Int).Set(c)
		den := big.NewInt(1)
		shift := 2*int(p) - int(a.exp)
		if shift >= 0 {
			num.Mul(num, pow10(shift))
		} else {
			den = pow10(-shift)
		}
		k := new(big.Int).Sqrt(new(big.Int).Quo(num, den))
		ks := new(big.Int).Mul(new(big.Int).Set(k), k)
		exact := new(big.Int).Mul(ks, den).Cmp(num) == 0
		if inx == exact {
			t.Fatalf("incorrect sqrt inexact: iteration %d", i)
		}
		midpoint := new(big.Int).Mul(ks, big.NewInt(4))
		midpoint.Add(midpoint, new(big.Int).Mul(k, big.NewInt(4)))
		midpoint.Add(midpoint, big.NewInt(1))
		midpoint.Mul(midpoint, den)
		cmp := new(big.Int).Mul(num, big.NewInt(4)).Cmp(midpoint)
		if cmp > 0 || cmp == 0 && k.Bit(0) == 1 {
			k.Add(k, big.NewInt(1))
		}
		if got.coefBig().Cmp(k) != 0 {
			t.Fatalf("sqrt round mismatch i=%d", i)
		}
	}
}

func FuzzOptimusArithmetic(f *testing.F) {
	f.Add("1.245", "-1.005")
	f.Add("0.00000000000000000001", "12345678901234567890")
	f.Add(strings.Repeat("9", 150), "0.5")
	f.Fuzz(func(t *testing.T, x, y string) {
		if len(x) > 300 || len(y) > 300 {
			return
		}
		a, b := FromString(x), FromString(y)
		if a.IsNaN() || b.IsNaN() {
			return
		}
		for _, item := range []Optimus{a.Add(b), a.Sub(b), a.Mul(b)} {
			if item.IsNaN() {
				continue
			}
			back := FromString(item.StringFixed())
			if back.IsNaN() || !back.Equal(item) {
				t.Fatalf("roundtrip failed: a=%q b=%q got=%s", x, y, item)
			}
		}
	})
}

func BenchmarkDec128AddBaseline(b *testing.B) {
	a, c := dec128.FromInt(1200), dec128.FromInt(3400)
	var result dec128.Dec128
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = a.Add(c)
	}
	benchmarkSink128 = result
}
func BenchmarkDec128MulBaseline(b *testing.B) {
	a, c := dec128.FromInt(1200), dec128.FromInt(3400)
	var result dec128.Dec128
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = a.Mul(c)
	}
	benchmarkSink128 = result
}

var benchmarkSink128 dec128.Dec128

func TestLargePowersHaveProvenBounds(t *testing.T) {
	if !FromInt(2).PowInt(-5000).Equal(Zero) {
		t.Fatal("2^-5000 must round to zero")
	}
	if !FromString("0.5").PowInt(5000).Equal(Zero) {
		t.Fatal("0.5^5000 must round to zero")
	}
	if !FromInt(2).PowInt(5000).IsNaN() {
		t.Fatal("2^5000 integer overflow")
	}
	if !FromString("0.5").PowInt(-5000).IsNaN() {
		t.Fatal("0.5^-5000 integer overflow")
	}
	base := FromString("1.000000000000000000000000000001")
	got := base.PowInt(5000)
	if got.IsNaN() {
		t.Fatal("near-one gigantic power should be representable", got.ErrorDetails())
	}
	exact := new(big.Rat).SetFrac(
		new(big.Int).Exp(base.coefBig(), big.NewInt(5000), nil),
		new(big.Int).Exp(pow10(int(base.exp)), big.NewInt(5000), nil),
	)
	wantRat, ok := oracleAuto(exact)
	if !ok || ratFrom(got).Cmp(wantRat) != 0 {
		t.Fatal("incorrect bounded power", got)
	}
	negative := base.Neg().PowInt(5001)
	if negative.IsNaN() || !negative.IsNegative() {
		t.Fatal("negative odd power must remain negative")
	}
	inverse := base.PowInt(-5000)
	if inverse.IsNaN() || !inverse.IsPositive() {
		t.Fatal("reciprocal near-one power must remain representable")
	}
}

func TestLargePowerDifferentialAgainstExactRationals(t *testing.T) {
	cases := []struct {
		base string
		exp  int
	}{
		{"1.00001", 4097},
		{"0.99999", 4097},
		{"1.00000000000000000001", 5000},
		{"0.99999999999999999999", 5000},
		{"-1.00000000000000000001", 5001},
		{"-0.99999999999999999999", 5001},
		{"0.95", 5000},
		{"0.95", -5000},
		{"1.000000000000000001", -5001},
		{"0.999999999999999999", -5001},
	}
	for _, tc := range cases {
		base := must(t, tc.base)
		got := base.PowInt(tc.exp)
		mag := tc.exp
		if mag < 0 {
			mag = -mag
		}
		c := new(big.Int).Exp(base.coefBig(), big.NewInt(int64(mag)), nil)
		s := new(big.Int).Exp(pow10(int(base.exp)), big.NewInt(int64(mag)), nil)
		if tc.exp < 0 {
			c, s = s, c
		}
		if base.neg && mag%2 == 1 {
			c.Neg(c)
		}
		wantRat, representable := oracleAuto(new(big.Rat).SetFrac(c, s))
		if !representable {
			if !got.IsNaN() {
				t.Fatalf("%s ^ %d: expected overflow, got %s", tc.base, tc.exp, got)
			}
			continue
		}
		if got.IsNaN() || ratFrom(got).Cmp(wantRat) != 0 {
			t.Fatalf("%s ^ %d: got %s (%v), want %s", tc.base, tc.exp, got, got.ErrorDetails(), wantRat.FloatString(154))
		}
	}
}

func TestLargePowerExponentExtremes(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	if !FromInt(2).PowInt(minInt).Equal(Zero) {
		t.Fatal("inverse huge exponent must underflow")
	}
	if !FromInt(2).PowInt(maxInt).IsNaN() {
		t.Fatal("huge integer power must overflow")
	}
	if !FromString("0.5").PowInt(maxInt).Equal(Zero) {
		t.Fatal("small huge power must underflow")
	}
	if !FromString("0.5").PowInt(minInt).IsNaN() {
		t.Fatal("inverse huge small power must overflow")
	}
	if !FromInt(-1).PowInt(minInt).Equal(One) {
		t.Fatal("(-1)^minInt")
	}
	if !FromInt(-1).PowInt(maxInt).Equal(NegativeOne) {
		t.Fatal("(-1)^maxInt")
	}
}
