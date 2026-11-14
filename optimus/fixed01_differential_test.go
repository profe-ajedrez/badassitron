package optimus

import (
	"fmt"
	"math/big"
	"math/rand"
	"runtime"
	"sync"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

func fixed01Compare(t *testing.T, label string, a, b Optimus, p uint8, mode RoundingMode) {
	t.Helper()
	got, gotInexact := a.MulRoundWithStatus(b, p, mode)
	want, wantInexact := a.mulRoundLegacy(b, p, mode)
	if gotInexact != wantInexact || got.Precision() != want.Precision() || got.IsNaN() != want.IsNaN() ||
		got.IsNegative() != want.IsNegative() || got.Coefficient() != want.Coefficient() ||
		fmt.Sprint(got.ErrorDetails()) != fmt.Sprint(want.ErrorDetails()) {
		t.Fatalf("%s a=%s b=%s p=%d mode=%d: fixed=%s (exp=%d inexact=%v err=%v) legacy=%s (exp=%d inexact=%v err=%v)",
			label, a.StringFixed(), b.StringFixed(), p, mode,
			got.StringFixed(), got.Precision(), gotInexact, got.ErrorDetails(),
			want.StringFixed(), want.Precision(), wantInexact, want.ErrorDetails())
	}
}

func TestFixed01Differential(t *testing.T) {
	rng := rand.New(rand.NewSource(0xBAAD512))
	powers := []int{0, 1, 63, 64, 65, 127, 128, 129, 255, 256, 257, 511, 512}
	scales := []uint8{0, 1, 2, 8, 18, 19, 38, 77, 153, 154}
	values := []Optimus{Zero, One, NegativeOne, FromString("2.5"), FromString("-3.5"),
		FromString("0.00000001"), NaN(errors.InvalidFormat), NaN(errors.DivisionByZero)}
	for _, width := range powers {
		for _, delta := range []int64{-1, 0, 1} {
			c := new(big.Int).Lsh(big.NewInt(1), uint(width))
			c.Add(c, big.NewInt(delta))
			if c.Sign() < 0 || c.BitLen() > 512 {
				continue
			}
			u, e := uint512.FromBigInt(c)
			if e != errors.None {
				t.Fatal(e)
			}
			for _, scale := range []uint8{0, 2, 154} {
				values = append(values, New(u, scale, delta == -1))
			}
		}
	}
	for i := 0; i < 2500; i++ {
		a, b := values[rng.Intn(len(values))], values[rng.Intn(len(values))]
		if i%4 == 0 {
			// Random 512-bit values include every coefficient tier.
			var c big.Int
			c.SetUint64(rng.Uint64())
			c.Lsh(&c, uint(rng.Intn(448)))
			u, e := uint512.FromBigInt(&c)
			if e != errors.None {
				t.Fatal(e)
			}
			a = New(u, scales[rng.Intn(len(scales))], rng.Intn(2) == 0)
		}
		p := scales[rng.Intn(len(scales))]
		m := RoundingMode(rng.Intn(7))
		fixed01Compare(t, fmt.Sprint(i), a, b, p, m)
	}
	for _, p := range []uint8{0, 2, 154, 155, 255} {
		for _, m := range []RoundingMode{RoundingTowardZero, RoundingAwayFromZero,
			RoundingFloor, RoundingCeil, RoundingHalfTowardZero, RoundingHalfAwayFromZero,
			RoundingHalfEven, RoundingMode(255)} {
			for _, a := range values[:8] {
				for _, b := range values[:8] {
					fixed01Compare(t, "boundary", a, b, p, m)
				}
			}
		}
	}
}

func TestFixed01SharedOperandsConcurrent(t *testing.T) {
	a := FromString("0.999999999999999999999999999999999999999999999999999999999999999999")
	b := FromString("0.888888888888888888888888888888888888888888888888888888888888888888")
	expected, status := a.MulRoundWithStatus(b, 19, RoundingHalfEven)
	if expected.IsNaN() {
		t.Fatal(expected.ErrorDetails())
	}
	var wg sync.WaitGroup
	for j := 0; j < min(runtime.GOMAXPROCS(0), 8); j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 250; i++ {
				got, inx := a.MulRoundWithStatus(b, 19, RoundingHalfEven)
				if got.Coefficient() != expected.Coefficient() || got.Precision() != expected.Precision() || inx != status {
					t.Error("concurrent result differs")
					return
				}
			}
		}()
	}
	wg.Wait()
}
