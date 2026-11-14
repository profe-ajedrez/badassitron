package optimus

import (
	"strings"
	"testing"
)

var fixed01Result Optimus
var fixed01Inexact bool

// Inputs are constructed outside the measured loop. The same file is used by
// both the original and experimental modules for matched comparisons.
func fixed01BenchCases() []struct {
	name string
	a, b Optimus
	p    uint8
	mode RoundingMode
} {
	return []struct {
		name string
		a, b Optimus
		p    uint8
		mode RoundingMode
	}{
		{"small_exact", FromString("123.45"), FromString("1.20"), 2, RoundingHalfEven},
		{"small_round", FromString("19.995"), One, 2, RoundingHalfEven},
		{"small_8dp", FromString("0.123456789"), FromString("1.0001"), 8, RoundingHalfEven},
		{"128x128_exact", FromString("12345678901234567890123456789012345678"), FromString("98765432101234567890987654321012345678"), 0, RoundingHalfEven},
		{"256x256_exact", FromString(strings.Repeat("8", 72)), FromString(strings.Repeat("7", 71)), 0, RoundingHalfEven},
		{"512xsmall_exact", FromString(strings.Repeat("9", 140)), One, 0, RoundingHalfEven},
		{"512x512_quantized", FromString("0." + strings.Repeat("9", 150)), FromString("0." + strings.Repeat("8", 150)), 2, RoundingHalfEven},
		{"512x512_8dp", FromString("0." + strings.Repeat("9", 154)), FromString("0." + strings.Repeat("8", 154)), 8, RoundingHalfAwayFromZero},
		{"512x512_overflow", FromString(strings.Repeat("9", 150)), FromString(strings.Repeat("8", 150)), 0, RoundingHalfEven},
		{"precision_154", FromString("0.123456789"), FromString("0.987654321"), 154, RoundingHalfEven},
		{"zero_scaled", Zero, FromString("0.000125"), 2, RoundingHalfEven},
		{"invalid_mode", FromString("12.3"), FromString("4.56"), 2, RoundingMode(255)},
	}
}

func BenchmarkFixed01MulRound(b *testing.B) {
	for _, tc := range fixed01BenchCases() {
		b.Run(tc.name, func(b *testing.B) {
			a, other, p, mode := tc.a, tc.b, tc.p, tc.mode
			b.ReportAllocs()
			b.ResetTimer()
			var result Optimus
			var inexact bool
			for i := 0; i < b.N; i++ {
				result, inexact = a.MulRoundWithStatus(other, p, mode)
			}
			fixed01Result, fixed01Inexact = result, inexact
		})
	}
}
