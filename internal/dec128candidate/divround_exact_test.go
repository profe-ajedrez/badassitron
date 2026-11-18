package dec128candidate_test

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	candidate "github.com/profe-ajedrez/badassitron/internal/dec128candidate"
)

// This oracle deliberately uses math/big rather than either decimal engine.
// It covers 64-bit and full-width divisors, positive and negative scale
// alignment, all seven rounding modes and the inexact bit.
func TestM2DivRoundIndependentRational(t *testing.T) {
	r := rand.New(rand.NewSource(2026111707))
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	pows := make([]*big.Int, 39)
	pows[0] = big.NewInt(1)
	for i := 1; i < len(pows); i++ {
		pows[i] = new(big.Int).Mul(pows[i-1], big.NewInt(10))
	}
	type input struct{ a, b candidate.Dec128 }
	cases := []input{
		{candidate.FromString("1"), candidate.FromString("2")},
		{candidate.FromString("-1"), candidate.FromString("2")},
		{candidate.FromString("5"), candidate.FromString("2")},
		{candidate.FromString("-5"), candidate.FromString("2")},
		{candidate.FromString("1.0000000000000000000"), candidate.FromString("0.0000000000000000001")},
		{candidate.FromString(uint128.MaxUint128Str), candidate.FromString("1")},
		{candidate.FromString("0.0000000000000000001"), candidate.FromString(uint128.MaxUint128Str)},
	}
	for i := 0; i < 1500; i++ {
		makeOperand := func() candidate.Dec128 {
			c := uint128.Uint128{Lo: r.Uint64(), Hi: r.Uint64()}
			switch r.Intn(4) {
			case 0:
				c.Hi = 0
			case 1:
				c.Hi = 0
				c.Lo = uint64(r.Intn(10000) + 1)
			case 2:
				c.Hi = 0
				c.Lo = uint64(r.Intn(10000000) + 1)
			}
			return candidate.New(c, uint8(r.Intn(20)), r.Intn(2) == 0)
		}
		a, b := makeOperand(), makeOperand()
		if b.IsZero() {
			b = candidate.One
		}
		cases = append(cases, input{a, b})
	}
	for i, tc := range cases {
		for _, precision := range []uint8{0, 1, 2, 6, 10, 19} {
			shift := int(tc.b.Precision()) + int(precision) - int(tc.a.Precision())
			num, den := tc.a.Coefficient().BigInt(), tc.b.Coefficient().BigInt()
			if shift >= 0 {
				num.Mul(num, pows[shift])
			} else {
				den.Mul(den, pows[-shift])
			}
			q, rem := new(big.Int), new(big.Int)
			q.QuoRem(num, den, rem)
			negative := tc.a.StoredNegative() != tc.b.StoredNegative()
			inexact := rem.Sign() != 0
			for mode := candidate.RoundingTowardZero; mode <= candidate.RoundingHalfEven; mode++ {
				expected := new(big.Int).Set(q)
				if inexact {
					increment := false
					switch mode {
					case candidate.RoundingAwayFromZero:
						increment = true
					case candidate.RoundingFloor:
						increment = negative
					case candidate.RoundingCeil:
						increment = !negative
					case candidate.RoundingHalfTowardZero, candidate.RoundingHalfAwayFromZero, candidate.RoundingHalfEven:
						cmp := new(big.Int).Lsh(rem, 1).Cmp(den)
						increment = cmp > 0 || (cmp == 0 && (mode == candidate.RoundingHalfAwayFromZero || (mode == candidate.RoundingHalfEven && expected.Bit(0) == 1)))
					}
					if increment {
						expected.Add(expected, big.NewInt(1))
					}
				}
				got, gotInexact := tc.a.DivRoundWithStatus(tc.b, precision, mode)
				if expected.Cmp(max) > 0 {
					if !got.IsNaN() || got.ErrorDetails() == nil {
						t.Fatalf("case=%d p=%d mode=%d wanted overflow, got %s", i, precision, mode, got.StringFixed())
					}
					continue
				}
				if gotInexact != inexact {
					t.Fatalf("case=%d p=%d mode=%d %s/%s inexact=%t want=%t", i, precision, mode, tc.a.StringFixed(), tc.b.StringFixed(), gotInexact, inexact)
				}
				if got.IsNaN() || got.Coefficient().BigInt().Cmp(expected) != 0 || got.Precision() != precision {
					t.Fatalf("case=%d p=%d mode=%d %s/%s expected coef=%s scale=%d got=%s", i, precision, mode, tc.a.StringFixed(), tc.b.StringFixed(), expected, precision, got.StringFixed())
				}
				if expected.Sign() != 0 && got.IsNegative() != negative {
					t.Fatalf("case=%d p=%d mode=%d wrong sign", i, precision, mode)
				}
			}
		}
	}
}
