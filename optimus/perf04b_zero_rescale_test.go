package optimus

import (
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec256/uint256"
	"github.com/profe-ajedrez/badassitron/dec512"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
	"github.com/profe-ajedrez/badassitron/internal/decimalwide"
)

// perf04bOriginalRescale is the pre-PERF-04B general path. It intentionally
// does not contain either the same-scale or inline shortcuts. This permits
// independent, reproducible comparisons for zero upscaling and invalid input.
func perf04bOriginalRescale(d Optimus, scale uint8, mode RoundingMode) (Optimus, bool) {
	if d.IsNaN() {
		return d, false
	}
	if scale <= MaxPrecision && mode <= RoundingHalfEven {
		rescaled, inexact := dec512.New(d.Coefficient(), d.exp, d.neg).RescaleRoundWithStatus(scale, mode)
		if !rescaled.IsNaN() {
			return New(rescaled.Coefficient(), scale, rescaled.IsNegative()), inexact
		}
	}
	n, inexact := decimalwide.Quantize(d.core(), int(scale), 512, int(MaxPrecision), mode)
	return pack(n), inexact
}

func perf04bCompare(t *testing.T, d Optimus, scale uint8, mode RoundingMode) {
	t.Helper()
	input := d
	got, gotInexact := d.RescaleRoundWithStatus(scale, mode)
	want, wantInexact := perf04bOriginalRescale(d, scale, mode)
	if gotInexact != wantInexact || !reflect.DeepEqual(got, want) {
		t.Fatalf("input=%q exp=%d dest=%d mode=%d: got=%q inexact=%t err=%v; want=%q inexact=%t err=%v",
			d.StringFixed(), d.exp, scale, mode, got.StringFixed(), gotInexact,
			got.ErrorDetails(), want.StringFixed(), wantInexact, want.ErrorDetails())
	}
	if !reflect.DeepEqual(input, d) {
		t.Fatalf("source mutated during rescale: %q", d.StringFixed())
	}
	if got.IsNaN() {
		return
	}
	// The newly optimized upscaling path must retain the target precision,
	// canonical unsigned zero, and exactness without sharing mutable storage.
	if d.wide == nil && d.small.IsZero() && scale > d.exp && scale <= MaxPrecision && mode <= RoundingHalfEven {
		if !got.IsZero() || got.neg || got.exp != scale || got.wide != nil || gotInexact || got.StringFixed() != want.StringFixed() {
			t.Fatalf("invalid canonical zero at target scale: %#v", got)
		}
	}
}

func TestPerf04BZeroRescaleContracts(t *testing.T) {
	if got := unsafe.Sizeof(Optimus{}); got != 32 {
		t.Fatalf("Optimus size changed: %d", got)
	}
	modeList := []RoundingMode{
		RoundingTowardZero, RoundingAwayFromZero, RoundingFloor,
		RoundingCeil, RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven,
	}
	count := 0
	for exp := 0; exp <= int(MaxPrecision); exp++ {
		e := uint8(exp)
		// The last two variants are noncanonical internal states: signed
		// zero must canonicalize; indirect zero retains the general path.
		wideZero := uint512.Uint512{}
		zeros := []Optimus{
			NewFromUint128(uint128.Uint128{}, e, true),
			NewFromUint256(uint256.Uint256{}, e, true),
			New(uint512.Uint512{}, e, true),
			{exp: e, neg: true},
			{exp: e, neg: true, wide: &wideZero},
		}
		targets := []uint8{e, 0, 1, 2, 19, 20, 77, 153, 154, 155, 255}
		if exp < int(MaxPrecision) {
			targets = append(targets, e+1, MaxPrecision)
		}
		for _, d := range zeros {
			for _, p := range targets {
				for _, m := range modeList {
					// The reference represents the original general algorithm;
					// comparisons target upscaling, same-scale canonical values,
					// and invalid scale, not other already optimized paths.
					if p != d.exp && p < d.exp && p <= MaxPrecision {
						continue
					}
					if p == d.exp && (d.neg || d.wide != nil) {
						continue
					}
					perf04bCompare(t, d, p, m)
					count++
				}
			}
		}
	}
	t.Logf("checked %d canonical/legacy zero cases across seven rounding modes", count)
}

func TestPerf04BInvalidArgumentsAndNaN(t *testing.T) {
	zeros := []Optimus{Zero, FromString("-0.000"), NaN(errors.InvalidFormat)}
	for _, d := range zeros {
		for _, p := range []uint8{0, 2, 19, 154, 155, 255} {
			for _, m := range []RoundingMode{RoundingHalfEven, RoundingMode(7), RoundingMode(255)} {
				if p == d.exp && d.IsZero() && m <= RoundingHalfEven {
					continue
				}
				perf04bCompare(t, d, p, m)
			}
		}
	}
}

func TestPerf04BRandomZeroUpscaleDifferential(t *testing.T) {
	const seed int64 = 0x04b0cafe
	rng := rand.New(rand.NewSource(seed))
	count := 0
	for i := 0; i < 30000; i++ {
		exp := uint8(rng.Intn(int(MaxPrecision)))
		delta := uint8(1 + rng.Intn(int(MaxPrecision-exp)))
		mode := RoundingMode(rng.Intn(int(RoundingHalfEven) + 1))
		d := Optimus{exp: exp}
		perf04bCompare(t, d, exp+delta, mode)
		count++
	}
	t.Logf("seed %d: %d random exact zero upscales compared", seed, count)
}

func TestPerf04BConcurrentZeroUpscale(t *testing.T) {
	src := Zero
	var wg sync.WaitGroup
	errorsCh := make(chan string, 32)
	for worker := 0; worker < 32; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 2000; i++ {
				scale := uint8((i+worker)%int(MaxPrecision) + 1)
				got, inexact := src.RescaleRoundWithStatus(scale, RoundingHalfEven)
				if inexact || !got.IsZero() || got.exp != scale || got.neg || got.wide != nil {
					errorsCh <- fmt.Sprintf("worker=%d iter=%d result=%#v", worker, i, got)
					return
				}
			}
		}(worker)
	}
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		t.Error(err)
	}
	if !reflect.DeepEqual(src, Zero) {
		t.Fatal("shared operand changed")
	}
}

var perf04bBenchmarkSink Optimus

func BenchmarkPerf04BRescale(b *testing.B) {
	// Fixed operands avoid mixing parsing and arithmetic costs.
	for _, tc := range []struct {
		name  string
		value Optimus
		scale uint8
	}{
		{"zero_up2", Zero, 2},
		{"zero_up154", Zero, 154},
		{"zero_down20", NewFromUint128(uint128.Uint128{}, 50, false), 30},
		{"nonzero_up2", FromString("12345.678901"), 8},
		{"nonzero_down2", FromString("12345.678901"), 4},
		{"same_scale", FromString("12345.678901"), 6},
		{"wide_nonzero_up2", FromString("123456789012345678901234567890123456789012345678901234567890"), 2},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			var result Optimus
			for i := 0; i < b.N; i++ {
				result = tc.value.RescaleRound(tc.scale, RoundingHalfEven)
			}
			perf04bBenchmarkSink = result
		})
	}
}
