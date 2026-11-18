package dec128candidate_test

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	legacy "github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	candidate "github.com/profe-ajedrez/badassitron/internal/dec128candidate"
)

// Independent of the upstream and historical square-root implementation:
// a rounded root is exact iff its squared integer coefficient, after common
// power-of-ten alignment, equals the input coefficient. math/big is used
// only in this test oracle; production paths remain fixed-width.
func squareEqualsDecimal(input, root candidate.Dec128) bool {
	left := new(big.Int).Mul(root.Coefficient().BigInt(), root.Coefficient().BigInt())
	right := input.Coefficient().BigInt()
	diff := 2*int(root.Precision()) - int(input.Precision())
	ten := big.NewInt(10)
	if diff > 0 {
		right.Mul(right, new(big.Int).Exp(ten, big.NewInt(int64(diff)), nil))
	} else if diff < 0 {
		left.Mul(left, new(big.Int).Exp(ten, big.NewInt(int64(-diff)), nil))
	}
	return left.Cmp(right) == 0
}

func TestM6SqrtSinglePassExactnessIndependent(t *testing.T) {
	r := rand.New(rand.NewSource(2026111706))
	inputs := []candidate.Dec128{
		candidate.FromString("0.0000000000000000001"),
		candidate.FromString("340282366920938463463374607431768211455"),
		candidate.FromString("0.25"),
		candidate.FromString("1.00"),
		candidate.FromString("2"),
		candidate.FromString("9.0000000000000000000"),
	}
	for i := 0; i < 400; i++ {
		inputs = append(inputs, candidate.New(uint128.Uint128{Hi: r.Uint64(), Lo: r.Uint64()}, uint8(r.Intn(20)), false))
	}
	for _, v := range inputs {
		for _, p := range []uint8{0, 1, 2, 5, 9, 10, 15, 19} {
			for m := candidate.RoundingTowardZero; m <= candidate.RoundingHalfEven; m++ {
				root, inexact := v.SqrtRoundWithStatus(p, m)
				if root.IsNaN() { // Large values may overflow at requested scale.
					continue
				}
				expected := !squareEqualsDecimal(v, root)
				if inexact != expected {
					t.Fatalf("sqrt independent exactness input=%s p=%d mode=%d result=%s: got inexact=%t want=%t", v.StringFixed(), p, m, root.StringFixed(), inexact, expected)
				}
			}
		}
	}
}

func TestM6ScannerTypesAndDestinationAtomicity(t *testing.T) {
	accepted := []any{nil, "1234.500", []byte("-2.000"), int(0), int(-123), int64(math.MinInt64), float32(1.125), float64(-12.25), "", []byte{}}
	for _, src := range accepted {
		old, now := legacy.FromString("1.23"), candidate.FromString("1.23")
		err1, err2 := legacy.ScanInto(&old, src), candidate.ScanInto(&now, src)
		if (err1 == nil) != (err2 == nil) {
			t.Fatalf("ScanInto(%T): legacy err=%v candidate err=%v", src, err1, err2)
		}
		same(t, fmt.Sprintf("scan %T", src), old, now)
	}
	rejected := []any{"123X", []byte("1.2.3"), "12345678901234567890123456789012345678901", struct{}{}, false, math.NaN(), math.Inf(1)}
	for _, src := range rejected {
		old, now := legacy.FromString("1.23"), candidate.FromString("1.23")
		err1, err2 := legacy.ScanInto(&old, src), candidate.ScanInto(&now, src)
		if err1 == nil || err2 == nil || err1.Error() != err2.Error() {
			t.Fatalf("ScanInto(%T): legacy err=%v candidate err=%v", src, err1, err2)
		}
		same(t, fmt.Sprintf("failed scan %T", src), old, now)
	}
	if err := candidate.ScanInto(nil, "1"); err == nil {
		t.Fatal("expected nil receiver error")
	}
}
