package optimus

import (
	"math/big"
	"math/rand"
	"strings"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

// A 512x512 product is allowed to exceed 512 bits internally, but an integer
// product cannot recover representability by removing fractional digits.
func TestPhaseB2IntegerProductOverflow(t *testing.T) {
	inputs := []struct{ a, b string }{
		{strings.Repeat("9", 154), strings.Repeat("9", 154)},
		{strings.Repeat("8", 110), strings.Repeat("7", 110)},
		{"-" + strings.Repeat("9", 150), strings.Repeat("8", 150)},
	}
	for _, pair := range inputs {
		a, b := FromString(pair.a), FromString(pair.b)
		if a.IsNaN() || b.IsNaN() {
			t.Fatalf("test input is not representable: %q %q", pair.a, pair.b)
		}
		got := a.Mul(b)
		if !got.IsNaN() || got.err != errors.Overflow {
			t.Fatalf("integer overflow: got %s, error %v", got, got.err)
		}
		phaseACheckAuto(t, "integer overflow", a, b, got, new(big.Rat).Mul(ratFrom(a), ratFrom(b)))
	}
	// The optimization must never reject a product which still fits 512 bits.
	for _, pair := range [][2]string{
		{strings.Repeat("9", 75), strings.Repeat("8", 75)},
		{strings.Repeat("9", 110), "10000000000"},
	} {
		a, b := FromString(pair[0]), FromString(pair[1])
		phaseACheckAuto(t, "integer fit", a, b, a.Mul(b), new(big.Rat).Mul(ratFrom(a), ratFrom(b)))
	}
}

func TestPhaseB2WideFallbackExactRationals(t *testing.T) {
	rng := rand.New(rand.NewSource(0xB2B2C512))
	for i := 0; i < 1200; i++ {
		a := randomPhaseBDecimal(rng)
		b := randomPhaseBDecimal(rng)
		if b.IsZero() {
			b = One
		}
		independentQuoRem(t, a, b)
	}
	// Stress the 512-bit upper limit at heterogeneous scales where the
	// aligned numerator may need more than 512 bits.
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	near := new(big.Int).Sub(max, big.NewInt(1234567))
	uMax, _ := uint512.FromBigInt(max)
	uNear, _ := uint512.FromBigInt(near)
	for _, scaleA := range []uint8{0, 1, 19, 76, 153, 154} {
		for _, scaleB := range []uint8{0, 2, 38, 77, 153, 154} {
			a := New(uMax, scaleA, scaleA&1 != 0)
			b := New(uNear, scaleB, scaleB&1 != 0)
			independentQuoRem(t, a, b)
			independentQuoRem(t, b, a)
		}
	}
}

// Differential fuzzing targets the full 1024-bit product without assuming that
// the coefficient-width check or any string representation is the oracle.
func FuzzPhaseB2Mul(f *testing.F) {
	f.Add([]byte{0, 0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	f.Add([]byte{154, 154, 0xff, 0xff, 0xff, 0xff, 0xee, 0xee, 0xee, 0xee})
	f.Add([]byte{1, 153, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) < 4 || len(input) > 130 {
			return
		}
		n := (len(input) - 2) / 2
		left := new(big.Int).SetBytes(input[2 : 2+n])
		right := new(big.Int).SetBytes(input[2+n:])
		if left.BitLen() > 512 || right.BitLen() > 512 {
			return
		}
		la, ea := uint512.FromBigInt(left)
		rb, eb := uint512.FromBigInt(right)
		if ea != 0 || eb != 0 {
			t.Fatal("invalid generated coefficient")
		}
		a := New(la, input[0]%155, input[0]&1 == 1)
		b := New(rb, input[1]%155, input[1]&1 == 1)
		phaseACheckAuto(t, "multiply fuzz", a, b, a.Mul(b), new(big.Rat).Mul(ratFrom(a), ratFrom(b)))
	})
}
