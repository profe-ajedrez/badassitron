package optimus

import (
	"encoding/json"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec512"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
	"github.com/profe-ajedrez/badassitron/internal/decimalwide"
)

// Retains the exact FIXED-02B algorithm as a test-only reference.
func fixed03bOriginalRescale(d Optimus, p uint8, m RoundingMode) (Optimus, bool) {
	if d.IsNaN() {
		return d, false
	}
	if p <= MaxPrecision && m <= RoundingHalfEven {
		result, inexact := dec512.New(d.Coefficient(), d.exp, d.neg).RescaleRoundWithStatus(p, m)
		if !result.IsNaN() {
			return New(result.Coefficient(), p, result.IsNegative()), inexact
		}
	}
	r, inexact := decimalwide.Quantize(d.core(), int(p), 512, int(MaxPrecision), m)
	return pack(r), inexact
}

// Independent mathematical oracle: does not call Optimus/Dec512 rescaling,
// uint128.QuoRem64, or decimalwide.Quantize. The input is already an exact
// coefficient/scale pair, and all computation is done with big.Int.
func fixed03bBigOracle(d Optimus, p uint8, mode RoundingMode) (Optimus, bool) {
	if d.IsNaN() {
		return d, false
	}
	if p > MaxPrecision {
		return NaN(errors.PrecisionOutOfRange), false
	}
	if mode > RoundingHalfEven {
		return NaN(errors.InvalidFormat), false
	}

	c := d.Coefficient().BigInt()
	if p >= d.exp {
		e := int(p - d.exp)
		factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(e)), nil)
		c.Mul(c, factor)
		if c.BitLen() > 512 {
			return NaN(errors.Overflow), false
		}
		v, err := uint512.FromBigInt(c)
		if err != errors.None {
			return NaN(err), false
		}
		return New(v, p, d.neg), false
	}
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(d.exp-p)), nil)
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(c, factor, r)
	inexact := r.Sign() != 0
	if inexact {
		increment := false
		switch mode {
		case RoundingAwayFromZero:
			increment = true
		case RoundingFloor:
			increment = d.neg
		case RoundingCeil:
			increment = !d.neg
		case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
			cmp := new(big.Int).Lsh(r, 1).Cmp(factor)
			increment = cmp > 0 || (cmp == 0 &&
				(mode == RoundingHalfAwayFromZero ||
					(mode == RoundingHalfEven && q.Bit(0) != 0)))
		}
		if increment {
			q.Add(q, big.NewInt(1))
		}
	}
	if q.BitLen() > 512 {
		return NaN(errors.Overflow), inexact
	}
	v, err := uint512.FromBigInt(q)
	if err != errors.None {
		return NaN(err), inexact
	}
	return New(v, p, d.neg), inexact
}

func fixed03bAssertParity(t *testing.T, d Optimus, p uint8, m RoundingMode) {
	t.Helper()
	got, inexact := d.RescaleRoundWithStatus(p, m)
	original, oldInexact := fixed03bOriginalRescale(d, p, m)
	independent, expectedInexact := fixed03bBigOracle(d, p, m)
	equal := func(a, b Optimus) bool {
		return a.err == b.err && a.exp == b.exp && a.neg == b.neg && a.Coefficient() == b.Coefficient()
	}
	if !equal(got, original) || inexact != oldInexact || !equal(got, independent) || inexact != expectedInexact {
		t.Fatalf("input=%s input-scale=%d requested=%d mode=%d: got=%s (%v, %v), baseline=%s (%v, %v), oracle=%s (%v, %v)",
			d.StringFixed(), d.exp, p, m, got.StringFixed(), got.err, inexact,
			original.StringFixed(), original.err, oldInexact, independent.StringFixed(), independent.err, expectedInexact)
	}
}

func TestFixed03BSizeAndExplicitCases(t *testing.T) {
	if size := unsafe.Sizeof(Optimus{}); size != 32 {
		t.Fatalf("Optimus: %d bytes, expected 32", size)
	}
	for _, s := range []string{"0", "0.000000000", "-0.000", "1", "-1", "2.5", "-2.5",
		"2.5000000000000000000", "-2.5000000000000000001", "0.0000000000000000005",
		"-0.0000000000000000005", "123456789012345678901234567890.123456789",
		"0." + strings.Repeat("7", 70), "0." + strings.Repeat("9", 150),
		"1." + strings.Repeat("0", 153) + "1"} {
		d := FromString(s)
		if d.IsNaN() {
			t.Fatalf("unexpected invalid input: %s", s)
		}
		for _, p := range []uint8{0, 1, 2, 6, 19, 20, 38, 76, 100, 153, 154, 155, 255, d.Precision()} {
			for m := RoundingMode(0); m <= RoundingHalfEven; m++ {
				fixed03bAssertParity(t, d, p, m)
			}
			fixed03bAssertParity(t, d, p, RoundingMode(255))
		}
	}
	for _, d := range []Optimus{NaN(errors.NotANumber), NaN(errors.Overflow)} {
		for _, p := range []uint8{0, 154, 255} {
			fixed03bAssertParity(t, d, p, RoundingHalfEven)
			fixed03bAssertParity(t, d, p, RoundingMode(255))
		}
	}
}

func TestFixed03BSeeded120000DifferentialAndOracle(t *testing.T) {
	const seed = int64(2026111403)
	rng := rand.New(rand.NewSource(seed))
	scales := [...]uint8{0, 1, 2, 3, 6, 19, 20, 38, 76, 100, 135, 154}
	for i := 0; i < 120000; i++ {
		exp := scales[rng.Intn(len(scales))]
		var words [16]uint64
		limbs := 1 + rng.Intn(8)
		for j := 0; j < limbs; j++ {
			words[j] = rng.Uint64()
		}
		if i%137 == 0 {
			for j := range words {
				words[j] = 0
			}
		}
		d := fixed01Pack(&words, limbs, exp, rng.Intn(2) == 0)
		if i%1031 == 0 {
			d = NaN(errors.NotANumber)
		}
		p := scales[rng.Intn(len(scales))]
		if i%3 == 0 {
			p = exp
		}
		if i%5 == 0 && exp > 0 {
			p = exp - uint8(1+rng.Intn(min(19, int(exp))))
		}
		if i%139 == 0 {
			p = 155
		}
		if i%149 == 0 {
			p = 255
		}
		mode := RoundingMode(rng.Intn(7))
		if i%173 == 0 {
			mode = RoundingMode(255)
		}
		fixed03bAssertParity(t, d, p, mode)
	}
}

func TestFixed03BExactTiesAndNeighbors(t *testing.T) {
	// Every short divisor, both quotient parities, positive/negative,
	// and exact midpoint +/-1: includes the critical 10^19 boundary.
	for delta := 1; delta <= 20; delta++ {
		den := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(delta)), nil)
		half := new(big.Int).Rsh(new(big.Int).Set(den), 1)
		for _, q64 := range []int64{0, 1, 2, 3, 1000000001} {
			for _, offset := range []int64{-1, 0, 1} {
				coef := new(big.Int).Mul(big.NewInt(q64), den)
				coef.Add(coef, new(big.Int).Add(half, big.NewInt(offset)))
				if coef.Sign() < 0 {
					continue
				}
				wide, err := uint512.FromBigInt(coef)
				if err != errors.None {
					t.Fatal(err)
				}
				for _, neg := range []bool{false, true} {
					d := New(wide, uint8(delta), neg)
					for m := RoundingMode(0); m <= RoundingHalfEven; m++ {
						fixed03bAssertParity(t, d, 0, m)
					}
				}
			}
		}
	}
}

func TestFixed03BSameScaleImmutabilitySerializationAndConcurrency(t *testing.T) {
	cases := []Optimus{FromString("-123.4500"), FromString("0." + strings.Repeat("1", 150)), FromString("0.000000"), NaN(errors.NotANumber)}
	for _, input := range cases {
		original := input.StringFixed()
		beforeJSON, _ := json.Marshal(input)
		got, status := input.RescaleRoundWithStatus(input.Precision(), RoundingHalfEven)
		if status || got.StringFixed() != original {
			t.Fatalf("same-scale changed value or status: %s", original)
		}
		j, _ := json.Marshal(got)
		if string(j) != string(beforeJSON) || got.Canonical().String() != input.Canonical().String() {
			t.Fatalf("serialization or canonicalization changed: %s", original)
		}
		const n = 16
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 200; j++ {
					v, _ := input.RescaleRoundWithStatus(input.Precision(), RoundingHalfEven)
					if v.StringFixed() != original {
						t.Errorf("concurrent read changed value")
						return
					}
				}
			}()
		}
		wg.Wait()
		if input.StringFixed() != original {
			t.Fatal("input mutated")
		}
	}
}
