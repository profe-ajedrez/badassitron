package dec1024

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec1024/uint1024"
	"github.com/profe-ajedrez/badassitron/internal/decimalwide"
)

// Exercises exact multiplication followed by quantization, independently of
// the fixed-limb algorithm. Compare coefficient, scale, sign, inexact and err.
func TestMulRoundOptFullDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(2026111401))
	scales := []uint16{0, 1, 2, 8, 18, 19, 20, 37, 38, 77, 154, 307, 308}
	for i := 0; i < 600; i++ {
		a, b := randomDecimal(rng), randomDecimal(rng)
		if i%4 == 0 {
			a = New(uint1024.FromUint64(uint64(rng.Intn(1000000000))), scales[rng.Intn(len(scales))], rng.Intn(2) == 0)
		}
		if i%4 == 1 {
			b = New(uint1024.FromUint64(uint64(rng.Intn(1000000000))), scales[rng.Intn(len(scales))], rng.Intn(2) == 0)
		}
		target := scales[rng.Intn(len(scales))]
		if i%3 == 0 {
			target = uint16(rng.Intn(309))
		}
		aa, bb := referenceNumber(a), referenceNumber(b)
		product := new(big.Int).Mul(aa.Coef, bb.Coef)
		for mode := RoundingMode(0); mode <= RoundingHalfEven; mode++ {
			got, inexact := a.MulRoundWithStatus(b, target, mode)
			want, wantInexact := decimalwide.Quantize(decimalwide.New(product, aa.Scale+bb.Scale, aa.Neg != bb.Neg), int(target), 1024, int(MaxPrecision), mode)
			if got.err != want.Err || inexact != wantInexact {
				t.Fatalf("case %d mode %d: err/inexact got=(%v,%v) want=(%v,%v)", i, mode, got.err, inexact, want.Err, wantInexact)
			}
			if got.IsNaN() {
				continue
			}
			if got.exp != uint16(want.Scale) || got.neg != want.Neg || got.coef.BigInt().Cmp(want.Coef) != 0 {
				t.Fatalf("case %d mode %d target %d: representation got=(%v,%d,%v) want=(%v,%d,%v)", i, mode, target, got.coef, got.exp, got.neg, want.Coef, want.Scale, want.Neg)
			}
		}
	}
}

func TestMulRoundOptHalfStickyAndSigns(t *testing.T) {
	half := new(big.Int).Mul(big.NewInt(5), new(big.Int).Exp(big.NewInt(10), big.NewInt(19), nil))
	for _, delta := range []int64{-1, 0, 1} {
		v := new(big.Int).Add(half, big.NewInt(delta))
		coeff, errCode := uint1024.FromBigInt(v)
		if errCode != 0 {
			t.Fatal(errCode)
		}
		for _, negative := range []bool{false, true} {
			x := New(coeff, 20, negative)
			ref := referenceNumber(x)
			for mode := RoundingMode(0); mode <= RoundingHalfEven; mode++ {
				got, inexact := x.MulRoundWithStatus(One, 0, mode)
				product := new(big.Int).Set(ref.Coef)
				want, wantInexact := decimalwide.Quantize(decimalwide.New(product, 20, negative), 0, 1024, 308, mode)
				if got.err != want.Err || got.exp != uint16(want.Scale) || got.neg != want.Neg || got.coef.BigInt().Cmp(want.Coef) != 0 || inexact != wantInexact {
					t.Fatalf("half-sticky delta=%d neg=%v mode=%d got=%s ix=%v want=%v ix=%v", delta, negative, mode, got, inexact, want, wantInexact)
				}
			}
		}
	}
}

func TestMulRoundOptValidationAndNoHeap(t *testing.T) {
	x, y := FromString("12345.6789"), FromString("12.3456789")
	wide := New(uint1024.Max, 308, false)
	for name, pair := range map[string][2]Dec1024{"small": {x, y}, "wide": {wide, wide}} {
		if allocations := testing.AllocsPerRun(200, func() { _, _ = pair[0].MulRoundWithStatus(pair[1], 2, RoundingHalfEven) }); allocations != 0 {
			t.Errorf("%s: %g allocs/op", name, allocations)
		}
	}
	if got, ix := x.MulRoundWithStatus(y, 309, RoundingHalfEven); !got.IsNaN() || ix {
		t.Fatal("invalid scale", got, ix)
	}
	if got, ix := x.MulRoundWithStatus(y, 2, RoundingMode(7)); !got.IsNaN() || ix {
		t.Fatal("invalid mode", got, ix)
	}
}

// Validate the chunked decimal algorithm in isolation, including products of
// full 2048-bit width and decimal shifts beyond a single uint64 divisor.
func TestFixedRoundPower10FullRange(t *testing.T) {
	rng := rand.New(rand.NewSource(0x10242048))
	shifts := []int{1, 2, 9, 18, 19, 20, 21, 37, 38, 39, 57, 77, 154, 307, 308, 309, 615, 616}
	for i := 0; i < 450; i++ {
		var coef fixedNum
		for k := 0; k < 32; k++ {
			coef[k] = rng.Uint64()
		}
		if i%8 == 0 {
			for k := 2; k < 32; k++ {
				coef[k] = 0
			}
		}
		n := shifts[i%len(shifts)]
		bigCoef := new(big.Int)
		for k := len(coef) - 1; k >= 0; k-- {
			bigCoef.Lsh(bigCoef, 64)
			bigCoef.Or(bigCoef, new(big.Int).SetUint64(coef[k]))
		}
		for _, neg := range []bool{false, true} {
			for mode := RoundingMode(0); mode <= RoundingHalfEven; mode++ {
				got, inexact := fixedRoundPower10(coef, n, neg, mode)
				want, wantInexact := decimalwide.Quantize(decimalwide.New(bigCoef, n, neg), 0, 4096, 616, mode)
				result := new(big.Int)
				for k := len(got) - 1; k >= 0; k-- {
					result.Lsh(result, 64)
					result.Or(result, new(big.Int).SetUint64(got[k]))
				}
				if want.Err != 0 || result.Cmp(want.Coef) != 0 || inexact != wantInexact {
					t.Fatalf("case=%d shift=%d neg=%v mode=%d result=%v ix=%v want=%v ix=%v err=%v", i, n, neg, mode, result, inexact, want.Coef, wantInexact, want.Err)
				}
			}
		}
	}
}
