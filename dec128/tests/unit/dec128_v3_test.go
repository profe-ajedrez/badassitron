package unit

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// v3ReferenceSqrt calculates the nearest integer to sqrt(value)*10^precision
// using exact rational arithmetic. Only the test oracle uses math/big.
func v3ReferenceSqrt(t *testing.T, value dec128.Dec128, precision uint8, mode dec128.RoundingMode) (*big.Int, bool) {
	t.Helper()
	input := v2DecimalRat(t, value)
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision)*2), nil)
	squaredScaled := new(big.Rat).Mul(input, new(big.Rat).SetInt(factor))
	quotient := new(big.Int).Quo(squaredScaled.Num(), squaredScaled.Denom())
	root := new(big.Int).Sqrt(quotient)
	rootSquared := new(big.Int).Mul(root, root)
	inexact := squaredScaled.Cmp(new(big.Rat).SetInt(rootSquared)) != 0
	if !inexact {
		return root, false
	}
	increment := false
	switch mode {
	case dec128.RoundingAwayFromZero, dec128.RoundingCeil:
		increment = true
	case dec128.RoundingHalfTowardZero, dec128.RoundingHalfAwayFromZero, dec128.RoundingHalfEven:
		midpointDouble := new(big.Int).Add(new(big.Int).Mul(root, big.NewInt(2)), big.NewInt(1))
		midpoint := new(big.Rat).SetFrac(new(big.Int).Mul(midpointDouble, midpointDouble), big.NewInt(4))
		compare := squaredScaled.Cmp(midpoint)
		increment = compare > 0 || (compare == 0 && (mode == dec128.RoundingHalfAwayFromZero || mode == dec128.RoundingHalfEven && root.Bit(0) != 0))
	}
	if increment {
		root.Add(root, big.NewInt(1))
	}
	return root, true
}

func TestV3SqrtRoundExactTieAndDirectedModes(t *testing.T) {
	tests := []struct {
		value, want string
		precision   uint8
		mode        dec128.RoundingMode
		inexact     bool
	}{
		{"2.25", "2", 0, dec128.RoundingHalfEven, true},
		{"2.25", "1", 0, dec128.RoundingHalfTowardZero, true},
		{"2.25", "2", 0, dec128.RoundingHalfAwayFromZero, true},
		{"6.25", "2", 0, dec128.RoundingHalfEven, true},
		{"6.25", "3", 0, dec128.RoundingHalfAwayFromZero, true},
		{"0.0225", "0.2", 1, dec128.RoundingHalfEven, true},
		{"0.0625", "0.2", 1, dec128.RoundingHalfEven, true},
		{"1.5625", "1.2", 1, dec128.RoundingHalfEven, true},
		{"2", "1.4142135623730950488", 19, dec128.RoundingHalfEven, true},
		{"2", "1.4142135623730950489", 19, dec128.RoundingAwayFromZero, true},
		{"2", "1", 0, dec128.RoundingTowardZero, true},
		{"2", "2", 0, dec128.RoundingCeil, true},
		{"0.0000000000000000001", "0", 0, dec128.RoundingHalfEven, true},
		{"0.0000000000000000001", "1", 0, dec128.RoundingAwayFromZero, true},
		{"4", "2", 0, dec128.RoundingHalfEven, false},
		{"4", "2", 0, dec128.RoundingAwayFromZero, false},
		{"0", "0", 19, dec128.RoundingCeil, false},
		{"1", "1", 19, dec128.RoundingHalfEven, false},
	}
	for _, tc := range tests {
		v := dec128.FromString(tc.value)
		got, inexact := v.SqrtRoundWithStatus(tc.precision, tc.mode)
		if got.IsNaN() || got.String() != tc.want || inexact != tc.inexact || got.Exponent() != tc.precision {
			t.Errorf("sqrt(%s), precision=%d, mode=%d: got=%s, exponent=%d, inexact=%t, error=%v; want=%s, inexact=%t", tc.value, tc.precision, tc.mode, got, got.Exponent(), inexact, got.ErrorDetails(), tc.want, tc.inexact)
		}
		if plain := v.SqrtRound(tc.precision, tc.mode); plain.IsNaN() || !plain.Equal(got) || plain.Exponent() != tc.precision {
			t.Errorf("SqrtRound mismatch for %s at precision %d", tc.value, tc.precision)
		}
	}
}

func TestV3SqrtRoundInvalidInputsAndGlobalIndependence(t *testing.T) {
	for _, tc := range []struct {
		name string
		got  dec128.Dec128
	}{
		{"negative", dec128.FromString("-1").SqrtRound(2, dec128.RoundingHalfEven)},
		{"precision", dec128.One.SqrtRound(20, dec128.RoundingHalfEven)},
		{"invalid mode", dec128.One.SqrtRound(2, dec128.RoundingMode(255))},
		{"NaN", dec128.NaN(1).SqrtRound(2, dec128.RoundingHalfEven)},
	} {
		if !tc.got.IsNaN() {
			t.Errorf("%s: expected NaN, got %s", tc.name, tc.got)
		}
	}
	negativeZero := dec128.New(uint128.Zero, 3, true)
	if got, inexact := negativeZero.SqrtRoundWithStatus(10, dec128.RoundingCeil); got.IsNaN() || !got.IsZero() || got.IsNegative() || got.Exponent() != 10 || inexact {
		t.Errorf("negative zero: got %s, inexact=%t", got, inexact)
	}
	dec128.SetDefaultPrecision(0)
	defer dec128.SetDefaultPrecision(dec128.MaxPrecision)
	if got := dec128.FromString("2").SqrtRound(19, dec128.RoundingHalfEven); got.String() != "1.4142135623730950488" {
		t.Errorf("explicit precision unexpectedly affected by global setting: %s", got)
	}
	if got := dec128.FromString("2").Sqrt(); got.String() != "1" {
		t.Errorf("legacy precision unexpectedly changed: %s", got)
	}
}

func TestV3SqrtRoundAgainstExactRationalOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(2026111204))
	cases := []dec128.Dec128{
		dec128.Zero, dec128.One, dec128.FromString("2"), dec128.FromString("2.25"),
		dec128.FromString("6.25"), dec128.FromString("0.0000000000000000001"),
		dec128.New(uint128.Max, 0, false), dec128.New(uint128.Max, 19, false),
	}
	for i := 0; i < 1300; i++ {
		cases = append(cases, dec128.New(uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), false))
	}
	for i, input := range cases {
		precision := uint8(rng.Intn(20))
		if i < 8 {
			precision = uint8([]int{0, 1, 6, 10, 19, 0, 19, 19}[i])
		}
		for mode := dec128.RoundingTowardZero; mode <= dec128.RoundingHalfEven; mode++ {
			want, inexact := v3ReferenceSqrt(t, input, precision, mode)
			got, gotInexact := input.SqrtRoundWithStatus(precision, mode)
			if got.IsNaN() || got.Exponent() != precision || got.Coefficient().String() != want.String() || gotInexact != inexact {
				t.Fatalf("case %d sqrt(%s) scale=%d mode=%d => %s [coef=%s inexact=%t error=%v], expected coef=%s inexact=%t", i, input, precision, mode, got, got.Coefficient(), gotInexact, got.ErrorDetails(), want, inexact)
			}
		}
	}
}

func TestV3ModHandlesOverflowingQuotients(t *testing.T) {
	max := dec128.FromString(uint128.MaxUint128Str)
	for _, input := range []dec128.Dec128{max, max.Neg()} {
		for _, divisor := range []string{"0.0000000000000000001", "0.0000000000000000003", "-0.0000000000000000007"} {
			b := dec128.FromString(divisor)
			q, r := input.QuoRem(b)
			if !q.IsNaN() || !r.IsNaN() {
				t.Fatalf("%s / %s: expected unrepresentable quotient, got (%s,%s)", input, b, q, r)
			}
			aRat, bRat := v2DecimalRat(t, input), v2DecimalRat(t, b)
			ratio := new(big.Rat).Quo(aRat, bRat)
			trunc := new(big.Int).Quo(ratio.Num(), ratio.Denom())
			want := new(big.Rat).Sub(aRat, new(big.Rat).Mul(new(big.Rat).SetInt(trunc), bRat))
			got := input.Mod(b)
			if got.IsNaN() || v2DecimalRat(t, got).Cmp(want) != 0 {
				t.Errorf("%s %% %s => %s (%v), want %s", input, b, got, got.ErrorDetails(), want)
			}
		}
	}
}

func TestV3ModNormalizesZeroText(t *testing.T) {
	if got := dec128.FromString("-2").Mod(dec128.FromString("1")); !got.IsZero() || got.String() != "0" {
		t.Fatalf("zero remainder should retain canonical text: %s", got)
	}
}

func TestV3ModAgainstExactRationalOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(2026111205))
	for i := 0; i < 3500; i++ {
		a, b := v2RandomDecimal(rng), v2RandomDecimal(rng)
		if b.IsZero() {
			b = dec128.One
		}
		aRat, bRat := v2DecimalRat(t, a), v2DecimalRat(t, b)
		ratio := new(big.Rat).Quo(aRat, bRat)
		trunc := new(big.Int).Quo(ratio.Num(), ratio.Denom())
		want := new(big.Rat).Sub(aRat, new(big.Rat).Mul(new(big.Rat).SetInt(trunc), bRat))
		got := a.Mod(b)
		if got.IsNaN() || v2DecimalRat(t, got).Cmp(want) != 0 {
			t.Fatalf("case %d: %s %% %s => %s (%v), want %s", i, a, b, got, got.ErrorDetails(), want)
		}
	}
}

func BenchmarkV3SqrtRound(b *testing.B) {
	input := dec128.FromString("2")
	b.ReportAllocs()
	var got dec128.Dec128
	for i := 0; i < b.N; i++ {
		got = input.SqrtRound(19, dec128.RoundingHalfEven)
	}
	benchmarkDecimalSink = got
}

func BenchmarkV3SqrtRoundMax(b *testing.B) {
	input := dec128.New(uint128.Max, 0, false)
	b.ReportAllocs()
	var got dec128.Dec128
	for i := 0; i < b.N; i++ {
		got = input.SqrtRound(19, dec128.RoundingHalfEven)
	}
	benchmarkDecimalSink = got
}

func BenchmarkV3ModOverflowQuotient(b *testing.B) {
	input := dec128.New(uint128.Max, 0, false)
	divisor := dec128.FromString("0.0000000000000000003")
	b.ReportAllocs()
	var got dec128.Dec128
	for i := 0; i < b.N; i++ {
		got = input.Mod(divisor)
	}
	benchmarkDecimalSink = got
}
