package unit

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func TestV4RescaleRoundDirected(t *testing.T) {
	tests := []struct {
		name, input, output string
		precision           uint8
		mode                dec128.RoundingMode
		inexact             bool
	}{
		{"half even from odd", "1.235", "1.24", 2, dec128.RoundingHalfEven, true},
		{"half even from even", "1.245", "1.24", 2, dec128.RoundingHalfEven, true},
		{"half away from even", "1.245", "1.25", 2, dec128.RoundingHalfAwayFromZero, true},
		{"half toward zero", "1.235", "1.23", 2, dec128.RoundingHalfTowardZero, true},
		{"positive floor", "1.239", "1.23", 2, dec128.RoundingFloor, true},
		{"positive ceil", "1.231", "1.24", 2, dec128.RoundingCeil, true},
		{"negative floor", "-1.231", "-1.24", 2, dec128.RoundingFloor, true},
		{"negative ceil", "-1.239", "-1.23", 2, dec128.RoundingCeil, true},
		{"negative half even odd", "-1.235", "-1.24", 2, dec128.RoundingHalfEven, true},
		{"negative half even even", "-1.245", "-1.24", 2, dec128.RoundingHalfEven, true},
		{"negative toward zero", "-1.239", "-1.23", 2, dec128.RoundingTowardZero, true},
		{"negative away from zero", "-1.231", "-1.24", 2, dec128.RoundingAwayFromZero, true},
		{"carry across decimal point", "9.995", "10", 2, dec128.RoundingHalfEven, true},
		{"exact scale reduction", "1.2300", "1.23", 2, dec128.RoundingHalfEven, false},
		{"exact scale increase", "1.23", "1.23", 6, dec128.RoundingHalfEven, false},
		{"exact scale unchanged", "-1.230", "-1.23", 3, dec128.RoundingHalfEven, false},
		{"underflow nearest", "-0.0000000000000000001", "0", 0, dec128.RoundingHalfEven, true},
		{"underflow floor", "-0.0000000000000000001", "-1", 0, dec128.RoundingFloor, true},
		{"underflow ceil", "-0.0000000000000000001", "0", 0, dec128.RoundingCeil, true},
		{"underflow away", "0.0000000000000000001", "1", 0, dec128.RoundingAwayFromZero, true},
		{"max coefficient exact unchanged", uint128.MaxUint128Str, uint128.MaxUint128Str, 0, dec128.RoundingHalfEven, false},
		{"max coefficient scale reduction", "34028236692093846346.3374607431768211455", "34028236692093846346.337460743176821146", 18, dec128.RoundingHalfEven, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := dec128.FromString(tc.input)
			if input.IsNaN() {
				t.Fatalf("invalid fixture %q: %v", tc.input, input.ErrorDetails())
			}
			got, inexact := input.RescaleRoundWithStatus(tc.precision, tc.mode)
			withoutStatus := input.RescaleRound(tc.precision, tc.mode)
			if got.IsNaN() || got.String() != tc.output || got.Exponent() != tc.precision || inexact != tc.inexact {
				t.Fatalf("%s -> scale %d mode %d: got=%s exp=%d inexact=%t error=%v; want %s exp=%d inexact=%t", tc.input, tc.precision, tc.mode, got, got.Exponent(), inexact, got.ErrorDetails(), tc.output, tc.precision, tc.inexact)
			}
			if withoutStatus.Exponent() != got.Exponent() || !withoutStatus.Equal(got) {
				t.Fatalf("variants disagree: %s vs %s", withoutStatus, got)
			}
		})
	}
}

func TestV4RescaleRoundErrorsAndLegacy(t *testing.T) {
	max := dec128.FromString(uint128.MaxUint128Str)
	for _, tc := range []struct {
		name  string
		got   dec128.Dec128
		cause error
	}{
		{"overflow", max.RescaleRound(1, dec128.RoundingHalfEven), errors.Overflow.Value()},
		{"precision", dec128.One.RescaleRound(20, dec128.RoundingHalfEven), errors.PrecisionOutOfRange.Value()},
		{"bad mode", dec128.One.RescaleRound(0, dec128.RoundingMode(255)), errors.InvalidFormat.Value()},
	} {
		if !tc.got.IsNaN() || tc.got.ErrorDetails() != tc.cause {
			t.Errorf("%s: got %s error %v; want NaN %v", tc.name, tc.got, tc.got.ErrorDetails(), tc.cause)
		}
	}
	for _, tc := range []struct {
		input string
		prec  uint8
	}{
		{"123.451", 2}, {"-123.451", 2}, {"1.23000", 2}, {"0", 19},
	} {
		input := dec128.FromString(tc.input)
		got := input.RescaleRound(tc.prec, dec128.RoundingTowardZero)
		legacy := input.Rescale(tc.prec)
		if got.IsNaN() != legacy.IsNaN() || (!got.IsNaN() && (got.Exponent() != legacy.Exponent() || got.Coefficient() != legacy.Coefficient() || got.Sign() != legacy.Sign())) {
			t.Fatalf("legacy Rescale mismatch for %s: %s vs %s", tc.input, got, legacy)
		}
	}
	originalNaN := dec128.NaN(errors.SqrtNegative)
	gotNaN, inexactNaN := originalNaN.RescaleRoundWithStatus(2, dec128.RoundingHalfEven)
	if !gotNaN.IsNaN() || gotNaN.ErrorDetails() != errors.SqrtNegative.Value() || inexactNaN {
		t.Fatalf("NaN propagation lost: %s %v %t", gotNaN, gotNaN.ErrorDetails(), inexactNaN)
	}
	if got, inexact := dec128.One.RescaleRoundWithStatus(20, dec128.RoundingHalfEven); !got.IsNaN() || inexact {
		t.Errorf("invalid precision should return NaN,false; got %s,%t", got, inexact)
	}
	if got, inexact := dec128.One.RescaleRoundWithStatus(0, dec128.RoundingMode(255)); !got.IsNaN() || inexact {
		t.Errorf("invalid mode should return NaN,false; got %s,%t", got, inexact)
	}
	if got, inexact := max.RescaleRoundWithStatus(1, dec128.RoundingHalfEven); !got.IsNaN() || inexact {
		t.Errorf("upscale overflow should return NaN,false; got %s,%t", got, inexact)
	}

	// Explicit-precision arithmetic must ignore the global setting.
	defer dec128.SetDefaultPrecision(dec128.MaxPrecision)
	input := dec128.FromString("123.4567")
	dec128.SetDefaultPrecision(0)
	a, aInexact := input.RescaleRoundWithStatus(2, dec128.RoundingHalfEven)
	dec128.SetDefaultPrecision(19)
	b, bInexact := input.RescaleRoundWithStatus(2, dec128.RoundingHalfEven)
	if a.Exponent() != b.Exponent() || !a.Equal(b) || aInexact != bInexact {
		t.Errorf("global precision affected result: %s,%t vs %s,%t", a, aInexact, b, bInexact)
	}
}

// The published two-decimal amount must be rounded from the high-precision
// result, not computed from already rounded display intermediates.
func TestV4SUNATInvoiceRescalePublication(t *testing.T) {
	mode := dec128.RoundingHalfEven
	unit := dec128.FromString("98.00").DivRound(dec128.FromString("1.18"), 19, mode)
	subtotal := unit.MulRound(dec128.FromString("2000"), 19, mode)
	net := subtotal.MulRound(dec128.FromString("0.90"), 19, mode)
	published, inexact := net.RescaleRoundWithStatus(2, mode)
	if published.IsNaN() || published.StringFixed() != "149491.53" || !inexact {
		t.Fatalf("expected 149491.53 inexact=true, got %s, %t (unit=%s subtotal=%s net=%s)", published.StringFixed(), inexact, unit, subtotal, net)
	}
}

// rescaleV4Oracle computes exact fixed-point rescaling with math/big only
// within tests. It independently implements all seven rounding modes.
func rescaleV4Oracle(input dec128.Dec128, prec uint8, mode dec128.RoundingMode) (*big.Int, bool) {
	coefficient := new(big.Int)
	coefficient.SetString(input.Coefficient().String(), 10)
	numerator := new(big.Int).Mul(coefficient, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(prec)), nil))
	denominator := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(input.Exponent())), nil)
	q, r := new(big.Int).QuoRem(numerator, denominator, new(big.Int))
	inexact := r.Sign() != 0
	if !inexact {
		return q, false
	}
	inc := false
	switch mode {
	case dec128.RoundingTowardZero:
	case dec128.RoundingAwayFromZero:
		inc = true
	case dec128.RoundingFloor:
		inc = input.Sign() < 0
	case dec128.RoundingCeil:
		inc = input.Sign() > 0
	case dec128.RoundingHalfTowardZero, dec128.RoundingHalfAwayFromZero, dec128.RoundingHalfEven:
		cmp := new(big.Int).Lsh(r, 1).Cmp(denominator)
		inc = cmp > 0 || cmp == 0 && (mode == dec128.RoundingHalfAwayFromZero || mode == dec128.RoundingHalfEven && q.Bit(0) == 1)
	}
	if inc {
		q.Add(q, big.NewInt(1))
	}
	return q, true
}

func TestV4RescaleRoundBigIntOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(2026111204))
	cases := []dec128.Dec128{
		dec128.Zero, dec128.New(uint128.Zero, 19, true),
		dec128.One, dec128.NegativeOne,
		dec128.New(uint128.Max, 0, false),
		dec128.New(uint128.Max, 19, false),
		dec128.New(uint128.Max, 19, true),
		dec128.FromString("0.0000000000000000001"),
		dec128.FromString("-0.0000000000000000001"),
		dec128.FromString("999999999999999999.995"),
	}
	for i := 0; i < 6000; i++ {
		cases = append(cases, dec128.New(uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 1))
	}
	maxCoefficient := new(big.Int)
	maxCoefficient.SetString(uint128.MaxUint128Str, 10)
	var checked, overflow, inexactCount int
	for i, input := range cases {
		var targetPrec uint8
		if i < 10 {
			// Sweep each scale for adversarial fixtures.
			for p := uint8(0); p <= dec128.MaxPrecision; p++ {
				for mode := dec128.RoundingTowardZero; mode <= dec128.RoundingHalfEven; mode++ {
					v4CheckRescaleOracle(t, input, p, mode, maxCoefficient)
					checked++
				}
			}
			continue
		}
		targetPrec = uint8(rng.Intn(20))
		for mode := dec128.RoundingTowardZero; mode <= dec128.RoundingHalfEven; mode++ {
			q, inexact := rescaleV4Oracle(input, targetPrec, mode)
			if q.Cmp(maxCoefficient) > 0 {
				overflow++
			}
			if inexact {
				inexactCount++
			}
			v4CheckRescaleOracle(t, input, targetPrec, mode, maxCoefficient)
			checked++
		}
	}
	t.Logf("verified %d independent mode/precision cases (%d overflow, %d inexact)", checked, overflow, inexactCount)
}

func v4CheckRescaleOracle(t *testing.T, input dec128.Dec128, prec uint8, mode dec128.RoundingMode, max *big.Int) {
	t.Helper()
	q, expectedInexact := rescaleV4Oracle(input, prec, mode)
	got, inexact := input.RescaleRoundWithStatus(prec, mode)
	if q.Cmp(max) > 0 {
		if !got.IsNaN() || got.ErrorDetails() != errors.Overflow.Value() || inexact {
			t.Fatalf("%s rescaled to %d mode %d: expected overflow,false, got %s,%t error=%v", input, prec, mode, got, inexact, got.ErrorDetails())
		}
		return
	}
	if got.IsNaN() || got.Coefficient().String() != q.String() || got.Exponent() != prec || inexact != expectedInexact {
		t.Fatalf("%s rescaled to %d mode %d: got %s coefficient=%s exp=%d inexact=%t error=%v, want coef=%s exp=%d inexact=%t", input, prec, mode, got, got.Coefficient(), got.Exponent(), inexact, got.ErrorDetails(), q, prec, expectedInexact)
	}
	if q.Sign() != 0 && (got.IsNegative() != input.IsNegative()) {
		t.Fatalf("%s rescaled to %d mode %d changed sign: %s", input, prec, mode, got)
	}
	if !input.RescaleRound(prec, mode).Equal(got) {
		t.Fatalf("%s rescaled to %d mode %d: methods disagree", input, prec, mode)
	}
}

var benchmarkV4Result dec128.Dec128

func BenchmarkV4RescaleRound(b *testing.B) {
	cases := []struct {
		name  string
		value dec128.Dec128
		prec  uint8
	}{
		{"round-down", dec128.FromString("1234567890123.123456789"), 2},
		{"exact-upscale", dec128.FromString("1234567890123.12"), 7},
		{"unchanged", dec128.FromString("1234567890123.12"), 2},
	}
	for _, tc := range cases {
		b.Run(tc.name+"/legacy", func(b *testing.B) {
			v := tc.value
			for i := 0; i < b.N; i++ {
				benchmarkV4Result = v.Rescale(tc.prec)
			}
		})
		b.Run(tc.name+"/new", func(b *testing.B) {
			v := tc.value
			for i := 0; i < b.N; i++ {
				benchmarkV4Result = v.RescaleRound(tc.prec, dec128.RoundingHalfEven)
			}
		})
		b.Run(tc.name+"/with-status", func(b *testing.B) {
			v := tc.value
			for i := 0; i < b.N; i++ {
				benchmarkV4Result, _ = v.RescaleRoundWithStatus(tc.prec, dec128.RoundingHalfEven)
			}
		})
	}
}
