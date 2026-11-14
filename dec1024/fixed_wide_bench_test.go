package dec1024

import (
	"github.com/profe-ajedrez/badassitron/dec1024/uint1024"
	"strings"
	"testing"
)

var wideBenchOut Dec1024
var wideBenchBool bool

func BenchmarkFixedWidePaths(b *testing.B) {
	unity := FromString("1." + strings.Repeat("0", 300))
	maxScale := New(uint1024.Max, 308, false)
	maxAtOne := New(uint1024.Max, 1, false)
	d2 := FromInt(2)
	powBase := FromString("1.0101")
	for _, tc := range []struct {
		name string
		run  func() (Dec1024, bool)
	}{
		{"compare", func() (Dec1024, bool) { return FromInt(maxScale.Compare(unity)), false }},
		{"add-overflow-fitable", func() (Dec1024, bool) { return maxAtOne.Add(maxAtOne), false }},
		{"mul-2000bit-fit", func() (Dec1024, bool) { return unity.Mul(unity), false }},
		{"mulround-2000bit", func() (Dec1024, bool) { return unity.MulRoundWithStatus(unity, 2, RoundingHalfEven) }},
		{"div-2048bit-scale", func() (Dec1024, bool) { return maxScale.DivRoundWithStatus(maxScale, 308, RoundingHalfEven) }},
		{"sqrt-2048bit-scale", func() (Dec1024, bool) { return d2.SqrtRoundWithStatus(308, RoundingHalfEven) }},
		{"pow", func() (Dec1024, bool) { return powBase.PowInt(40), false }},
		{"mod-quotient-overflow", func() (Dec1024, bool) { return FromInt(5).Mod(New(uint1024.FromUint64(3), 308, false)), false }},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				wideBenchOut, wideBenchBool = tc.run()
			}
		})
	}
}
