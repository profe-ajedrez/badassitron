package uint1024

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
)

func TestMulCarrySignificantLimbsAgainstBigInt(t *testing.T) {
	r := rand.New(rand.NewSource(20261114))
	for na := 0; na <= 16; na++ {
		for nb := 0; nb <= 16; nb++ {
			for pattern := 0; pattern < 5; pattern++ {
				var aw, bw [16]uint64
				for i := 0; i < na; i++ {
					switch pattern {
					case 0:
						aw[i] = ^uint64(0)
					case 1:
						aw[i] = 1
					case 2:
						aw[i] = r.Uint64()
					case 3:
						aw[i] = ^uint64(0) - uint64(i)
					case 4:
						aw[i] = uint64(i+1) << 32
					}
				}
				for i := 0; i < nb; i++ {
					switch pattern {
					case 0:
						bw[i] = ^uint64(0)
					case 1:
						bw[i] = 1
					case 2:
						bw[i] = r.Uint64()
					case 3:
						bw[i] = ^uint64(0) - uint64(i+17)
					case 4:
						bw[i] = uint64(i+3) << 32
					}
				}
				a, b := fromWords(aw), fromWords(bw)
				low, high := a.MulCarry(b)
				got := new(big.Int).Lsh(high.BigInt(), 1024)
				got.Add(got, low.BigInt())
				want := new(big.Int).Mul(a.BigInt(), b.BigInt())
				if got.Cmp(want) != 0 {
					t.Fatalf("MulCarry widths=(%d,%d) pattern=%d got=%s want=%s", na, nb, pattern, got, want)
				}
				narrowed, err := a.Mul(b)
				if want.BitLen() > 1024 {
					if err != errors.Overflow {
						t.Fatalf("Mul widths=(%d,%d) pattern=%d expected overflow, got %v", na, nb, pattern, err)
					}
				} else if err != errors.None || narrowed.BigInt().Cmp(want) != 0 {
					t.Fatalf("Mul widths=(%d,%d) pattern=%d mismatch err=%v", na, nb, pattern, err)
				}
			}
		}
	}
}

var benchmarkMulLow, benchmarkMulHigh Uint1024

// The width pair is part of the benchmark name, making small and fully
// occupied products directly comparable across the baseline and candidate.
func BenchmarkMulCarryLimbWidths(b *testing.B) {
	for _, tc := range []struct {
		name string
		a, b int
	}{
		{"1x1", 1, 1}, {"2x2", 2, 2}, {"4x4", 4, 4},
		{"8x8", 8, 8}, {"12x12", 12, 12}, {"16x16", 16, 16},
		{"1x16", 1, 16}, {"4x16", 4, 16},
	} {
		var aw, bw [16]uint64
		for i := 0; i < tc.a; i++ {
			aw[i] = ^uint64(0) - uint64(i)
		}
		for i := 0; i < tc.b; i++ {
			bw[i] = ^uint64(0) - uint64(i+3)
		}
		a, c := fromWords(aw), fromWords(bw)
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmarkMulLow, benchmarkMulHigh = a.MulCarry(c)
			}
		})
	}
}

func TestMulCarryNoHeap(t *testing.T) {
	small := FromUint64(10101)
	large := Max
	for name, operands := range map[string][2]Uint1024{
		"small": {small, small}, "wide": {large, large}, "mixed": {small, large}, "zero": {Zero, large},
	} {
		if got := testing.AllocsPerRun(500, func() { benchmarkMulLow, benchmarkMulHigh = operands[0].MulCarry(operands[1]) }); got != 0 {
			t.Errorf("%s: %g allocations", name, got)
		}
	}
}
