package optimus

import (
	"encoding/json"
	"math/big"
	"math/bits"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec256/uint256"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

// Exact copy of the original Mul routing, apart from calling mulLegacy for
// the old math/big final branch. This reference is intentionally test-only.
func fixed02BOriginalMul(d, b Optimus) Optimus {
	if d.IsNaN() {
		return d
	}
	if b.IsNaN() {
		return b
	}
	if d.IsZero() || b.IsZero() {
		return Zero
	}
	if d.wide == nil && b.wide == nil && d.small.Hi == 0 && b.small.Hi == 0 && int(d.exp)+int(b.exp) <= int(MaxPrecision) {
		hi, lo := bits.Mul64(d.small.Lo, b.small.Lo)
		return NewFromUint128(uint128.Uint128{Lo: lo, Hi: hi}, d.exp+b.exp, d.neg != b.neg)
	}
	if d.wide == nil && b.wide == nil && int(d.exp)+int(b.exp) <= int(MaxPrecision) {
		lo, hi := d.small.MulCarry(b.small)
		if hi.IsZero() {
			return NewFromUint128(lo, d.exp+b.exp, d.neg != b.neg)
		}
		return NewFromUint256(uint256.Uint256{Lo: lo, Hi: hi}, d.exp+b.exp, d.neg != b.neg)
	}
	if int(d.exp)+int(b.exp) <= int(MaxPrecision) {
		lo, hi := d.Coefficient().MulCarry(b.Coefficient())
		if hi.IsZero() {
			return New(lo, d.exp+b.exp, d.neg != b.neg)
		}
		if d.exp+b.exp == 0 {
			return NaN(errors.Overflow)
		}
	}
	return d.mulLegacy(b)
}

func fixed02BEqualRepresentation(a, b Optimus) bool {
	return a.err == b.err && a.exp == b.exp && a.neg == b.neg && a.Coefficient() == b.Coefficient()
}

func fixed02BCheck(t *testing.T, a, b Optimus) {
	t.Helper()
	got, want := a.Mul(b), fixed02BOriginalMul(a, b)
	if !fixed02BEqualRepresentation(got, want) {
		t.Fatalf("Mul mismatch: a=%s [scale=%d] b=%s [scale=%d] got=%s [scale=%d err=%v] want=%s [scale=%d err=%v]", a.StringFixed(), a.exp, b.StringFixed(), b.exp, got.StringFixed(), got.exp, got.err, want.StringFixed(), want.exp, want.err)
	}
	if got.StringFixed() != want.StringFixed() || got.String() != want.String() || got.Width() != want.Width() {
		t.Fatalf("observable representation mismatch: got=%s, want=%s", got.StringFixed(), want.StringFixed())
	}
	// An independent big.Int oracle checks the rounded mathematical value at
	// the result's reported precision (without using decimalwide.Quantize).
	if !got.IsNaN() {
		if a.IsNaN() || b.IsNaN() {
			t.Fatal("finite product with invalid operand")
		}
		numer := new(big.Int).Mul(a.coefBig(), b.coefBig())
		shift := int(a.exp) + int(b.exp) - int(got.exp)
		if shift >= 0 {
			den := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(shift)), nil)
			quo, rem := new(big.Int).QuoRem(numer, den, new(big.Int))
			twice := new(big.Int).Lsh(rem, 1)
			if c := twice.Cmp(den); c > 0 || (c == 0 && quo.Bit(0) != 0) {
				quo.Add(quo, big.NewInt(1))
			}
			if quo.Cmp(got.coefBig()) != 0 {
				t.Fatalf("big.Int half-even oracle failed: got=%s", got.StringFixed())
			}
		} else {
			factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-shift)), nil)
			numer.Mul(numer, factor)
			if numer.Cmp(got.coefBig()) != 0 {
				t.Fatal("big.Int scale-up oracle failed")
			}
		}
	}
	if ja, e1 := json.Marshal(got); e1 == nil {
		jb, e2 := json.Marshal(want)
		if e2 != nil || string(ja) != string(jb) {
			t.Fatal("JSON compatibility mismatch")
		}
	}
}

func fixed02BRandom(r *rand.Rand) Optimus {
	if r.Intn(70) == 0 {
		return Zero
	}
	if r.Intn(120) == 0 {
		return NaN(errors.InvalidFormat)
	}
	var v [16]uint64
	sizes := []int{1, 2, 3, 4, 5, 6, 7, 8}
	n := sizes[r.Intn(len(sizes))]
	for i := 0; i < n; i++ {
		v[i] = r.Uint64()
	}
	v[n-1] |= 1
	// Give some numbers trailing decimal zeros without losing width.
	if r.Intn(8) == 0 {
		v[0] = v[0] / 10 * 10
	}
	scales := []uint8{0, 1, 2, 8, 19, 38, 76, 77, 110, 127, 153, 154}
	return fixed01Pack(&v, n, scales[r.Intn(len(scales))], r.Intn(2) == 0)
}

func TestFixed02BMulDifferential(t *testing.T) {
	r := rand.New(rand.NewSource(2026111402))
	for i := 0; i < 20000; i++ {
		a, b := fixed02BRandom(r), fixed02BRandom(r)
		fixed02BCheck(t, a, b)
	}
	values := []Optimus{
		Zero, FromString("-0.000"), One, NegativeOne,
		FromString("0.5"), FromString("1.5"), FromString("2.5"), FromString("-1.5"),
		FromString("0." + strings.Repeat("0", 153) + "1"),
		FromString("0." + strings.Repeat("9", 154)),
		FromString("0." + strings.Repeat("8", 154)),
		FromString(strings.Repeat("9", 154)),
		FromString(strings.Repeat("9", 153) + "0"),
		FromString(strings.Repeat("5", 150) + ".0000"),
		FromString("0." + strings.Repeat("0", 130) + strings.Repeat("1", 24)),
		NaN(errors.Overflow), NaN(errors.InvalidFormat),
	}
	for _, v := range values {
		for _, w := range values {
			fixed02BCheck(t, v, w)
		}
	}
	for _, tc := range phaseB2MulCases() {
		fixed02BCheck(t, tc.a, tc.b)
	}

	// Maximal 512-bit operands at scales 0, 75, 77 and 154,
	// and near-rounding boundaries around powers of ten.
	maxWord := uint128.Uint128{Lo: ^uint64(0), Hi: ^uint64(0)}
	maxCoef := uint512.Uint512{Lo: uint256.Uint256{Lo: maxWord, Hi: maxWord}, Hi: uint256.Uint256{Lo: maxWord, Hi: maxWord}}
	for _, s := range []uint8{0, 1, 75, 77, 154} {
		for _, q := range []uint8{0, 1, 75, 77, 154} {
			fixed02BCheck(t, New(maxCoef, s, true), New(maxCoef, q, false))
		}
	}
}

func TestFixed02BMulResultIndependence(t *testing.T) {
	a := FromString("0." + strings.Repeat("9", 150))
	b := FromString("0." + strings.Repeat("8", 150))
	x, y := a.Mul(b), a.Mul(b)
	if x.IsNaN() || y.IsNaN() || x.wide == nil || y.wide == nil {
		t.Skip("test case did not produce two wide results")
	}
	if x.wide == y.wide || x.wide == a.wide || x.wide == b.wide {
		t.Fatal("aliased mutable coefficient")
	}
	before := y.Coefficient()
	x.wide.Lo.Lo.Lo++ // Mutate only this private test result; no live data races.
	if y.Coefficient() != before {
		t.Fatal("independent result was modified")
	}
}

func TestFixed02BMulConcurrent(t *testing.T) {
	a := FromString("0." + strings.Repeat("9", 152))
	b := FromString("0." + strings.Repeat("8", 152))
	want := fixed02BOriginalMul(a, b)
	n := min(runtime.GOMAXPROCS(0), 8)
	var wg sync.WaitGroup
	for k := 0; k < n; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 250; i++ {
				got := a.Mul(b)
				if !fixed02BEqualRepresentation(got, want) {
					t.Error("concurrent Mul changed result")
					return
				}
			}
		}()
	}
	wg.Wait()
}

var fixed02BMulSink Optimus

func BenchmarkFixed02BMulCompare(b *testing.B) {
	for _, tc := range phaseB2MulCases() {
		b.Run(tc.name+"/original", func(b *testing.B) {
			a, c := tc.a, tc.b
			b.ReportAllocs()
			b.ResetTimer()
			var v Optimus
			for i := 0; i < b.N; i++ {
				v = fixed02BOriginalMul(a, c)
			}
			fixed02BMulSink = v
		})
		b.Run(tc.name+"/fixed02b", func(b *testing.B) {
			a, c := tc.a, tc.b
			b.ReportAllocs()
			b.ResetTimer()
			var v Optimus
			for i := 0; i < b.N; i++ {
				v = a.Mul(c)
			}
			fixed02BMulSink = v
		})
	}
	a, c := FromInt(17), FromInt(23)
	b.Run("small/original", func(b *testing.B) {
		var v Optimus
		for i := 0; i < b.N; i++ {
			v = fixed02BOriginalMul(a, c)
		}
		fixed02BMulSink = v
	})
	b.Run("small/fixed02b", func(b *testing.B) {
		var v Optimus
		for i := 0; i < b.N; i++ {
			v = a.Mul(c)
		}
		fixed02BMulSink = v
	})
}

func TestFixed02BOverflowStress(t *testing.T) {
	word := uint128.Uint128{Lo: ^uint64(0), Hi: ^uint64(0)}
	maximum := uint512.Uint512{Lo: uint256.Uint256{Lo: word, Hi: word}, Hi: uint256.Uint256{Lo: word, Hi: word}}
	// In particular scale zero cannot represent the product after rounding.
	for _, s := range []uint8{0, 75, 77, 154} {
		a, b := New(maximum, s, false), New(maximum, s, true)
		got, want := a.Mul(b), fixed02BOriginalMul(a, b)
		if !fixed02BEqualRepresentation(got, want) {
			t.Fatalf("maximal coefficient at scale %d: got=%v want=%v", s, got.ErrorDetails(), want.ErrorDetails())
		}
	}
}

// Directed tests for half-even ties at the 512-bit boundary and for
// very small products whose best representable result is zero.
func TestFixed02BRoundingBoundaries(t *testing.T) {
	word := uint128.Uint128{Lo: ^uint64(0), Hi: ^uint64(0)}
	maximum := uint512.Uint512{Lo: uint256.Uint256{Lo: word, Hi: word}, Hi: uint256.Uint256{Lo: word, Hi: word}}
	coef := new(big.Int).Set(maximum.BigInt())
	for _, diff := range []int64{0, 2, 4, 6, 8, 10, 20} {
		u, e := uint512.FromBigInt(new(big.Int).Sub(coef, big.NewInt(diff)))
		if e != errors.None {
			t.Fatal(e)
		}
		for _, exp := range []uint8{1, 19, 77, 153, 154} {
			for _, neg := range []bool{false, true} {
				fixed02BCheck(t, New(u, exp, neg), FromString("0.5"))
			}
		}
	}
	tiny := FromString("0." + strings.Repeat("0", 153) + "1")
	for _, neg := range []bool{false, true} {
		x := tiny
		if neg {
			x = x.Neg()
		}
		fixed02BCheck(t, x, tiny)
	}
}

func BenchmarkFixed02BOverflowStress(b *testing.B) {
	word := uint128.Uint128{Lo: ^uint64(0), Hi: ^uint64(0)}
	maximum := uint512.Uint512{Lo: uint256.Uint256{Lo: word, Hi: word}, Hi: uint256.Uint256{Lo: word, Hi: word}}
	a, c := New(maximum, 75, false), New(maximum, 75, false)
	b.Run("original", func(b *testing.B) {
		var v Optimus
		for i := 0; i < b.N; i++ {
			v = fixed02BOriginalMul(a, c)
		}
		fixed02BMulSink = v
	})
	b.Run("fixed02b", func(b *testing.B) {
		var v Optimus
		for i := 0; i < b.N; i++ {
			v = a.Mul(c)
		}
		fixed02BMulSink = v
	})
}
