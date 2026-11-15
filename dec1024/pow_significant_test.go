package dec1024

import (
	"math"
	"testing"

	"github.com/profe-ajedrez/badassitron/internal/decimalwide"
)

func TestPowIntVariableWidthDifferential(t *testing.T) {
	bases := []string{"0", "-0", "1", "-1", "1.0101", "-1.0101", "1.0000", "2.5", "0.1", "-0.001", "12345678901234567890", "0.0000000000000000000000000000001"}
	exponents := []int{0, 1, 2, 3, 5, 9, 16, 32, 40, 64, 128, 256, -1, -2, -5, -40, math.MinInt}
	for _, bs := range bases {
		a := FromString(bs)
		if a.IsNaN() {
			t.Fatal(bs)
		}
		for _, n := range exponents {
			got := a.PowInt(n)
			want := decimalwide.PowInt(referenceNumber(a), n, 1024, int(MaxPrecision), int(defaultPrecision.Load()))
			assertReference(t, bs, got, want)
			if !got.IsNaN() && int(got.Precision()) != want.Scale {
				t.Fatalf("%s^%d scale got %d want %d", bs, n, got.Precision(), want.Scale)
			}
		}
	}
}

var benchmarkPowOutput Dec1024

func BenchmarkPowIntWidthCases(b *testing.B) {
	for _, tc := range []struct {
		name  string
		base  string
		power int
	}{
		{"1.0101^40", "1.0101", 40},
		{"1.0101^(-40)", "1.0101", -40},
		{"1.25^16", "1.25", 16},
		{"0.1^100", "0.1", 100},
		{"-2^40", "-2", 40},
		{"99999^3", "99999", 3},
	} {
		x := FromString(tc.base)
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmarkPowOutput = x.PowInt(tc.power)
			}
		})
	}
}

func TestPowIntNoHeap(t *testing.T) {
	x := FromString("1.0101")
	if got := testing.AllocsPerRun(300, func() { benchmarkPowOutput = x.PowInt(40) }); got != 0 {
		t.Fatalf("PowInt(40) allocated %g", got)
	}
}
