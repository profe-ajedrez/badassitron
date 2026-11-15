package dec1024

import (
	"math/big"
	"math/rand"
	"strings"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec1024/uint1024"
	"github.com/profe-ajedrez/badassitron/internal/decimalwide"
)

// big.Int belongs only to the independent reference path in these tests.
func referenceNumber(a Dec1024) decimalwide.Number {
	if a.IsNaN() {
		return decimalwide.Invalid(a.err)
	}
	return decimalwide.New(a.coef.BigInt(), int(a.exp), a.neg)
}

func assertReference(t *testing.T, label string, result Dec1024, reference decimalwide.Number) {
	t.Helper()
	if reference.Err != result.err {
		t.Fatalf("%s: error mismatch got=%v ref=%v", label, result.err, reference.Err)
	}
	if result.IsNaN() {
		return
	}
	actual := referenceNumber(result)
	if decimalwide.Cmp(actual, reference) != 0 {
		t.Fatalf("%s: value mismatch got=%v ref=%s (scale=%d)", label, result, decimalwide.Format(reference, true), reference.Scale)
	}
}

func randomDecimal(r *rand.Rand) Dec1024 {
	var b [128]byte
	r.Read(b[:])
	// Include tiny and large coefficients, all decimal scale ranges.
	if r.Intn(3) == 0 {
		for i := range b[20:] {
			b[i+20] = 0
		}
	}
	scale := uint16(r.Intn(int(MaxPrecision) + 1))
	return New(uint1024.FromBytes(b), scale, r.Intn(2) == 0)
}

func TestFixedWideDifferentialVsOriginalEngine(t *testing.T) {
	r := rand.New(rand.NewSource(20261114))
	for i := 0; i < 160; i++ {
		a, b := randomDecimal(r), randomDecimal(r)
		if b.coef.IsZero() {
			b = FromInt(3)
		}
		an, bn := referenceNumber(a), referenceNumber(b)
		if got, want := a.Compare(b), decimalwide.Cmp(an, bn); got != want {
			t.Fatalf("compare[%d] got=%d want=%d", i, got, want)
		}
		assertReference(t, "add", a.Add(b), decimalwide.Add(an, bn, 1024, int(MaxPrecision)))
		assertReference(t, "sub", a.Sub(b), decimalwide.Sub(an, bn, 1024, int(MaxPrecision)))
		assertReference(t, "mul", a.Mul(b), decimalwide.Mul(an, bn, 1024, int(MaxPrecision)))
		wantQ, wantR := decimalwide.QuoRem(an, bn, 1024, int(MaxPrecision))
		gotQ, gotR := a.QuoRem(b)
		assertReference(t, "quorem/quotient", gotQ, wantQ)
		assertReference(t, "quorem/remainder", gotR, wantR)
		assertReference(t, "mod", a.Mod(b), decimalwide.Mod(an, bn, 1024, int(MaxPrecision)))
		target := uint16(r.Intn(int(MaxPrecision) + 1))
		mode := RoundingMode(r.Intn(7))
		got, inexact := a.MulRoundWithStatus(b, target, mode)
		prod := new(big.Int).Mul(an.Coef, bn.Coef)
		want, wantInexact := decimalwide.Quantize(decimalwide.New(prod, an.Scale+bn.Scale, an.Neg != bn.Neg), int(target), 1024, int(MaxPrecision), mode)
		assertReference(t, "mulround", got, want)
		if inexact != wantInexact {
			t.Fatalf("mulround[%d] inexact got=%v want=%v", i, inexact, wantInexact)
		}
		got, inexact = a.DivRoundWithStatus(b, target, mode)
		want, wantInexact = decimalwide.Div(an, bn, int(target), 1024, int(MaxPrecision), mode)
		assertReference(t, "divround", got, want)
		if inexact != wantInexact {
			t.Fatalf("divround[%d] inexact got=%v want=%v", i, inexact, wantInexact)
		}
		if i < 75 {
			got, inexact = a.Abs().SqrtRoundWithStatus(target, mode)
			want, wantInexact = decimalwide.Sqrt(referenceNumber(a.Abs()), int(target), 1024, int(MaxPrecision), mode)
			assertReference(t, "sqrt", got, want)
			if inexact != wantInexact {
				t.Fatalf("sqrt[%d] inexact got=%v want=%v", i, inexact, wantInexact)
			}
		}
		if i < 80 {
			exponent := r.Intn(17) - 8
			assertReference(t, "pow", a.PowInt(exponent), decimalwide.PowInt(an, exponent, 1024, int(MaxPrecision), int(defaultPrecision.Load())))
		}
		refCanonical := decimalwide.Canonical(an)
		assertReference(t, "canonical", a.Canonical(), refCanonical)
		if got := a.StringFixed(); got != decimalwide.Format(an, true) {
			t.Fatalf("format fixed: got=%q expected=%q", got, decimalwide.Format(an, true))
		}
		if got := a.String(); got != decimalwide.Format(an, false) {
			t.Fatalf("format: got=%q expected=%q", got, decimalwide.Format(an, false))
		}
	}
}

func TestFixedWideDecimalInputParity(t *testing.T) {
	inputs := []string{"", "0", "-0", "+0.00", "-.1", "1.", ".1", "-", "+", "1..2", "1e5", "00.00100", "--2", "2.5", "NaN", "1 ", " 1", strings.Repeat("9", 309), strings.Repeat("9", 310), "0." + strings.Repeat("0", 307) + "1", "0." + strings.Repeat("0", 308) + "1"}
	for _, s := range inputs {
		got := FromString(s)
		ref := decimalwide.Parse(s, 1024, int(MaxPrecision))
		assertReference(t, "parse:"+s, got, ref)
		if got.err == 0 && got.StringFixed() != decimalwide.Format(ref, true) {
			t.Fatalf("parse representation mismatch input=%q result=%q expected=%q", s, got.StringFixed(), decimalwide.Format(ref, true))
		}
	}
}

func TestFixedWideRoundingBoundaries(t *testing.T) {
	a := FromString("1" + strings.Repeat("0", 300) + "." + strings.Repeat("0", 7) + "5")
	if a.IsNaN() {
		t.Fatal("bad fixture")
	}
	for _, m := range []RoundingMode{RoundingTowardZero, RoundingAwayFromZero, RoundingFloor, RoundingCeil, RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven} {
		for _, p := range []uint16{0, 1, 2, 17, 300, 308} {
			for _, x := range []Dec1024{a, a.Neg(), FromString("0." + strings.Repeat("0", 307) + "1"), FromString("1.25"), FromString("2.5")} {
				// The numerator can overflow for high output scales; compare the reported errors as well.
				got, ix := x.SqrtRoundWithStatus(p, m)
				want, iw := decimalwide.Sqrt(referenceNumber(x), int(p), 1024, int(MaxPrecision), m)
				assertReference(t, "sqrt boundaries", got, want)
				if ix != iw {
					t.Fatalf("sqrt boundary inexact mismatch mode=%d scale=%d", m, p)
				}
			}
		}
	}
}

func TestFixedWideNoHeapArithmetic(t *testing.T) {
	x := FromString("123.4567")
	y := FromString("89.1234")
	cases := map[string]func(){
		"add":      func() { _ = x.Add(y) },
		"mul":      func() { _ = x.Mul(y) },
		"mulround": func() { _, _ = x.MulRoundWithStatus(y, 2, RoundingHalfEven) },
		"divround": func() { _, _ = x.DivRoundWithStatus(y, 2, RoundingHalfEven) },
		"compare":  func() { _ = x.Compare(y) },
		"sqrt":     func() { _, _ = x.SqrtRoundWithStatus(2, RoundingHalfEven) },
	}
	for name, run := range cases {
		allocations := testing.AllocsPerRun(100, run)
		if allocations != 0 {
			t.Errorf("%s allocated: %v per run", name, allocations)
		}
	}
}

func TestFixedWideFallbackNoHeap(t *testing.T) {
	unity := FromString("1." + strings.Repeat("0", 300))
	maxScale := New(uint1024.Max, 308, false)
	maxAtOne := New(uint1024.Max, 1, false)
	tiny := New(uint1024.FromUint64(3), 308, false)
	cases := map[string]func(){
		"wide-add":      func() { _ = maxAtOne.Add(maxAtOne) },
		"wide-mul":      func() { _ = unity.Mul(unity) },
		"wide-mulround": func() { _, _ = unity.MulRoundWithStatus(unity, 2, RoundingHalfEven) },
		"wide-div":      func() { _, _ = maxScale.DivRoundWithStatus(maxScale, 308, RoundingHalfEven) },
		"wide-sqrt":     func() { _, _ = FromInt(2).SqrtRoundWithStatus(308, RoundingHalfEven) },
		"wide-mod":      func() { _ = FromInt(5).Mod(tiny) },
	}
	for name, run := range cases {
		if got := testing.AllocsPerRun(20, run); got != 0 {
			t.Errorf("%s: %v allocs/op", name, got)
		}
	}
}
