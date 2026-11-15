package benchmark

import (
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

var v2Sink dec128.Dec128

func BenchmarkV2SqrtRegular(b *testing.B) {
	dec128.SetDefaultPrecision(19)
	v := dec128.FromString("2")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v2Sink = v.Sqrt()
	}
}

func BenchmarkV2SqrtFullWidth(b *testing.B) {
	dec128.SetDefaultPrecision(19)
	v := dec128.FromString(uint128.MaxUint128Str)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v2Sink = v.Sqrt()
	}
}

func BenchmarkV2QuoRemScaleOverflow(b *testing.B) {
	a := dec128.FromString("1.2345678901234567890")
	v := dec128.FromString(uint128.MaxUint128Str)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v2Sink, _ = a.QuoRem(v)
	}
}

func BenchmarkV2ScanInto(b *testing.B) {
	var v dec128.Dec128
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := dec128.ScanInto(&v, "12345.678901234"); err != nil {
			b.Fatal(err)
		}
	}
	v2Sink = v
}
