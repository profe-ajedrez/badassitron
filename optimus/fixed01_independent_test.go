package optimus

import (
	"math/big"
	"math/rand"
	"testing"
	"unsafe"

	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

func TestFixed01IndependentRationalOracle(t *testing.T) {
	if unsafe.Sizeof(Optimus{}) != 32 {
		t.Fatalf("Optimus layout changed: %d bytes", unsafe.Sizeof(Optimus{}))
	}
	rng := rand.New(rand.NewSource(0xF1ED01))
	scales := []uint8{0, 1, 2, 8, 19, 38, 77, 153, 154}
	widths := []int{0, 1, 63, 64, 65, 127, 128, 129, 255, 256, 257, 383, 511, 512}
	for i := 0; i < 750; i++ {
		var ab, bb big.Int
		for j, x := range []*big.Int{&ab, &bb} {
			w := widths[rng.Intn(len(widths))]
			if w == 0 {
				continue
			}
			var bytes [64]byte
			for k := range bytes {
				bytes[k] = byte(rng.Intn(256))
			}
			x.SetBytes(bytes[:])
			x.Rsh(x, uint(512-w))
			if x.Sign() == 0 {
				x.SetInt64(1)
			}
			if j == 1 && i%10 == 0 {
				x.SetInt64(5)
			}
		}
		ac, err := uint512.FromBigInt(&ab)
		if err != 0 {
			t.Fatal(err)
		}
		bc, err := uint512.FromBigInt(&bb)
		if err != 0 {
			t.Fatal(err)
		}
		a := New(ac, scales[rng.Intn(len(scales))], rng.Intn(2) == 0)
		b := New(bc, scales[rng.Intn(len(scales))], rng.Intn(2) == 0)
		target := scales[rng.Intn(len(scales))]
		product := new(big.Rat).Mul(phaseCPrecisionRat(a), phaseCPrecisionRat(b))
		for mode := RoundingTowardZero; mode <= RoundingHalfEven; mode++ {
			want, wantInexact := phaseCPrecisionQuantize(product, target, mode)
			got, inexact := a.MulRoundWithStatus(b, target, mode)
			if inexact != wantInexact {
				t.Fatalf("case %d mode %d status %t != %t", i, mode, inexact, wantInexact)
			}
			if new(big.Int).Abs(new(big.Int).Set(want)).BitLen() > 512 {
				if !got.IsNaN() {
					t.Fatalf("case %d mode %d expected overflow", i, mode)
				}
				continue
			}
			if got.IsNaN() || got.Precision() != target || got.IsNegative() != (want.Sign() < 0) || got.Coefficient().BigInt().Cmp(new(big.Int).Abs(want)) != 0 {
				t.Fatalf("case %d mode %d unexpected output %s exp=%d want coef=%s target=%d", i, mode, got.StringFixed(), got.Precision(), want.String(), target)
			}
		}
	}
}

func TestFixed01WideResultIndependence(t *testing.T) {
	a := FromString("123456789012345678901234567890123456789012345678901234567890")
	first, inexact1 := a.MulRoundWithStatus(FromInt(2), 0, RoundingHalfEven)
	second, inexact2 := a.MulRoundWithStatus(FromInt(2), 0, RoundingHalfEven)
	if first.IsNaN() || second.IsNaN() || inexact1 || inexact2 || first.wide == nil || second.wide == nil {
		t.Fatalf("unexpected wide results %s, %s", first, second)
	}
	if first.wide == second.wide {
		t.Fatal("distinct wide results share mutable coefficient storage")
	}
	if first.Coefficient() != second.Coefficient() || first.Precision() != second.Precision() {
		t.Fatal("independent results have different values")
	}
}
