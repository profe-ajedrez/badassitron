package benchmark

import (
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// BenchmarkM2Cutover uses identical public-method calls in the legacy and
// experimental builds. The two builds must be benchmarked separately with
// the same toolchain, machine, GOMAXPROCS and benchtime.
var (
	cutoverDecimalSink dec128.Dec128
	cutoverCompareSink int
	cutoverStringSink  string
	cutoverErrSink     error
)

func BenchmarkM2Cutover(b *testing.B) {
	smallA, smallB := dec128.FromString("1234567.89"), dec128.FromString("9876.54")
	fracA, fracB := dec128.FromString("1234567.890123"), dec128.FromString("9876.5401")
	largeA := dec128.FromString("34028236692093846346337460743176821145")
	nearMax := dec128.FromString(uint128.MaxUint128Str)
	mulTiny := dec128.FromString("0.0000000001")
	divExactA, divExactB := dec128.FromString("400"), dec128.FromString("8")
	three := dec128.FromInt(3)
	cases := []struct {
		label string
		call  func()
	}{
		{"Add/Small", func() { cutoverDecimalSink = smallA.Add(smallB) }},
		{"Add/MixedScale", func() { cutoverDecimalSink = fracA.Add(fracB) }},
		{"Add/OppositeSign", func() { cutoverDecimalSink = smallA.Sub(smallB) }},
		{"Add/NearOverflow", func() { cutoverDecimalSink = nearMax.Add(dec128.One) }},
		{"Mul/Small", func() { cutoverDecimalSink = smallA.Mul(smallB) }},
		{"Mul/Large", func() { cutoverDecimalSink = largeA.Mul(smallB) }},
		{"Mul/ScaleOverflow", func() { cutoverDecimalSink = mulTiny.Mul(mulTiny) }},
		{"Div/Exact", func() { cutoverDecimalSink = divExactA.Div(divExactB) }},
		{"Div/Repeating", func() { cutoverDecimalSink = smallA.Div(three) }},
		{"Compare/MixedScale", func() { cutoverCompareSink = fracA.Compare(fracB) }},
		{"Parse/Short", func() { cutoverDecimalSink = dec128.FromString("12.345") }},
		{"Parse/Long", func() { cutoverDecimalSink = dec128.FromString("123456789012345678901234567.123456789") }},
		{"String/Fixed", func() { cutoverStringSink = fracA.StringFixed() }},
		{"String/Canonical", func() { cutoverStringSink = fracA.String() }},
		{"Rescale/Truncate", func() { cutoverDecimalSink = fracA.Rescale(2) }},
		{"Rescale/RoundHalfEven", func() { cutoverDecimalSink = fracA.RescaleRound(2, dec128.RoundingHalfEven) }},
		{"Rounding/MulRound", func() { cutoverDecimalSink = fracA.MulRound(fracB, 2, dec128.RoundingHalfEven) }},
		{"Rounding/DivRound", func() { cutoverDecimalSink = fracA.DivRound(fracB, 2, dec128.RoundingHalfEven) }},
		{"Sqrt/FullWidth", func() { cutoverDecimalSink = nearMax.Sqrt() }},
		{"SQL/ScanInto", func() {
			var dst dec128.Dec128
			cutoverErrSink = dec128.ScanInto(&dst, "12345.678901234")
			cutoverDecimalSink = dst
		}},
	}
	for _, tc := range cases {
		b.Run(tc.label, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tc.call()
			}
		})
	}
}
