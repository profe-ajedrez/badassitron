package optimus

import (
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

// phaseAFromParts constructs values without passing through decimal parsing.
func phaseAFromParts(t *testing.T, c *big.Int, scale uint8, negative bool) Optimus {
	t.Helper()
	u, err := uint512.FromBigInt(c)
	if err != 0 {
		t.Fatalf("invalid test coefficient: %v", err)
	}
	d := New(u, scale, negative)
	if d.IsNaN() {
		t.Fatalf("invalid test operand: %v", d.ErrorDetails())
	}
	return d
}

func phaseACheckAuto(t *testing.T, op string, a, b, got Optimus, want *big.Rat) {
	t.Helper()
	rounded, representable := oracleAuto(want)
	if !representable {
		if !got.IsNaN() {
			t.Fatalf("%s: a=%s b=%s: expected overflow, got %s", op, a, b, got)
		}
		return
	}
	if got.IsNaN() {
		t.Fatalf("%s: a=%s b=%s: unexpected NaN: %v", op, a, b, got.ErrorDetails())
	}
	if ratFrom(got).Cmp(rounded) != 0 {
		t.Fatalf("%s: a=%s b=%s: got %s (scale %d) expected %s", op, a, b, got, got.Precision(), rounded.FloatString(154))
	}
	if got.Width() != 64 && got.Width() != 128 && got.Width() != 256 && got.Width() != 512 {
		t.Fatalf("%s: invalid coefficient tier %d", op, got.Width())
	}
}

// Independently check all four coefficient tier boundaries, including carries,
// lossless cancellation, and heterogeneous scales up to the documented limit.
func TestPhaseACoefficientBoundariesAgainstRationals(t *testing.T) {
	coefficients := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(5)}
	for _, n := range []uint{64, 128, 256, 512} {
		power := new(big.Int).Lsh(big.NewInt(1), n)
		coefficients = append(coefficients, new(big.Int).Sub(power, big.NewInt(1)))
		if n < 512 {
			coefficients = append(coefficients, power)
		}
	}
	scales := []uint8{0, 1, 19, 153, 154}
	for i, ac := range coefficients {
		for j, bc := range coefficients {
			for _, as := range scales {
				bs := scales[(i+j+int(as))%len(scales)]
				a := phaseAFromParts(t, ac, as, i%2 != 0)
				b := phaseAFromParts(t, bc, bs, j%2 == 0)
				ar, br := ratFrom(a), ratFrom(b)
				phaseACheckAuto(t, "add", a, b, a.Add(b), new(big.Rat).Add(ar, br))
				phaseACheckAuto(t, "sub", a, b, a.Sub(b), new(big.Rat).Sub(ar, br))
				phaseACheckAuto(t, "mul", a, b, a.Mul(b), new(big.Rat).Mul(ar, br))
				if a.Compare(b) != ar.Cmp(br) {
					t.Fatalf("compare: a=%s b=%s", a, b)
				}
			}
		}
	}
}

func phaseARoundRat(v *big.Rat, scale int, mode RoundingMode) (*big.Int, bool) {
	numerator := new(big.Int).Mul(new(big.Int).Abs(v.Num()), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil))
	divisor := v.Denom()
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, divisor, remainder)
	inexact := remainder.Sign() != 0
	if inexact {
		more := new(big.Int).Lsh(new(big.Int).Set(remainder), 1).Cmp(divisor)
		increment := false
		switch mode {
		case RoundingTowardZero:
		case RoundingAwayFromZero:
			increment = true
		case RoundingFloor:
			increment = v.Sign() < 0
		case RoundingCeil:
			increment = v.Sign() > 0
		case RoundingHalfTowardZero:
			increment = more > 0
		case RoundingHalfAwayFromZero:
			increment = more >= 0
		case RoundingHalfEven:
			increment = more > 0 || more == 0 && quotient.Bit(0) == 1
		}
		if increment {
			quotient.Add(quotient, big.NewInt(1))
		}
	}
	if v.Sign() < 0 {
		quotient.Neg(quotient)
	}
	return quotient, inexact
}

// Explicit scale and status are compared against independent rational arithmetic.
func TestPhaseARoundingAndDivisionStatus(t *testing.T) {
	rng := rand.New(rand.NewSource(0xA551CE5))
	maxCoef := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	for i := 0; i < 2500; i++ {
		widthA := 1 + rng.Intn(512)
		widthB := 1 + rng.Intn(512)
		ac := new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(widthA)))
		bc := new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(widthB)))
		if bc.Sign() == 0 {
			bc.SetInt64(1)
		}
		a := phaseAFromParts(t, ac, uint8(rng.Intn(155)), rng.Intn(2) == 1)
		b := phaseAFromParts(t, bc, uint8(rng.Intn(155)), rng.Intn(2) == 1)
		p := uint8(rng.Intn(155))
		mode := RoundingMode(rng.Intn(7))
		v := new(big.Rat).Quo(ratFrom(a), ratFrom(b))
		expected, inexact := phaseARoundRat(v, int(p), mode)
		got, status := a.DivRoundWithStatus(b, p, mode)
		if new(big.Int).Abs(new(big.Int).Set(expected)).Cmp(maxCoef) > 0 {
			if !got.IsNaN() {
				t.Fatalf("round overflow a=%s b=%s p=%d mode=%d got=%s", a, b, p, mode, got)
			}
			continue
		}
		if got.IsNaN() || got.Precision() != p || got.coefBig().Cmp(new(big.Int).Abs(expected)) != 0 || got.IsNegative() != (expected.Sign() < 0) || status != inexact {
			t.Fatalf("DivRound a=%s b=%s p=%d mode=%d got=%s inexact=%t want coef=%s inexact=%t", a, b, p, mode, got, status, expected, inexact)
		}
	}
}

func TestPhaseAHalfEvenTiesAndOverflow(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"0.5", "0"}, {"1.5", "2"}, {"2.5", "2"}, {"3.5", "4"},
		{"-0.5", "0"}, {"-1.5", "-2"}, {"-2.5", "-2"}, {"-3.5", "-4"},
		{"0.005", "0"}, {"0.015", "0.02"}, {"-0.005", "0"}, {"-0.015", "-0.02"},
	} {
		p := uint8(0)
		if dot := strings.LastIndexByte(tc.in, '.'); dot >= 0 && len(tc.in)-dot-1 == 3 {
			p = 2
		}
		got := FromString(tc.in).RescaleRound(p, RoundingHalfEven)
		if got.IsNaN() || got.String() != tc.want {
			t.Fatalf("bank %s scale %d: %s, want %s", tc.in, p, got, tc.want)
		}
	}
	// The largest coefficient must not wrap when half-even rounding carries.
	max512 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	almost := phaseAFromParts(t, max512, 0, false)
	if !almost.Add(FromString("0.5")).IsNaN() {
		t.Fatal("half-even carry above maximum coefficient must report overflow")
	}
}

func TestPhaseAPowersAgainstExactRationals(t *testing.T) {
	cases := []struct {
		s         string
		exponents []int
	}{
		{"1.000000000000000000000000000001", []int{4095, 4096, 4097, 4098, 5000, -4097, -5000}},
		{"0.999999999999999999999999999999", []int{4095, 4097, 5001, -4097, -5001}},
		{"-1.000000000000000000000000000001", []int{4095, 4096, 4097, 5001, -5001}},
		{"-0.999999999999999999999999999999", []int{4095, 4097, 5001, -5001}},
		{"1.0000000000000000000000000000000000000001", []int{4097, 5001, -4097}},
		{"0.9", []int{4096, 4097, 5000, -4097}},
		{"0.95", []int{4097, 5000, -5000}},
		{"1.1", []int{4096, 4097, 5000, -4097}},
		{"0.5", []int{4096, 4097, 5000, -4097}},
		{"2", []int{4096, 4097, 5000, -4097}},
	}
	for _, tc := range cases {
		base := FromString(tc.s)
		for _, exp := range tc.exponents {
			magnitude := exp
			if magnitude < 0 {
				magnitude = -magnitude
			}
			numerator := new(big.Int).Exp(base.coefBig(), big.NewInt(int64(magnitude)), nil)
			denominator := new(big.Int).Exp(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(base.exp)), nil), big.NewInt(int64(magnitude)), nil)
			if exp < 0 {
				numerator, denominator = denominator, numerator
			}
			if base.neg && magnitude%2 == 1 {
				numerator.Neg(numerator)
			}
			want := new(big.Rat).SetFrac(numerator, denominator)
			got := base.PowInt(exp)
			phaseACheckAuto(t, "power", base, FromInt(exp), got, want)
		}
	}
}

func TestPhaseAQuoRemInvariant(t *testing.T) {
	rng := rand.New(rand.NewSource(math.MaxInt32))
	for i := 0; i < 1000; i++ {
		a := phaseAFromParts(t, new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(1+rng.Intn(512)))), uint8(rng.Intn(155)), rng.Intn(2) == 1)
		b := phaseAFromParts(t, new(big.Int).Add(big.NewInt(1), new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(1+rng.Intn(512))))), uint8(rng.Intn(155)), rng.Intn(2) == 1)
		av, bv := ratFrom(a), ratFrom(b)
		qwant := new(big.Int).Quo(new(big.Int).Mul(av.Num(), bv.Denom()), new(big.Int).Mul(av.Denom(), bv.Num()))
		q, r := a.QuoRem(b)
		if qwant.BitLen() > 512 {
			if !q.IsNaN() || !r.IsNaN() {
				t.Fatalf("overflow quotient %s / %s", a, b)
			}
			continue
		}
		if q.IsNaN() || r.IsNaN() {
			t.Fatalf("unexpected NaN %s / %s: %v %v", a, b, q.ErrorDetails(), r.ErrorDetails())
		}
		if ratFrom(q).Cmp(new(big.Rat).SetInt(qwant)) != 0 {
			t.Fatalf("quotient %s / %s: %s want %s", a, b, q, qwant)
		}
		remwant := new(big.Rat).Sub(av, new(big.Rat).Mul(bv, ratFrom(q)))
		rounded, ok := oracleAuto(remwant)
		if !ok || ratFrom(r).Cmp(rounded) != 0 {
			t.Fatalf("remainder %s / %s: %s want %s", a, b, r, remwant.RatString())
		}
	}
}

// Native Go fuzzing can replay or shrink failing exact arithmetic cases.
func FuzzPhaseAAdaptiveArithmetic(f *testing.F) {
	f.Add([]byte{1, 0, 0, 0, 0, 0, 0, 1}, []byte{2, 0, 0, 0, 0, 0, 0, 1}, uint8(154), uint8(0), true, false)
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, []byte{1}, uint8(0), uint8(154), false, true)
	f.Add([]byte{1}, []byte{1}, uint8(0), uint8(0), false, false)
	f.Fuzz(func(t *testing.T, ab, bb []byte, sa, sb uint8, na, nb bool) {
		if len(ab) > 64 {
			ab = ab[:64]
		}
		if len(bb) > 64 {
			bb = bb[:64]
		}
		ua, err := uint512.FromBigInt(new(big.Int).SetBytes(ab))
		if err != 0 {
			t.Fatal(err)
		}
		ub, err := uint512.FromBigInt(new(big.Int).SetBytes(bb))
		if err != 0 {
			t.Fatal(err)
		}
		a, b := New(ua, sa%155, na), New(ub, sb%155, nb)
		ar, br := ratFrom(a), ratFrom(b)
		phaseACheckAuto(t, "fuzz-add", a, b, a.Add(b), new(big.Rat).Add(ar, br))
		phaseACheckAuto(t, "fuzz-sub", a, b, a.Sub(b), new(big.Rat).Sub(ar, br))
		phaseACheckAuto(t, "fuzz-mul", a, b, a.Mul(b), new(big.Rat).Mul(ar, br))
	})
}

// An adversarial near-one power has its second binomial term just below
// the half-even boundary at scale 154. Higher terms have a rigorous upper
// bound less than the remaining distance, so its correctly rounded value
// can be proved without materializing an astronomical integer power.
func TestPhaseAPowBoundedNearRoundingMidpoint(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("large integer exponent requires 64-bit int")
	}
	exponent64 := int64(1000000000000000003)
	exponent := int(exponent64)
	n := big.NewInt(int64(exponent))
	one := pow10(int(MaxPrecision))
	pairs := new(big.Int).Mul(n, new(big.Int).Sub(n, big.NewInt(1)))
	// floor(sqrt(10^154 / (n*(n-1)))) minus a safety unit.
	k := new(big.Int).Sqrt(new(big.Int).Quo(new(big.Int).Set(one), pairs))
	k.Sub(k, big.NewInt(1))
	if k.Sign() <= 0 {
		t.Fatal("test construction invalid")
	}
	t2Num := new(big.Int).Mul(pairs, new(big.Int).Mul(k, k))
	delta := new(big.Int).Sub(one, t2Num)
	if delta.Sign() <= 0 {
		t.Fatal("the second term is not below the midpoint")
	}
	t3Num := new(big.Int).Mul(new(big.Int).Set(pairs), new(big.Int).Sub(n, big.NewInt(2)))
	t3Num.Mul(t3Num, new(big.Int).Exp(k, big.NewInt(3), nil))
	// 2*T3 < gap = 1/2 - T2 proves that the entire positive binomial
	// tail cannot cross the midpoint (successive ratios are tiny).
	lhs := new(big.Int).Mul(t3Num, big.NewInt(2))
	rhs := new(big.Int).Mul(new(big.Int).Mul(delta, one), big.NewInt(3))
	if lhs.Cmp(rhs) >= 0 {
		t.Fatal("binomial tail is not smaller than the midpoint gap")
	}
	// Avoid a test that can hide a loss of a decimal digit due to a trailing zero.
	if new(big.Int).Mod(new(big.Int).Mul(n, k), big.NewInt(10)).Sign() == 0 {
		k.Sub(k, big.NewInt(1))
	}
	baseCoefficient := new(big.Int).Add(new(big.Int).Set(one), k)
	base := phaseAFromParts(t, baseCoefficient, MaxPrecision, false)
	got := base.PowInt(exponent)
	expectedCoefficient := new(big.Int).Add(new(big.Int).Set(one), new(big.Int).Mul(n, k))
	expected := phaseAFromParts(t, expectedCoefficient, MaxPrecision, false)
	if got.IsNaN() || !got.Equal(expected) {
		t.Fatalf("near-boundary power incorrectly rounded: got %s (scale %d, error %v), expected %s (scale %d)", got, got.Precision(), got.ErrorDetails(), expected, expected.Precision())
	}
	// On the other side of the same midpoint, rounding must carry one
	// extra unit at scale 154, without dropping to a coarser scale.
	above := new(big.Int).Add(new(big.Int).Sqrt(new(big.Int).Quo(new(big.Int).Set(one), pairs)), big.NewInt(1))
	aboveT2 := new(big.Int).Mul(new(big.Int).Set(pairs), new(big.Int).Mul(above, above))
	if aboveT2.Cmp(one) <= 0 || aboveT2.Cmp(new(big.Int).Mul(one, big.NewInt(2))) >= 0 {
		t.Fatal("upper-side midpoint construction invalid")
	}
	aboveBase := phaseAFromParts(t, new(big.Int).Add(new(big.Int).Set(one), above), MaxPrecision, false)
	aboveWantCoef := new(big.Int).Add(new(big.Int).Set(one), new(big.Int).Mul(n, above))
	aboveWantCoef.Add(aboveWantCoef, big.NewInt(1))
	aboveWant := phaseAFromParts(t, aboveWantCoef, MaxPrecision, false)
	if got := aboveBase.PowInt(exponent); got.IsNaN() || !got.Equal(aboveWant) {
		t.Fatalf("above-midpoint power incorrectly rounded: got %s (scale %d), expected %s (scale %d)", got, got.Precision(), aboveWant, aboveWant.Precision())
	}
}

// Extremely large powers of a number one unit above 1 at scale 154 can be
// bounded by the first binomial terms; the second term is much less than
// half a unit at that scale even for the platform's largest exponent.
func TestPhaseAPowIntNearOneExponentExtremes(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("large integer exponent requires 64-bit int")
	}
	largest := int(^uint(0) >> 1)
	smallest := -largest - 1
	one := pow10(int(MaxPrecision))
	base := phaseAFromParts(t, new(big.Int).Add(new(big.Int).Set(one), big.NewInt(1)), MaxPrecision, false)
	inverseBase := phaseAFromParts(t, new(big.Int).Sub(new(big.Int).Set(one), big.NewInt(1)), MaxPrecision, false)
	expectedPositive := phaseAFromParts(t, new(big.Int).Add(new(big.Int).Set(one), big.NewInt(int64(largest))), MaxPrecision, false)
	expectedReciprocal := phaseAFromParts(t, new(big.Int).Sub(new(big.Int).Set(one), new(big.Int).Abs(big.NewInt(int64(smallest)))), MaxPrecision, false)
	if got := base.PowInt(largest); got.IsNaN() || !got.Equal(expectedPositive) {
		t.Fatalf("(1+1e-154)^MaxInt got %s, expected %s", got, expectedPositive)
	}
	if got := base.PowInt(smallest); got.IsNaN() || !got.Equal(expectedReciprocal) {
		t.Fatalf("(1+1e-154)^MinInt got %s, expected %s", got, expectedReciprocal)
	}
	expectedBelow := phaseAFromParts(t, new(big.Int).Sub(new(big.Int).Set(one), big.NewInt(int64(largest))), MaxPrecision, false)
	if got := inverseBase.PowInt(largest); got.IsNaN() || !got.Equal(expectedBelow) {
		t.Fatalf("(1-1e-154)^MaxInt got %s, expected %s", got, expectedBelow)
	}
}

func TestPhaseAPowBoundedInverseMidpoint(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("large integer exponent requires 64-bit int")
	}
	exponent64 := int64(1000000000000000003)
	n := big.NewInt(exponent64)
	one := pow10(int(MaxPrecision))
	nPlusOne := new(big.Int).Add(new(big.Int).Set(n), big.NewInt(1))
	pairs := new(big.Int).Mul(n, nPlusOne)
	center := new(big.Int).Sqrt(new(big.Int).Quo(new(big.Int).Set(one), pairs))
	for _, tc := range []struct {
		name    string
		offset  int64
		roundUp bool
	}{
		{"below", -1, false}, {"above", 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			k := new(big.Int).Add(new(big.Int).Set(center), big.NewInt(tc.offset))
			t2Num := new(big.Int).Mul(new(big.Int).Set(pairs), new(big.Int).Mul(k, k))
			delta := new(big.Int).Sub(new(big.Int).Set(one), t2Num)
			if tc.roundUp {
				delta.Neg(delta)
			}
			if delta.Sign() <= 0 {
				t.Fatal("wrong side of the midpoint")
			}
			t3Numerator := new(big.Int).Mul(new(big.Int).Set(pairs), new(big.Int).Add(new(big.Int).Set(n), big.NewInt(2)))
			t3Numerator.Mul(t3Numerator, new(big.Int).Exp(k, big.NewInt(3), nil))
			// The reciprocal binomial series is alternating with decreasing
			// terms, and |tail after T2| < T3 < gap to the midpoint.
			if new(big.Int).Mul(t3Numerator, big.NewInt(2)).Cmp(new(big.Int).Mul(new(big.Int).Mul(delta, one), big.NewInt(3))) >= 0 {
				t.Fatal("reciprocal tail bound does not prove rounding")
			}
			base := phaseAFromParts(t, new(big.Int).Add(new(big.Int).Set(one), k), MaxPrecision, false)
			wantCoef := new(big.Int).Sub(new(big.Int).Set(one), new(big.Int).Mul(n, k))
			if tc.roundUp {
				wantCoef.Add(wantCoef, big.NewInt(1))
			}
			want := phaseAFromParts(t, wantCoef, MaxPrecision, false)
			got := base.PowInt(-int(exponent64))
			if got.IsNaN() || !got.Equal(want) {
				t.Fatalf("(1+epsilon)^(-n) got %s at p=%d, want %s at p=%d", got, got.Precision(), want, want.Precision())
			}
		})
	}
}

func TestPhaseAMulAndRescaleStatus(t *testing.T) {
	rng := rand.New(rand.NewSource(0xC0FFEE23))
	maxCoef := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	for i := 0; i < 4000; i++ {
		a := phaseAFromParts(t, new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(1+rng.Intn(512)))), uint8(rng.Intn(155)), rng.Intn(2) == 1)
		b := phaseAFromParts(t, new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(1+rng.Intn(512)))), uint8(rng.Intn(155)), rng.Intn(2) == 1)
		p := uint8(rng.Intn(155))
		mode := RoundingMode(rng.Intn(7))
		mul, mulStatus := a.MulRoundWithStatus(b, p, mode)
		rescale, rescaleStatus := a.RescaleRoundWithStatus(p, mode)
		cases := []struct {
			name    string
			result  Optimus
			inexact bool
			value   *big.Rat
		}{
			{"mulround", mul, mulStatus, new(big.Rat).Mul(ratFrom(a), ratFrom(b))},
			{"rescale", rescale, rescaleStatus, ratFrom(a)},
		}
		for _, tc := range cases {
			expected, status := phaseARoundRat(tc.value, int(p), mode)
			if new(big.Int).Abs(new(big.Int).Set(expected)).Cmp(maxCoef) > 0 {
				if !tc.result.IsNaN() {
					t.Fatalf("%s: expected overflow a=%s b=%s p=%d got=%s", tc.name, a, b, p, tc.result)
				}
				continue
			}
			if tc.result.IsNaN() || tc.result.Precision() != p || tc.result.coefBig().Cmp(new(big.Int).Abs(expected)) != 0 || tc.result.IsNegative() != (expected.Sign() < 0) || tc.inexact != status {
				t.Fatalf("%s: a=%s b=%s p=%d mode=%d got=%s (inx=%t) want %s (inx=%t)", tc.name, a, b, p, mode, tc.result, tc.inexact, expected, status)
			}
		}
	}
}

// The integer-square-root oracle evaluates sqrt(a) at scale p using an
// integer floor and a separately justified rational midpoint inequality.
func phaseASqrtOracle(a *big.Rat, p uint8, mode RoundingMode) (*big.Int, bool) {
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(2*int64(p)), nil)
	num := new(big.Int).Mul(new(big.Int).Set(a.Num()), factor)
	den := a.Denom()
	floor := new(big.Int).Sqrt(new(big.Int).Quo(new(big.Int).Set(num), den))
	floorSquare := new(big.Int).Mul(new(big.Int).Set(floor), floor)
	inexact := new(big.Int).Mul(floorSquare, den).Cmp(num) != 0
	if !inexact {
		return floor, false
	}
	twiceFloorPlusOne := new(big.Int).Add(new(big.Int).Lsh(new(big.Int).Set(floor), 1), big.NewInt(1))
	left := new(big.Int).Mul(num, big.NewInt(4))
	right := new(big.Int).Mul(new(big.Int).Mul(twiceFloorPlusOne, twiceFloorPlusOne), den)
	comparison := left.Cmp(right)
	increment := false
	switch mode {
	case RoundingAwayFromZero, RoundingCeil:
		increment = true
	case RoundingHalfTowardZero:
		increment = comparison > 0
	case RoundingHalfAwayFromZero:
		increment = comparison >= 0
	case RoundingHalfEven:
		increment = comparison > 0 || comparison == 0 && floor.Bit(0) == 1
	}
	if increment {
		floor.Add(floor, big.NewInt(1))
	}
	return floor, true
}

func TestPhaseASqrtRoundWithIntegerOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(0x5A17))
	for i := 0; i < 2000; i++ {
		a := phaseAFromParts(t, new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(1+rng.Intn(512)))), uint8(rng.Intn(155)), false)
		p := uint8(rng.Intn(155))
		mode := RoundingMode(rng.Intn(7))
		expect, inexact := phaseASqrtOracle(ratFrom(a), p, mode)
		got, status := a.SqrtRoundWithStatus(p, mode)
		if expect.BitLen() > 512 {
			if !got.IsNaN() {
				t.Fatalf("sqrt expected overflow a=%s p=%d got=%s", a, p, got)
			}
			continue
		}
		if got.IsNaN() || got.Precision() != p || got.coefBig().Cmp(expect) != 0 || status != inexact {
			t.Fatalf("sqrt a=%s p=%d mode=%d got=%s inx=%t want %s inx=%t", a, p, mode, got, status, expect, inexact)
		}
	}
}

// The large-exponent midpoint case exercises interval refinement and ensures
// that a precision improvement is not accidentally optimized away.
func BenchmarkPhaseAPowNearMidpoint(b *testing.B) {
	if strconv.IntSize < 64 {
		b.Skip("requires a 64-bit int")
	}
	exponent64 := int64(1000000000000000003)
	n := big.NewInt(exponent64)
	one := pow10(int(MaxPrecision))
	pairs := new(big.Int).Mul(n, new(big.Int).Sub(n, big.NewInt(1)))
	k := new(big.Int).Sub(new(big.Int).Sqrt(new(big.Int).Quo(new(big.Int).Set(one), pairs)), big.NewInt(1))
	u, err := uint512.FromBigInt(new(big.Int).Add(one, k))
	if err != 0 {
		b.Fatal(err)
	}
	base := New(u, MaxPrecision, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = base.PowInt(int(exponent64))
	}
}
