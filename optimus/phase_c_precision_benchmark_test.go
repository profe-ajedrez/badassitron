package optimus

import "testing"

var phaseCPrecisionQuotient, phaseCPrecisionRemainder Optimus

// This benchmark is also run unchanged against v7. The overflow-alignment
// case produces an incorrect quotient in v7; its timing is diagnostic only.
func BenchmarkPhaseCPrecisionQuoRemAligned128(b *testing.B) {
	cases := []struct {
		name string
		a, c Optimus
	}{
		{"AlignedFits128", FromString("8507059173023461588.1519253575114968768"), FromString("30")},
		{"AlignedOverflows128", FromString("-8507059173023461588.1519253575114968768"), FromString("34389314001497466858")},
		{"DifferentSmall", FromString("12345.67"), FromString("3.1")},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			a, c := tc.a, tc.c
			b.ResetTimer()
			var q, r Optimus
			for i := 0; i < b.N; i++ {
				q, r = a.QuoRem(c)
			}
			phaseCPrecisionQuotient, phaseCPrecisionRemainder = q, r
		})
	}
}
