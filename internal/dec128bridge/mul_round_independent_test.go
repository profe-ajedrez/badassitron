package dec128bridge

import (
	"math/big"
	"math/rand"
	"testing"

	legacy "github.com/profe-ajedrez/badassitron/dec128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// Independent quantization oracle: math/big supplies the exact product,
// with the rounding decision expressed directly in terms of quotient/remainder.
func TestMulRoundIndependentAllModes(t *testing.T) {
	rng := rand.New(rand.NewSource(17002611))
	checked := 0
	ten := big.NewInt(10)
	for i := 0; i < 2000; i++ {
		a := legacy.New(lu.Uint128{Lo: rng.Uint64() % 1000000000000}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		b := legacy.New(lu.Uint128{Lo: rng.Uint64() % 1000000000000}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		p := uint8(rng.Intn(20))
		ac, bc := a.Coefficient(), b.Coefficient()
		num := new(big.Int).Mul(new(big.Int).SetUint64(ac.Lo), new(big.Int).SetUint64(bc.Lo))
		exp := int(a.Precision()) + int(b.Precision()) - int(p)
		den := big.NewInt(1)
		if exp >= 0 {
			den.Exp(ten, big.NewInt(int64(exp)), nil)
		} else {
			num.Mul(num, new(big.Int).Exp(ten, big.NewInt(int64(-exp)), nil))
		}
		q, r := new(big.Int), new(big.Int)
		q.QuoRem(num, den, r)
		neg := a.IsNegative() != b.IsNegative()
		for mode := legacy.RoundingMode(0); mode <= legacy.RoundingHalfEven; mode++ {
			got, inexact := MulRoundWithStatus(a, b, p, mode)
			if got.IsNaN() {
				if q.BitLen() > 128 {
					continue
				}
				t.Fatalf("unexpected NaN i=%d mode=%d: %v", i, mode, got.ErrorDetails())
			}
			want := new(big.Int).Set(q)
			increment := false
			if r.Sign() != 0 {
				switch mode {
				case legacy.RoundingAwayFromZero:
					increment = true
				case legacy.RoundingFloor:
					increment = neg
				case legacy.RoundingCeil:
					increment = !neg
				case legacy.RoundingHalfTowardZero, legacy.RoundingHalfAwayFromZero, legacy.RoundingHalfEven:
					cmp := new(big.Int).Mul(r, big.NewInt(2)).Cmp(den)
					increment = cmp > 0 || cmp == 0 && (mode == legacy.RoundingHalfAwayFromZero || mode == legacy.RoundingHalfEven && q.Bit(0) == 1)
				}
			}
			if increment {
				want.Add(want, big.NewInt(1))
			}
			gc := got.Coefficient()
			gotCoef := new(big.Int).SetUint64(gc.Hi)
			gotCoef.Lsh(gotCoef, 64)
			gotCoef.Add(gotCoef, new(big.Int).SetUint64(gc.Lo))
			if gotCoef.Cmp(want) != 0 || got.Precision() != p || inexact != (r.Sign() != 0) || got.IsNegative() != (neg && want.Sign() != 0) {
				t.Fatalf("i=%d mode=%d expected=%s actual=%s scale=%d inexact=%t", i, mode, want, got, p, inexact)
			}
			checked++
		}
	}
	t.Logf("independent product+round checks: %d", checked)
}
