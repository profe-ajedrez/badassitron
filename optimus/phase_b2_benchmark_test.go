package optimus

import (
	"strings"
	"testing"
)

var phaseB2Value Optimus
var phaseB2Status bool

func phaseB2MulCases() []struct {
	name string
	a, b Optimus
	p    uint8
} {
	return []struct {
		name string
		a, b Optimus
		p    uint8
	}{
		{"OverflowInteger", FromString(strings.Repeat("9", 150)), FromString(strings.Repeat("8", 150)), 19},
		{"Intermediate1024", FromString("0." + strings.Repeat("9", 152)), FromString("0." + strings.Repeat("8", 152)), 154},
		{"Intermediate768", FromString("0." + strings.Repeat("9", 110)), FromString("0." + strings.Repeat("8", 110)), 154},
		{"WideExactFit", FromString(strings.Repeat("9", 70)), FromString(strings.Repeat("8", 70)), 19},
		{"WideRounding19", FromString("0." + strings.Repeat("9", 150)), FromString("0." + strings.Repeat("8", 150)), 19},
	}
}
func BenchmarkPhaseB2Mul(b *testing.B) {
	for _, tc := range phaseB2MulCases() {
		b.Run(tc.name, func(b *testing.B) {
			a, c := tc.a, tc.b
			b.ReportAllocs()
			b.ResetTimer()
			var v Optimus
			for i := 0; i < b.N; i++ {
				v = a.Mul(c)
			}
			phaseB2Value = v
		})
	}
}
func BenchmarkPhaseB2MulRound(b *testing.B) {
	for _, tc := range phaseB2MulCases() {
		b.Run(tc.name, func(b *testing.B) {
			a, c, p := tc.a, tc.b, tc.p
			b.ReportAllocs()
			b.ResetTimer()
			var v Optimus
			var x bool
			for i := 0; i < b.N; i++ {
				v, x = a.MulRoundWithStatus(c, p, RoundingHalfEven)
			}
			phaseB2Value, phaseB2Status = v, x
		})
	}
}
func BenchmarkPhaseB2DivRound(b *testing.B) {
	for _, tc := range phaseBDivisionCases() {
		if tc.name != "Wide512Equal" && tc.name != "WideDifferent" && tc.name != "WideBy64" && tc.name != "Wide256Equal" && tc.name != "AlignmentOverflow" {
			continue
		}
		b.Run(tc.name, func(b *testing.B) {
			a, c := tc.numerator, tc.denominator
			b.ReportAllocs()
			b.ResetTimer()
			var v Optimus
			var x bool
			for i := 0; i < b.N; i++ {
				v, x = a.DivRoundWithStatus(c, 19, RoundingHalfEven)
			}
			phaseB2Value, phaseB2Status = v, x
		})
	}
}
