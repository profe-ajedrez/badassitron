package optimus

import (
	"strings"
	"testing"
)

var phaseBQuotient, phaseBRemainder Optimus

// These cases deliberately exercise both small and wide divisors, including
// alignment that cannot fit in the 512-bit arithmetic backend.
func phaseBDivisionCases() []struct {
	name                   string
	numerator, denominator Optimus
} {
	return []struct {
		name                   string
		numerator, denominator Optimus
	}{
		{"SmallEqual", FromInt(123456789012345), FromInt(123)},
		{"SmallDifferent", FromString("12345.67"), FromString("3.1")},
		{"Wide256Equal", FromString(strings.Repeat("9", 70)), FromString(strings.Repeat("7", 37))},
		{"Wide512Equal", FromString(strings.Repeat("9", 145)), FromString(strings.Repeat("7", 70))},
		{"WideDifferent", FromString(strings.Repeat("8", 88) + ".123"), FromString(strings.Repeat("7", 50) + ".00001")},
		{"AlignmentOverflow", FromString(strings.Repeat("9", 154)), FromString("0." + strings.Repeat("0", 153) + "1")},
		{"WideBy64", FromString(strings.Repeat("9", 145)), FromInt(12345)},
		{"WideNear", FromString(strings.Repeat("9", 145)), FromString(strings.Repeat("8", 144))},
		{"WideLess", FromString(strings.Repeat("7", 80)), FromString(strings.Repeat("9", 145))},
	}
}

func BenchmarkPhaseBMod(b *testing.B) {
	for _, tc := range phaseBDivisionCases() {
		b.Run(tc.name, func(b *testing.B) {
			a, c := tc.numerator, tc.denominator
			b.ReportAllocs()
			b.ResetTimer()
			var result Optimus
			for i := 0; i < b.N; i++ {
				result = a.Mod(c)
			}
			phaseBRemainder = result
		})
	}
}

func BenchmarkPhaseBQuoRem(b *testing.B) {
	for _, tc := range phaseBDivisionCases() {
		b.Run(tc.name, func(b *testing.B) {
			a, c := tc.numerator, tc.denominator
			b.ReportAllocs()
			b.ResetTimer()
			var q, r Optimus
			for i := 0; i < b.N; i++ {
				q, r = a.QuoRem(c)
			}
			phaseBQuotient, phaseBRemainder = q, r
		})
	}
}

// PowCoarse exercises the first representable scale well below MaxPrecision.
func BenchmarkPhaseBPowCoarse(b *testing.B) {
	a := FromString("1.01")
	b.ReportAllocs()
	b.ResetTimer()
	var result Optimus
	for i := 0; i < b.N; i++ {
		result = a.PowInt(5000)
	}
	phaseBQuotient = result
}
