package optimus

import (
	"math/big"
	"math/rand"
	"strings"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

// independentQuoRem uses rational arithmetic rather than the fixed-width
// division implementation, and independently checks both quotient and residue.
func independentQuoRem(t *testing.T, a, b Optimus) {
	t.Helper()
	if a.IsNaN() || b.IsNaN() || b.IsZero() {
		t.Fatal("oracle requires valid, nonzero divisor")
	}
	numerator, denominator := ratFrom(a), ratFrom(b)
	quotientRat := new(big.Rat).Quo(numerator, denominator)
	quotientInt := new(big.Int).Quo(quotientRat.Num(), quotientRat.Denom())
	quotientExpected := new(big.Rat).SetInt(quotientInt)
	remainderExpected := new(big.Rat).Sub(numerator, new(big.Rat).Mul(quotientExpected, denominator))

	quotient, remainder := a.QuoRem(b)
	onlyRemainder := a.Mod(b)
	if quotientInt.BitLen() > 512 {
		if !quotient.IsNaN() || !remainder.IsNaN() {
			t.Fatalf("quotient overflow a=%s b=%s q=%s r=%s", a, b, quotient, remainder)
		}
	} else {
		if quotient.IsNaN() || remainder.IsNaN() {
			t.Fatalf("unexpected error a=%s b=%s q=%s r=%s", a, b, quotient, remainder)
		}
		if ratFrom(quotient).Cmp(quotientExpected) != 0 {
			t.Fatalf("quotient mismatch a=%s b=%s got=%s want=%s", a, b, quotient, quotientExpected)
		}
		if ratFrom(remainder).Cmp(remainderExpected) != 0 {
			t.Fatalf("remainder mismatch a=%s b=%s got=%s want=%s", a, b, remainder, remainderExpected)
		}
		if remainder.Precision() != max(a.Precision(), b.Precision()) {
			t.Fatalf("remainder scale changed a=%s b=%s r=%s", a, b, remainder)
		}
	}
	if onlyRemainder.IsNaN() || ratFrom(onlyRemainder).Cmp(remainderExpected) != 0 {
		t.Fatalf("Mod mismatch a=%s b=%s got=%s want=%s", a, b, onlyRemainder, remainderExpected)
	}
	if onlyRemainder.Precision() != max(a.Precision(), b.Precision()) {
		t.Fatalf("Mod scale changed a=%s b=%s r=%s", a, b, onlyRemainder)
	}
}

func TestPhaseBDivisionDirected(t *testing.T) {
	cases := [][2]string{
		{"0.00", "3"}, {"3", "12"}, {"-12.5", "3"}, {"12.5", "-3"},
		{"-12.5", "-3"}, {"18446744073709551615", "0.00001"},
		{"340282366920938463463374607431768211455", "11"},
		{"340282366920938463463374607431768211455", "0.0001"},
		{"1.0000000000000000000000001", "1.0000000000000000000000002"},
		{"1", "0.000000000000000000000000000000000000000000000000000000000000001"},
		{"1", "0." + strings.Repeat("0", 153) + "1"},
	}
	for _, pair := range cases {
		a, b := FromString(pair[0]), FromString(pair[1])
		if a.IsNaN() || b.IsNaN() {
			t.Fatalf("invalid test input %q", pair)
		}
		if b.IsZero() {
			t.Fatalf("zero divisor parsed from %q", pair[1])
		}
		independentQuoRem(t, a, b)
	}
	for _, tc := range phaseBDivisionCases() {
		independentQuoRem(t, tc.numerator, tc.denominator)
	}
	if !FromInt(2).Mod(Zero).IsNaN() {
		t.Fatal("Mod zero divisor must be NaN")
	}
	q, r := FromInt(2).QuoRem(Zero)
	if !q.IsNaN() || !r.IsNaN() {
		t.Fatal("QuoRem zero divisor must return NaNs")
	}
	if !NaN(1).Mod(One).IsNaN() {
		t.Fatal("NaN must propagate")
	}
}

func randomPhaseBDecimal(rng *rand.Rand) Optimus {
	bitWidth := []int{1, 63, 64, 65, 127, 128, 129, 255, 256, 257, 511, 512}[rng.Intn(12)]
	var buf [64]byte
	for i := 0; i < (bitWidth+7)/8; i++ {
		buf[63-i] = byte(rng.Intn(256))
	}
	coefficient := new(big.Int).SetBytes(buf[:])
	coefficient.SetBit(coefficient, bitWidth-1, 1)
	fixed, err := uint512.FromBigInt(coefficient)
	if err != 0 {
		panic("test generator overflowed")
	}
	scale := []uint8{0, 1, 2, 19, 38, 76, 128, 153, 154}[rng.Intn(9)]
	return New(fixed, scale, rng.Intn(2) == 0)
}

func TestPhaseBDivisionDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(0xbada5517))
	for i := 0; i < 2400; i++ {
		a, b := randomPhaseBDecimal(rng), randomPhaseBDecimal(rng)
		independentQuoRem(t, a, b)
	}
}

func FuzzPhaseBQuoRem(f *testing.F) {
	f.Add([]byte{0, 154, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{154, 0, 255, 255, 255, 255, 0, 0, 0, 1})
	f.Add([]byte{38, 19, 0, 0, 1, 2, 3, 4, 5, 6})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) < 4 || len(input) > 132 {
			return
		}
		half := (len(input) - 2) / 2
		numerator := new(big.Int).SetBytes(input[2 : 2+half])
		denominator := new(big.Int).SetBytes(input[2+half:])
		if numerator.Sign() == 0 {
			numerator.SetUint64(1)
		}
		if denominator.Sign() == 0 {
			denominator.SetUint64(1)
		}
		if numerator.BitLen() > 512 || denominator.BitLen() > 512 {
			return
		}
		aFixed, ea := uint512.FromBigInt(numerator)
		bFixed, eb := uint512.FromBigInt(denominator)
		if ea != 0 || eb != 0 {
			t.Fatal("fuzz input conversion failed")
		}
		a := New(aFixed, input[0]%155, input[0]&1 != 0)
		b := New(bFixed, input[1]%155, input[1]&1 != 0)
		independentQuoRem(t, a, b)
	})
}

func TestPhaseBPowIntervalScale(t *testing.T) {
	base := FromString("1.01")
	power := new(big.Int).Exp(big.NewInt(101), big.NewInt(5000), nil)
	divisor := new(big.Int).Exp(big.NewInt(100), big.NewInt(5000), nil)
	exact := new(big.Rat).SetFrac(power, divisor)
	for _, tc := range []struct {
		exponent int
		expected *big.Rat
	}{
		{5000, exact},
		{-5000, new(big.Rat).Inv(exact)},
	} {
		got := base.PowInt(tc.exponent)
		want, ok := oracleAuto(tc.expected)
		if !ok || got.IsNaN() || ratFrom(got).Cmp(want) != 0 {
			t.Fatalf("PowInt(%d) got=%s want=%s, representable=%v", tc.exponent, got, want, ok)
		}
	}
}
