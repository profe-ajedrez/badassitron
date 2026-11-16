package unit

import (
	"math/big"
	"math/rand"
	"sync"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func TestComparisonWithScaleOverflowAndNegativeZero(t *testing.T) {
	max := dec128.FromString(uint128.MaxUint128Str)
	small := dec128.FromString("0.1")
	if got := max.Neg().Compare(small.Neg()); got != -1 {
		t.Fatalf("negative values with overflowing scale alignment: got %d, want -1", got)
	}
	if got := small.Neg().Compare(max.Neg()); got != 1 {
		t.Fatalf("reversed negative values: got %d, want 1", got)
	}

	minusZero := dec128.New(uint128.Zero, 3, true)
	plusZero := dec128.New(uint128.Zero, 4, false)
	for _, zero := range []dec128.Dec128{dec128.Zero, plusZero} {
		if !minusZero.Equal(zero) || minusZero.Compare(zero) != 0 || zero.Compare(minusZero) != 0 {
			t.Fatalf("signed and scaled zeros must compare equal: %v and %v", minusZero, zero)
		}
	}
	if got := minusZero.Compare(dec128.One); got != -1 {
		t.Fatalf("negative zero must be less than +1: got %d", got)
	}
	if got := minusZero.Compare(dec128.NegativeOne); got != 1 {
		t.Fatalf("negative zero must be greater than -1: got %d", got)
	}
}

func TestCompareMatchesBigRatForFullWidthCoefficients(t *testing.T) {
	rng := rand.New(rand.NewSource(128))
	for i := 0; i < 2000; i++ {
		a := dec128.New(uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 1)
		b := dec128.New(uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 1)
		ra, ok := new(big.Rat).SetString(a.String())
		if !ok {
			t.Fatalf("invalid decimal from a: %s", a)
		}
		rb, ok := new(big.Rat).SetString(b.String())
		if !ok {
			t.Fatalf("invalid decimal from b: %s", b)
		}
		want := ra.Cmp(rb)
		got := a.Compare(b)
		if got != want || b.Compare(a) != -want || a.Equal(b) != (want == 0) {
			t.Fatalf("case %d: %s vs %s: Compare=%d reverse=%d Equal=%t want=%d", i, a, b, got, b.Compare(a), a.Equal(b), want)
		}
	}
}

func TestDivRoundModesAndInexact(t *testing.T) {
	tests := []struct {
		name, a, b, want string
		prec             uint8
		mode             dec128.RoundingMode
		inexact          bool
	}{
		{"half even down", "1", "8", "0.12", 2, dec128.RoundingHalfEven, true},
		{"half away up", "1", "8", "0.13", 2, dec128.RoundingHalfAwayFromZero, true},
		{"half toward zero", "1", "8", "0.12", 2, dec128.RoundingHalfTowardZero, true},
		{"half even odd digit", "3", "40", "0.08", 2, dec128.RoundingHalfEven, true},
		{"half even even digit", "1", "40", "0.02", 2, dec128.RoundingHalfEven, true},
		{"negative floor", "-1", "8", "-0.13", 2, dec128.RoundingFloor, true},
		{"negative ceil", "-1", "8", "-0.12", 2, dec128.RoundingCeil, true},
		{"negative toward zero", "-1", "8", "-0.12", 2, dec128.RoundingTowardZero, true},
		{"negative away", "-1", "8", "-0.13", 2, dec128.RoundingAwayFromZero, true},
		{"positive floor", "1", "8", "0.12", 2, dec128.RoundingFloor, true},
		{"positive ceil", "1", "8", "0.13", 2, dec128.RoundingCeil, true},
		{"exact", "1", "8", "0.125", 3, dec128.RoundingHalfEven, false},
		{"repeating", "1", "3", "0.3333333333333333333", 19, dec128.RoundingHalfEven, true},
		{"underflow to zero", "1", "1000", "0", 2, dec128.RoundingHalfEven, true},
		{"large input preserves quotient", uint128.MaxUint128Str, uint128.MaxUint128Str, "1", 19, dec128.RoundingHalfEven, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := dec128.FromString(tc.a)
			b := dec128.FromString(tc.b)
			got, inexact := a.DivRoundWithStatus(b, tc.prec, tc.mode)
			if got.IsNaN() || got.String() != tc.want || inexact != tc.inexact || got.Exponent() != tc.prec {
				t.Fatalf("%s / %s: got %s (exp=%d, inexact=%t, err=%v), want %s (exp=%d, inexact=%t)", tc.a, tc.b, got, got.Exponent(), inexact, got.ErrorDetails(), tc.want, tc.prec, tc.inexact)
			}
			if withNoStatus := a.DivRound(b, tc.prec, tc.mode); !withNoStatus.Equal(got) {
				t.Fatalf("DivRound and DivRoundWithStatus disagree: %v vs %v", withNoStatus, got)
			}
		})
	}
}

func TestMulRoundModesAndInexact(t *testing.T) {
	tests := []struct {
		a, b, want string
		prec       uint8
		mode       dec128.RoundingMode
		inexact    bool
	}{
		{"1.25", "0.1", "0.12", 2, dec128.RoundingHalfEven, true},
		{"1.25", "0.1", "0.13", 2, dec128.RoundingHalfAwayFromZero, true},
		{"-1.25", "0.1", "-0.13", 2, dec128.RoundingFloor, true},
		{"-1.25", "0.1", "-0.12", 2, dec128.RoundingCeil, true},
		{"3", "0.5", "2", 0, dec128.RoundingHalfEven, true},
		{"5", "0.5", "2", 0, dec128.RoundingHalfEven, true},
		{"0.0000000000000000001", "0.0000000000000000001", "0", 19, dec128.RoundingHalfAwayFromZero, true},
		{"1.23000000", "2.0", "2.46", 2, dec128.RoundingHalfEven, false},
		{"0", "0.123", "0", 2, dec128.RoundingHalfEven, false},
	}
	for _, tc := range tests {
		a := dec128.FromString(tc.a)
		b := dec128.FromString(tc.b)
		got, inexact := a.MulRoundWithStatus(b, tc.prec, tc.mode)
		if got.IsNaN() || got.String() != tc.want || inexact != tc.inexact || got.Exponent() != tc.prec {
			t.Errorf("%s * %s: got %s (exp=%d, inexact=%t, err=%v), want %s (exp=%d, inexact=%t)", tc.a, tc.b, got, got.Exponent(), inexact, got.ErrorDetails(), tc.want, tc.prec, tc.inexact)
		}
		if withNoStatus := a.MulRound(b, tc.prec, tc.mode); !withNoStatus.Equal(got) {
			t.Errorf("MulRound and MulRoundWithStatus disagree: %v vs %v", withNoStatus, got)
		}
	}
}

func TestRoundedOperationsFailures(t *testing.T) {
	one := dec128.One
	zero := dec128.Zero
	max := dec128.FromString(uint128.MaxUint128Str)
	for name, got := range map[string]dec128.Dec128{
		"division by zero":   one.DivRound(zero, 2, dec128.RoundingHalfEven),
		"division overflow":  max.DivRound(dec128.FromString("0.1"), 0, dec128.RoundingTowardZero),
		"product overflow":   max.MulRound(dec128.FromString("1.1"), 0, dec128.RoundingHalfEven),
		"division precision": one.DivRound(one, 20, dec128.RoundingHalfEven),
		"product precision":  one.MulRound(one, 20, dec128.RoundingHalfEven),
		"division bad mode":  one.DivRound(one, 2, dec128.RoundingMode(255)),
		"product bad mode":   one.MulRound(one, 2, dec128.RoundingMode(255)),
		"division NaN":       dec128.NaN(1).DivRound(one, 2, dec128.RoundingHalfEven),
		"multiplication NaN": dec128.NaN(1).MulRound(one, 2, dec128.RoundingHalfEven),
	} {
		if !got.IsNaN() {
			t.Errorf("%s: expected NaN, got %s", name, got)
		}
	}
}

// The SUNAT invoice example proves why the published two-decimal values must
// not be fed back as the next calculation's inputs.
func TestSUNATInvoiceAvoidsPrematureRounding(t *testing.T) {
	dec128.SetDefaultPrecision(2)
	defer dec128.SetDefaultPrecision(dec128.MaxPrecision)

	mode := dec128.RoundingHalfEven
	unitValue := dec128.FromString("98.00").DivRound(dec128.FromString("1.18"), 19, mode)
	if unitValue.IsNaN() {
		t.Fatalf("unit value: %v", unitValue.ErrorDetails())
	}
	gross := unitValue.MulRound(dec128.FromString("2000"), 19, mode)
	net := gross.MulRound(dec128.FromString("0.90"), 2, mode)
	if net.StringFixed() != "149491.53" {
		t.Fatalf("expected 149491.53, got %s (unit=%s, gross=%s)", net.StringFixed(), unitValue, gross)
	}
	if net.IsNaN() {
		t.Fatalf("net is NaN: %v", net.ErrorDetails())
	}
}

// math/big.Rat.FloatString is an independent reference for nearest rounding
// with ties away from zero. It also exercises full-width 128-bit inputs.
func TestRoundedOperationsMatchBigRat(t *testing.T) {
	rng := rand.New(rand.NewSource(20261112))
	for i := 0; i < 600; i++ {
		a := dec128.New(uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 1)
		b := dec128.New(uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 1)
		prec := uint8(rng.Intn(20))
		ra, _ := new(big.Rat).SetString(a.String())
		rb, _ := new(big.Rat).SetString(b.String())
		for _, op := range []struct {
			name string
			got  dec128.Dec128
			want *big.Rat
		}{
			{"multiply", a.MulRound(b, prec, dec128.RoundingHalfAwayFromZero), new(big.Rat).Mul(ra, rb)},
			{"divide", a.DivRound(b, prec, dec128.RoundingHalfAwayFromZero), new(big.Rat).Quo(ra, rb)},
		} {
			want := dec128.FromString(op.want.FloatString(int(prec)))
			if want.IsNaN() {
				if !op.got.IsNaN() {
					t.Fatalf("case %d %s: expected unrepresentable, got %s", i, op.name, op.got)
				}
				continue
			}
			if op.got.IsNaN() || !op.got.Equal(want) || op.got.Exponent() != prec {
				t.Fatalf("case %d %s %s %s at scale %d: got %s (err %v), want %s", i, op.name, a, b, prec, op.got, op.got.ErrorDetails(), want)
			}
		}
	}
}

// SetDefaultPrecision remains process-wide for legacy Div/Sqrt, but its
// concurrent reads and updates must not be a data race.
func TestConcurrentDefaultPrecisionAccess(t *testing.T) {
	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if n == 0 {
				for i := 0; i < 500; i++ {
					dec128.SetDefaultPrecision(uint8(i % 20))
				}
			} else {
				for i := 0; i < 500; i++ {
					if got := dec128.One.Div(dec128.FromInt(7)); got.IsNaN() {
						t.Errorf("legacy division returned %v", got.ErrorDetails())
					}
					if got := dec128.FromInt(4).Sqrt(); got.IsNaN() {
						t.Errorf("legacy square root returned %v", got.ErrorDetails())
					}
				}
			}
		}(n)
	}
	wg.Wait()
	dec128.SetDefaultPrecision(dec128.MaxPrecision)
}

var benchmarkDecimalSink dec128.Dec128

func BenchmarkMulRound(b *testing.B) {
	a, c := dec128.FromString("123456.123456789"), dec128.FromString("0.123456789")
	b.ReportAllocs()
	var got dec128.Dec128
	for i := 0; i < b.N; i++ {
		got = a.MulRound(c, 2, dec128.RoundingHalfEven)
	}
	benchmarkDecimalSink = got
}

func BenchmarkDivRound(b *testing.B) {
	a, c := dec128.FromString("123456.123456789"), dec128.FromString("1.18")
	b.ReportAllocs()
	var got dec128.Dec128
	for i := 0; i < b.N; i++ {
		got = a.DivRound(c, 19, dec128.RoundingHalfEven)
	}
	benchmarkDecimalSink = got
}

func BenchmarkMulLegacy(b *testing.B) {
	a, c := dec128.FromString("123456.123456789"), dec128.FromString("0.123456789")
	b.ReportAllocs()
	var got dec128.Dec128
	for i := 0; i < b.N; i++ {
		got = a.Mul(c)
	}
	benchmarkDecimalSink = got
}

func BenchmarkDivLegacy(b *testing.B) {
	a, c := dec128.FromString("123456.123456789"), dec128.FromString("1.18")
	b.ReportAllocs()
	var got dec128.Dec128
	for i := 0; i < b.N; i++ {
		got = a.Div(c)
	}
	benchmarkDecimalSink = got
}
