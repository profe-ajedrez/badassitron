package dec128

import (
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	"testing"
)

var v4BenchDecSink Dec128
var v4BenchCmpSink int

func BenchmarkV4Common(b *testing.B) {
	mk := func(lo, hi uint64, exp uint8, neg bool) Dec128 { return New(uint128.Uint128{Lo: lo, Hi: hi}, exp, neg) }
	cases := []struct {
		name  string
		pairs [8][2]Dec128
	}{
		{name: "SameScale", pairs: [8][2]Dec128{
			{mk(12345, 0, 2, false), mk(6789, 0, 2, false)},
			{mk(100, 0, 2, false), mk(200, 0, 2, false)},
			{mk(34567, 0, 3, true), mk(1221, 0, 3, true)},
			{mk(1, 0, 0, false), mk(3, 0, 0, false)},
			{mk(999999999, 0, 2, false), mk(1, 0, 2, false)},
			{mk(19901, 0, 5, true), mk(901, 0, 5, true)},
			{mk(1024, 0, 0, false), mk(2048, 0, 0, false)},
			{mk(77, 0, 1, false), mk(42, 0, 1, false)},
		}},
		{name: "DifferentScale", pairs: [8][2]Dec128{
			{mk(12345, 0, 2, false), mk(6789, 0, 3, false)},
			{mk(100, 0, 1, false), mk(200, 0, 4, false)},
			{mk(34567, 0, 3, true), mk(1221, 0, 5, true)},
			{mk(1, 0, 0, false), mk(3, 0, 8, false)},
			{mk(999999999, 0, 2, false), mk(1, 0, 3, false)},
			{mk(19901, 0, 5, true), mk(901, 0, 7, true)},
			{mk(1024, 0, 0, false), mk(2048, 0, 1, false)},
			{mk(77, 0, 1, false), mk(42, 0, 2, false)},
		}},
		{name: "OppositeSign", pairs: [8][2]Dec128{
			{mk(12345, 0, 2, false), mk(6789, 0, 2, true)},
			{mk(100, 0, 2, true), mk(200, 0, 2, false)},
			{mk(34567, 0, 3, true), mk(1221, 0, 3, false)},
			{mk(1, 0, 0, false), mk(3, 0, 0, true)},
			{mk(999999999, 0, 2, false), mk(1, 0, 2, true)},
			{mk(19901, 0, 5, true), mk(901, 0, 5, false)},
			{mk(1024, 0, 0, false), mk(2048, 0, 0, true)},
			{mk(77, 0, 1, true), mk(42, 0, 1, false)},
		}},
		{name: "Wide", pairs: [8][2]Dec128{
			{mk(0xffffffffff, 1, 0, false), mk(6789, 2, 0, false)},
			{mk(100, 1, 1, false), mk(200, 2, 1, false)},
			{mk(34567, 1, 3, true), mk(1221, 1, 3, true)},
			{mk(1, 4, 0, false), mk(3, 8, 0, false)},
			{mk(999999999, 3, 2, false), mk(1, 2, 2, false)},
			{mk(19901, 1, 5, true), mk(901, 2, 5, true)},
			{mk(1024, 4, 0, false), mk(2048, 8, 0, false)},
			{mk(77, 1, 1, false), mk(42, 2, 1, false)},
		}},
		{name: "Overflow", pairs: [8][2]Dec128{
			{mk(^uint64(0), ^uint64(0), 0, false), mk(1, 0, 0, false)},
			{mk(^uint64(0)-10, ^uint64(0), 1, false), mk(100, 0, 1, false)},
			{mk(^uint64(0), ^uint64(0), 5, true), mk(7, 0, 5, true)},
			{mk(1, ^uint64(0), 0, false), mk(1, 1, 0, false)},
			{mk(^uint64(0), ^uint64(0), 2, false), mk(10, 0, 2, false)},
			{mk(^uint64(0), ^uint64(0), 5, true), mk(1, 0, 5, true)},
			{mk(^uint64(0), ^uint64(0), 0, false), mk(^uint64(0), 0, 0, false)},
			{mk(0, ^uint64(0), 1, false), mk(1, 5, 1, false)},
		}},
	}
	for _, op := range []string{"Add", "Sub", "Compare", "Mul"} {
		for _, test := range cases {
			b.Run(op+"/"+test.name, func(b *testing.B) {
				a := test.pairs
				b.ReportAllocs()
				b.ResetTimer()
				switch op {
				case "Add":
					var r Dec128
					for i := 0; i < b.N; i++ {
						r = a[i&7][0].Add(a[i&7][1])
					}
					v4BenchDecSink = r
				case "Sub":
					var r Dec128
					for i := 0; i < b.N; i++ {
						r = a[i&7][0].Sub(a[i&7][1])
					}
					v4BenchDecSink = r
				case "Compare":
					var r int
					for i := 0; i < b.N; i++ {
						r = a[i&7][0].Compare(a[i&7][1])
					}
					v4BenchCmpSink = r
				case "Mul":
					var r Dec128
					for i := 0; i < b.N; i++ {
						r = a[i&7][0].Mul(a[i&7][1])
					}
					v4BenchDecSink = r
				}
			})
		}
	}
}
