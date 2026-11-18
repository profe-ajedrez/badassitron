package dec128

import (
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

var fastPathBenchmarkResult Dec128
var fastPathBenchmarkCompare int

func benchOperands() ([8]Dec128, [8]Dec128) {
	var a, b [8]Dec128
	for i := 0; i < 8; i++ {
		a[i] = New(uint128.Uint128{Lo: uint64(1234567 + 101*i), Hi: 0}, 2, false)
		b[i] = New(uint128.Uint128{Lo: uint64(12345 + 13*i), Hi: 0}, 2, false)
	}
	return a, b
}

func benchmarkFastPath(b *testing.B, op string, mode string) {
	aa, bb := benchOperands()
	if mode == "DifferentScale" {
		for i := range bb {
			bb[i] = New(bb[i].Coefficient(), 3, false)
		}
	}
	if mode == "OppositeSign" {
		for i := range bb {
			bb[i] = New(bb[i].Coefficient(), 2, true)
		}
	}
	if mode == "Wide" {
		for i := range aa {
			aa[i] = New(uint128.Uint128{Lo: uint64(1234567 + i), Hi: 1 << 48}, 0, false)
			bb[i] = New(uint128.Uint128{Lo: uint64(12345 + i), Hi: 1}, 0, false)
		}
	}
	if mode == "Overflow" {
		for i := range aa {
			aa[i] = New(uint128.Uint128{Lo: ^uint64(0) - uint64(i), Hi: ^uint64(0)}, 0, false)
			bb[i] = New(uint128.Uint128{Lo: 2 + uint64(i), Hi: 0}, 0, false)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink Dec128
	var cmp int
	for i := 0; i < b.N; i++ {
		a, c := aa[i&7], bb[i&7]
		switch op {
		case "Add":
			sink = a.Add(c)
		case "Sub":
			sink = a.Sub(c)
		case "Mul":
			sink = a.Mul(c)
		case "Compare":
			cmp = a.Compare(c)
		}
	}
	fastPathBenchmarkResult = sink
	fastPathBenchmarkCompare = cmp
}

func BenchmarkFastPath(b *testing.B) {
	for _, op := range []string{"Add", "Sub", "Compare", "Mul"} {
		for _, mode := range []string{"SameScale", "DifferentScale", "OppositeSign", "Wide", "Overflow"} {
			b.Run(op+"/"+mode, func(b *testing.B) { benchmarkFastPath(b, op, mode) })
		}
	}
}
