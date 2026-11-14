package optimus

import (
	"encoding/json"
	"fmt"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
)

// concurrentFingerprint compares the representation as well as the value:
// a race that changes a shared coefficient must not be hidden by Equal.
func concurrentFingerprint(d Optimus) string {
	return fmt.Sprintf("%s/%d/%t/%d/%d", d.Coefficient(), d.exp, d.neg, d.err, d.Width())
}

func TestPhaseCSharedOperandsAreImmutableUnderConcurrency(t *testing.T) {
	wide512 := FromString(strings.Repeat("7", 123) + ".25001")
	wide256 := FromString(strings.Repeat("8", 56) + ".125")
	small := FromString("1.0000125")
	negative := FromString("-1234.50000")
	zero := Zero
	invalid := NaN(errors.NotANumber)
	inputs := []Optimus{wide512, wide256, small, negative, zero, invalid}
	before := make([]string, len(inputs))
	for i, v := range inputs {
		before[i] = concurrentFingerprint(v)
	}

	operations := []struct {
		name string
		run  func() Optimus
	}{
		{"add-wide", func() Optimus { return wide512.Add(wide256) }},
		{"sub-wide", func() Optimus { return wide512.Sub(wide256) }},
		{"mul-wide", func() Optimus { return wide256.Mul(small) }},
		{"div-wide", func() Optimus { return wide512.DivRound(wide256, 22, RoundingHalfEven) }},
		{"mod-wide", func() Optimus { return wide512.Mod(wide256) }},
		{"quotient-wide", func() Optimus { q, _ := wide512.QuoRem(wide256); return q }},
		{"remainder-wide", func() Optimus { _, r := wide512.QuoRem(wide256); return r }},
		{"power", func() Optimus { return small.PowInt(-37) }},
		{"sqrt-wide", func() Optimus { return wide256.SqrtRound(14, RoundingHalfEven) }},
		{"canonical", func() Optimus { return negative.Canonical() }},
		{"bank-round", func() Optimus { return negative.RoundBank(2) }},
		{"rescale", func() Optimus { return wide256.RescaleRound(2, RoundingHalfEven) }},
		{"nan-propagation", func() Optimus { return invalid.Add(wide512) }},
		{"zero", func() Optimus { return zero.Sub(zero) }},
	}
	want := make([]string, len(operations))
	for i, op := range operations {
		want[i] = concurrentFingerprint(op.run())
	}

	const goroutines, iterations = 12, 60
	var wg sync.WaitGroup
	failures := make(chan string, goroutines)
	start := make(chan struct{})
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < iterations; i++ {
				for j, op := range operations {
					if got := concurrentFingerprint(op.run()); got != want[j] {
						failures <- fmt.Sprintf("%s: got %s, want %s", op.name, got, want[j])
						return
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(failures)
	for failure := range failures {
		t.Error(failure)
	}
	for i, v := range inputs {
		if got := concurrentFingerprint(v); got != before[i] {
			t.Errorf("input %d changed: %s instead of %s", i, got, before[i])
		}
	}
	if wide512.wide == nil || wide256.wide == nil {
		t.Fatal("wide test fixture did not exercise indirect coefficients")
	}
	// The immutable representation permits a shallow copy to share storage.
	if wide512.Copy().wide != wide512.wide {
		t.Fatal("Copy unexpectedly changed the aliasing contract")
	}
}

func TestPhaseCWideBigIntConversionDoesNotExposeStorage(t *testing.T) {
	source := FromString(strings.Repeat("9", 135) + ".125")
	original := concurrentFingerprint(source)
	copyOfSource := source.Copy()
	converted := source.coefBig()
	converted.SetInt64(-1)
	core := source.core()
	core.Coef.SetInt64(0)
	coefficient := source.Coefficient()
	coefficient.Hi = coefficient.Lo
	if concurrentFingerprint(source) != original || concurrentFingerprint(copyOfSource) != original {
		t.Fatal("a mutable conversion changed a shared Optimus coefficient")
	}
	if coefficient == source.Coefficient() {
		t.Fatal("the copy of the coefficient unexpectedly mutated the source")
	}
}

func TestPhaseCConcurrentDefaultPrecisionProducesOnlyValidSnapshots(t *testing.T) {
	old := uint8(defaultPrecision.Load())
	defer SetDefaultPrecision(old)
	precisions := []uint8{2, 9, 19, 37, 78, 150}
	value := FromString("1.3")
	divisionNumerator, divisionDenominator := FromInt(1), FromInt(7)
	squareRoot := FromInt(2)
	operations := []struct {
		name string
		run  func() Optimus
	}{
		{"division", func() Optimus { return divisionNumerator.Div(divisionDenominator) }},
		{"square-root", func() Optimus { return squareRoot.Sqrt() }},
		{"negative-power", func() Optimus { return value.PowInt(-5) }},
	}
	allowed := make([]map[string]struct{}, len(operations))
	for j := range allowed {
		allowed[j] = make(map[string]struct{})
	}
	for _, precision := range precisions {
		SetDefaultPrecision(precision)
		for j, operation := range operations {
			allowed[j][concurrentFingerprint(operation.run())] = struct{}{}
		}
	}

	const readers, iterations = 8, 55
	start := make(chan struct{})
	failures := make(chan string, readers)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 1600; i++ {
			SetDefaultPrecision(precisions[i%len(precisions)])
			if i%4 == 0 {
				runtime.Gosched()
			}
		}
	}()
	for reader := 0; reader < readers; reader++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < iterations; i++ {
				for j, operation := range operations {
					got := concurrentFingerprint(operation.run())
					if _, ok := allowed[j][got]; !ok {
						failures <- fmt.Sprintf("%s returned non-snapshot value %s", operation.name, got)
						return
					}
				}
				if i%4 == 0 {
					runtime.Gosched()
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(failures)
	for failure := range failures {
		t.Error(failure)
	}
}

// Exercise the interval-based PowInt implementation on shared, indirect
// 512-bit coefficients. Its expected result follows from the bounded binomial
// argument covered by the independent Phase A exponent-extreme regression.
func TestPhaseCConcurrentBoundedPowerOnSharedWideCoefficient(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("large integer exponents require 64-bit int")
	}
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	oneScaled := pow10(int(MaxPrecision))
	base := FromString("1." + strings.Repeat("0", int(MaxPrecision)-1) + "1")
	if base.IsNaN() || base.wide == nil {
		t.Fatal("fixture must use an indirect 512-bit coefficient")
	}
	positiveWant := packBigPowerExpected(new(big.Int).Add(new(big.Int).Set(oneScaled), new(big.Int).SetUint64(uint64(maxInt))))
	negativeWant := packBigPowerExpected(new(big.Int).Sub(new(big.Int).Set(oneScaled), new(big.Int).SetUint64(uint64(maxInt)+1)))
	input := concurrentFingerprint(base)
	const goroutines, iterations = 6, 10
	start := make(chan struct{})
	failures := make(chan string, goroutines)
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < iterations; i++ {
				plus, minus := base.PowInt(maxInt), base.PowInt(minInt)
				if got := concurrentFingerprint(plus); got != positiveWant {
					failures <- "PowInt(MaxInt) changed under concurrent reads: " + got
					return
				}
				if got := concurrentFingerprint(minus); got != negativeWant {
					failures <- "PowInt(MinInt) changed under concurrent reads: " + got
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(failures)
	for failure := range failures {
		t.Error(failure)
	}
	if concurrentFingerprint(base) != input {
		t.Fatal("bounded power modified the shared coefficient")
	}
}

// Construct a numerically independent expected decimal at the maximum scale.
func packBigPowerExpected(coefficient *big.Int) string {
	digits := coefficient.String()
	digits = strings.Repeat("0", int(MaxPrecision)+1-len(digits)) + digits
	return concurrentFingerprint(FromString(digits[:1] + "." + digits[1:]))
}

func TestPhaseCConcurrentReadOnlySQLAndSerializers(t *testing.T) {
	shared := FromString(strings.Repeat("8", 90) + ".123456789")
	sharedSQL := SQLDecimal{Decimal: shared, Valid: true}
	serialized, err := json.Marshal(shared)
	if err != nil {
		t.Fatal(err)
	}
	text, err := shared.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	want := concurrentFingerprint(shared)
	const goroutines, iterations = 10, 80
	start := make(chan struct{})
	failures := make(chan string, goroutines)
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < iterations; i++ {
				var decoded, textDecoded Optimus
				if err := json.Unmarshal(serialized, &decoded); err != nil {
					failures <- err.Error()
					return
				}
				if err := textDecoded.UnmarshalText(text); err != nil {
					failures <- err.Error()
					return
				}
				var scanned SQLDecimal
				if err := scanned.Scan([]byte(shared.String())); err != nil {
					failures <- err.Error()
					return
				}
				value, err := sharedSQL.Value()
				if err != nil || value == nil {
					failures <- fmt.Sprintf("Value returned %v, %v", value, err)
					return
				}
				if concurrentFingerprint(decoded) != want || concurrentFingerprint(textDecoded) != want || concurrentFingerprint(scanned.Decimal) != want || concurrentFingerprint(shared) != want {
					failures <- "concurrent read or local destination changed numeric state"
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(failures)
	for failure := range failures {
		t.Error(failure)
	}
}

func BenchmarkPhaseCSharedWideReadParallel(b *testing.B) {
	left := FromString(strings.Repeat("7", 110) + ".125")
	right := FromString(strings.Repeat("8", 70) + ".25")
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var result Optimus
		for pb.Next() {
			result = left.Add(right)
		}
		runtime.KeepAlive(result)
	})
}

func BenchmarkPhaseCSharedSmallReadParallel(b *testing.B) {
	left, right := FromInt(1234567), FromInt(431)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var result Optimus
		for pb.Next() {
			result = left.Add(right)
		}
		runtime.KeepAlive(result)
	})
}
