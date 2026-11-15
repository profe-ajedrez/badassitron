package dec512

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"testing"
)

func must(t *testing.T, s string) Dec512 {
	t.Helper()
	d := FromString(s)
	if d.IsNaN() {
		t.Fatalf("parse %q: %v", s, d.ErrorDetails())
	}
	return d
}
func rat(d Dec512) *big.Rat {
	r, ok := new(big.Rat).SetString(d.String())
	if !ok {
		panic("invalid decimal: " + d.String())
	}
	return r
}
func scaledRound(x *big.Rat, scale int, mode RoundingMode) (*big.Int, bool) {
	n := new(big.Int).Mul(x.Num(), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil))
	den := x.Denom()
	neg := n.Sign() < 0
	n.Abs(n)
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(n, den, r)
	inexact := r.Sign() != 0
	if inexact {
		twice := new(big.Int).Lsh(new(big.Int).Set(r), 1)
		cmp := twice.Cmp(den)
		up := false
		switch mode {
		case RoundingTowardZero:
		case RoundingAwayFromZero:
			up = true
		case RoundingFloor:
			up = neg
		case RoundingCeil:
			up = !neg
		case RoundingHalfTowardZero:
			up = cmp > 0
		case RoundingHalfAwayFromZero:
			up = cmp >= 0
		case RoundingHalfEven:
			up = cmp > 0 || cmp == 0 && q.Bit(0) == 1
		}
		if up {
			q.Add(q, big.NewInt(1))
		}
	}
	if neg {
		q.Neg(q)
	}
	return q, inexact
}
func valueAtScale(d Dec512, scale int) *big.Int {
	q := new(big.Int).Set(d.Coefficient().BigInt())
	if int(d.Precision()) <= scale {
		q.Mul(q, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale-int(d.Precision()))), nil))
	} else {
		panic("invalid scale")
	}
	if d.IsNegative() {
		q.Neg(q)
	}
	return q
}
func TestBasicAPI(t *testing.T) {
	if got := FromString("").String(); got != "0" {
		t.Fatal(got)
	}
	if got := must(t, "-12.3400").String(); got != "-12.34" {
		t.Fatal(got)
	}
	if got := must(t, "-12.3400").StringFixed(); got != "-12.3400" {
		t.Fatal(got)
	}
	if got := must(t, "-0.000").StringFixed(); got != "0.000" {
		t.Fatal(got)
	}
	if got := must(t, "0.000").Canonical().StringFixed(); got != "0" {
		t.Fatal(got)
	}
	for _, invalid := range []string{".", "-", "+", "1.", ".2", "1..2", "1e3", "1 2"} {
		if !FromString(invalid).IsNaN() {
			t.Fatalf("accepted invalid %q", invalid)
		}
	}
	if !FromString("0." + strings.Repeat("0", int(MaxPrecision)+1)).IsNaN() {
		t.Fatal("accepted excessive scale")
	}
	tiny := must(t, "0."+strings.Repeat("0", int(MaxPrecision)-1)+"1")
	if tiny.Precision() != MaxPrecision || tiny.Coefficient().BigInt().Cmp(big.NewInt(1)) != 0 {
		t.Fatal(tiny)
	}
	maximum := uint512.Max
	if got := New(maximum, 0, false).Coefficient(); !got.Equal(maximum) {
		t.Fatal("max coefficient changed")
	}
	over := new(big.Int).Add(maximum.BigInt(), big.NewInt(1))
	if !FromString(over.String()).IsNaN() {
		t.Fatal("missing coefficient overflow")
	}
	if !New(uint512.One, MaxPrecision+1, false).IsNaN() {
		t.Fatal("max scale")
	}
	if got := FromInt64(math.MinInt64).String(); got != "-9223372036854775808" {
		t.Fatal(got)
	}
	if got, e := FromInt64(math.MinInt64).Int64(); e != nil || got != math.MinInt64 {
		t.Fatal(got, e)
	}
	if got := must(t, "1.23").Rescale(1).StringFixed(); got != "1.2" {
		t.Fatal(got)
	}
	if got := must(t, "1.23").Rescale(4).StringFixed(); got != "1.2300" {
		t.Fatal(got)
	}
	if got := must(t, "2.5").RoundBank(0).String(); got != "2" {
		t.Fatal(got)
	}
	if got := must(t, "3.5").RoundBank(0).String(); got != "4" {
		t.Fatal(got)
	}
	if got := must(t, "-1.21").RoundDown(1).String(); got != "-1.3" {
		t.Fatal(got)
	}
	if got := must(t, "-1.21").RoundUp(1).String(); got != "-1.2" {
		t.Fatal(got)
	}
	if got := must(t, "12.0").Div(must(t, "0")).IsNaN(); !got {
		t.Fatal("division by zero")
	}
	if got := must(t, "1").Sqrt().String(); got != "1" {
		t.Fatal(got)
	}
	if got := must(t, "-4").SqrtRound(2, RoundingHalfEven); !got.IsNaN() {
		t.Fatal("negative sqrt")
	}
	if got := must(t, "2").PowInt(10).String(); got != "1024" {
		t.Fatal(got)
	}
	if got := must(t, "2").PowInt(-3).String(); got != "0.125" {
		t.Fatal(got)
	}
	if got := must(t, "2").Add(must(t, "3")).String(); got != "5" {
		t.Fatal(got)
	}
	if got := must(t, "2").Sub(must(t, "3")).String(); got != "-1" {
		t.Fatal(got)
	}
	if got := must(t, "2").Mul(must(t, "3")).String(); got != "6" {
		t.Fatal(got)
	}
	if got := must(t, "2").DivRound(must(t, "3"), 2, RoundingHalfEven).String(); got != "0.67" {
		t.Fatal(got)
	}
	if got := must(t, "2.25").SqrtRound(0, RoundingHalfEven).String(); got != "2" {
		t.Fatal(got)
	}
	if got := must(t, "2.25").SqrtRound(0, RoundingHalfTowardZero).String(); got != "1" {
		t.Fatal(got)
	}
	if got := must(t, "2.25").SqrtRound(0, RoundingHalfAwayFromZero).String(); got != "2" {
		t.Fatal(got)
	}
	if got, ix := must(t, "2.25").SqrtRoundWithStatus(1, RoundingHalfEven); ix || got.String() != "1.5" {
		t.Fatal(got, ix)
	}
	if got, ix := must(t, "1.001").RescaleRoundWithStatus(2, RoundingHalfEven); !ix || got.String() != "1" {
		t.Fatal(got, ix)
	}
	if got := must(t, "1.20").Equal(must(t, "1.2")); !got {
		t.Fatal("scale equality")
	}
	if got := must(t, "-0.00").Equal(Zero); !got {
		t.Fatal("negative zero equality")
	}
}

func TestSerializationAndSQL(t *testing.T) {
	original := must(t, "-12345.678900")
	text, err := original.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	var parsed Dec512
	if err := parsed.UnmarshalText(text); err != nil || !original.Equal(parsed) {
		t.Fatal(parsed, err)
	}
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &parsed); err != nil || !original.Equal(parsed) {
		t.Fatal(parsed, err)
	}
	if err := json.Unmarshal([]byte("null"), &parsed); err != nil || !parsed.IsZero() {
		t.Fatal(parsed, err)
	}
	if err := parsed.Scan("123.45"); err != nil || parsed.String() != "123.45" {
		t.Fatal(parsed, err)
	}
	if err := parsed.Scan([]byte("456.78")); err != nil || parsed.String() != "456.78" {
		t.Fatal(parsed, err)
	}
	if err := parsed.Scan(int64(math.MinInt64)); err != nil || parsed.String() != "-9223372036854775808" {
		t.Fatal(parsed, err)
	}
	if err := parsed.Scan(nil); err != nil || !parsed.IsZero() {
		t.Fatal(parsed, err)
	}
	if _, ok := any(&parsed).(sql.Scanner); !ok {
		t.Fatal("missing sql.Scanner")
	}
	if got, err := original.Value(); err != nil || got != "-12345.6789" {
		t.Fatal(got, err)
	}
}

func TestWideExactAndOverflow(t *testing.T) {
	maxInt := New(uint512.Max, 0, false)
	if got := maxInt.Add(One); !got.IsNaN() {
		t.Fatal("max+1 must overflow")
	}
	if got := maxInt.Sub(One); got.IsNaN() {
		t.Fatal("max-1 must succeed")
	}
	if got := maxInt.Mul(One); got.IsNaN() || !got.Equal(maxInt) {
		t.Fatal("max*1 must succeed")
	}
	if got := must(t, "2").Mul(maxInt); !got.IsNaN() {
		t.Fatal("max*2 must overflow")
	}
	// Exact cancellation across different scales must not overflow spuriously.
	n := New(uint512.Max, 1, false)
	if got := n.Sub(n); !got.IsZero() {
		t.Fatal(got)
	}
	// Divisor smaller than the dividend by hundreds of decimal places: the
	// quotient cannot fit, yet Mod remains exact.
	divisor := must(t, "0."+strings.Repeat("0", int(MaxPrecision)-1)+"3")
	if q, r := maxInt.QuoRem(divisor); !q.IsNaN() || !r.IsNaN() {
		t.Fatalf("QuoRem overflow %v %v", q, r)
	}
	if r := maxInt.Mod(divisor); r.IsNaN() {
		t.Fatalf("Mod may not overflow because quotient does: %v", r.ErrorDetails())
	}
	if q, r := must(t, "1").QuoRem(must(t, "3")); q.String() != "0" || r.String() != "1" {
		t.Fatal(q, r)
	}
}

func TestRandomReferenceArithmetic(t *testing.T) {
	rand := rand.New(rand.NewSource(20261112 + 512))
	for i := 0; i < 3000; i++ {
		ai := rand.Int63n(2_000_000_000) - 1_000_000_000
		bi := rand.Int63n(2_000_000_000) - 1_000_000_000
		as, bs := rand.Intn(21), rand.Intn(21)
		a := New(uint512.FromUint64(abs64(ai)), uint8(as), ai < 0)
		b := New(uint512.FromUint64(abs64(bi)), uint8(bs), bi < 0)
		ar, br := rat(a), rat(b)
		for _, test := range []struct {
			name string
			got  Dec512
			want *big.Rat
		}{
			{"add", a.Add(b), new(big.Rat).Add(ar, br)},
			{"sub", a.Sub(b), new(big.Rat).Sub(ar, br)},
			{"mul", a.Mul(b), new(big.Rat).Mul(ar, br)},
		} {
			if test.got.IsNaN() || rat(test.got).Cmp(test.want) != 0 {
				t.Fatalf("%s (%s %s): %s != %s", test.name, a, b, test.got, test.want.RatString())
			}
		}
		for _, mode := range []RoundingMode{RoundingTowardZero, RoundingAwayFromZero, RoundingFloor, RoundingCeil, RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven} {
			scale := rand.Intn(21)
			if bi != 0 {
				got, inexact := a.DivRoundWithStatus(b, uint8(scale), mode)
				want, wantInexact := scaledRound(new(big.Rat).Quo(ar, br), scale, mode)
				if got.IsNaN() || valueAtScale(got, scale).Cmp(want) != 0 || inexact != wantInexact {
					t.Fatalf("DivRound(%s/%s) scale %d mode %d: got %s %t want %s %t", a, b, scale, mode, got, inexact, want, wantInexact)
				}
			}
			got, ix := a.MulRoundWithStatus(b, uint8(scale), mode)
			want, wantIx := scaledRound(new(big.Rat).Mul(ar, br), scale, mode)
			if got.IsNaN() || valueAtScale(got, scale).Cmp(want) != 0 || ix != wantIx {
				t.Fatalf("MulRound(%s*%s) scale %d mode %d: got %s %t want %s %t", a, b, scale, mode, got, ix, want, wantIx)
			}
		}
		if bi != 0 {
			q, r := a.QuoRem(b)
			if q.IsNaN() || r.IsNaN() {
				t.Fatalf("unexpected QuoRem NaN %s %s", a, b)
			}
			if rat(a).Cmp(new(big.Rat).Add(new(big.Rat).Mul(rat(q), br), rat(r))) != 0 {
				t.Fatalf("bad quo/rem %s=%s*%s+%s", a, q, b, r)
			}
			if !a.Mod(b).Equal(r) {
				t.Fatalf("Mod mismatch %v %v", a.Mod(b), r)
			}
			if !r.IsZero() && r.Sign() != a.Sign() {
				t.Fatalf("sign remainder")
			}
			if rat(r.Abs()).Cmp(new(big.Rat).Abs(br)) >= 0 {
				t.Fatalf("remainder magnitude")
			}
		}
	}
}
func abs64(i int64) uint64 {
	if i < 0 {
		return uint64(-(i + 1)) + 1
	}
	return uint64(i)
}

func TestRandomSqrtReference(t *testing.T) {
	rnd := rand.New(rand.NewSource(61236 + 512))
	for i := 0; i < 1800; i++ {
		v := rnd.Uint64() >> 1
		scale := rnd.Intn(40)
		value := New(uint512.FromUint64(v), uint8(scale), false)
		precision := rnd.Intn(40)
		for _, mode := range []RoundingMode{RoundingHalfEven, RoundingTowardZero, RoundingCeil} {
			got, inexact := value.SqrtRoundWithStatus(uint8(precision), mode)
			if got.IsNaN() {
				t.Fatalf("unexpected sqrt NaN %s/%d", value, precision)
			}
			n := new(big.Int).SetUint64(v)
			den := big.NewInt(1)
			diff := 2*precision - scale
			if diff >= 0 {
				n.Mul(n, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(diff)), nil))
			} else {
				den.Exp(big.NewInt(10), big.NewInt(int64(-diff)), nil)
			}
			k := new(big.Int).Sqrt(new(big.Int).Quo(n, den))
			sq := new(big.Int).Mul(k, k)
			exact := new(big.Int).Mul(sq, den).Cmp(n) == 0
			if inexact == exact {
				t.Fatalf("sqrt inexact mismatch %v %v", inexact, exact)
			}
			midpoint := new(big.Int).Mul(sq, big.NewInt(4))
			midpoint.Add(midpoint, new(big.Int).Mul(k, big.NewInt(4)))
			midpoint.Add(midpoint, big.NewInt(1))
			midpoint.Mul(midpoint, den)
			if !exact {
				cmp := new(big.Int).Mul(n, big.NewInt(4)).Cmp(midpoint)
				if mode == RoundingCeil || mode == RoundingHalfEven && (cmp > 0 || cmp == 0 && k.Bit(0) == 1) {
					k.Add(k, big.NewInt(1))
				}
			}
			if valueAtScale(got, precision).Cmp(k) != 0 {
				t.Fatalf("sqrt %s precision %d mode %d got=%s want=%s", value, precision, mode, got, k)
			}
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	a := FromString("123456789.123456789")
	c := FromString("987654321.987654321")
	var sink Dec512
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = a.Add(c)
	}
	_ = sink
}
func BenchmarkMul(b *testing.B) {
	a := FromString("123456789.123456789")
	c := FromString("987654321.987654321")
	var sink Dec512
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = a.Mul(c)
	}
	_ = sink
}
func BenchmarkDivRound(b *testing.B) {
	a := FromString("123456789.123456789")
	c := FromString("3.141592653589793")
	var sink Dec512
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = a.DivRound(c, 19, RoundingHalfEven)
	}
	_ = sink
}
func TestExampleSUNAT(t *testing.T) {
	// 98 / 1.18 * 2000 * 0.90, rounded once to cents.
	gross := must(t, "98").DivRound(must(t, "1.18"), 28, RoundingHalfEven)
	if gross.IsNaN() {
		t.Fatal("overflow")
	}
	net := gross.Mul(must(t, "2000")).Mul(must(t, "0.9"))
	if net.IsNaN() {
		t.Fatalf("intermediate %v", net.ErrorDetails())
	}
	actual := net.RoundBank(2)
	if got := actual.String(); got != "149491.53" {
		t.Fatalf("SUNAT expected 149491.53, got %s (base %v)", got, net)
	}
	_ = fmt.Sprintf("%v", actual)
}
