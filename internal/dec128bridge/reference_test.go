package dec128bridge

import (
	"math/big"
	"math/rand"
	"testing"

	legacy "github.com/profe-ajedrez/badassitron/dec128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// This math/big oracle checks finite products independently of both decimal
// implementations. It does not assume the old code is mathematically right.
func TestMulStrictFiniteAgainstBigInt(t *testing.T) {
	rng := rand.New(rand.NewSource(20261118))
	for i := 0; i < 5000; i++ {
		a := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: uint64(rng.Intn(20))}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		b := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: uint64(rng.Intn(20))}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		got := MulStrict(a, b)
		if got.IsNaN() {
			continue
		}
		exact := new(big.Int).Mul(a.Coefficient().BigInt(), b.Coefficient().BigInt())
		if a.IsNegative() != b.IsNegative() {
			exact.Neg(exact)
		}
		output := new(big.Int).Set(got.Coefficient().BigInt())
		if got.IsNegative() {
			output.Neg(output)
		}
		// Bring coefficients to the same decimal scale for an exact equality.
		leftScale := int(a.Precision()) + int(b.Precision())
		rightScale := int(got.Precision())
		if leftScale > rightScale {
			output.Mul(output, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(leftScale-rightScale)), nil))
		} else if rightScale > leftScale {
			exact.Mul(exact, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(rightScale-leftScale)), nil))
		}
		if exact.Cmp(output) != 0 {
			t.Fatalf("product mismatch: a=%s b=%s got=%s", a.StringFixed(), b.StringFixed(), got.StringFixed())
		}
	}
}
