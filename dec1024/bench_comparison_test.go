package dec1024

import (
	"github.com/profe-ajedrez/badassitron/dec512"
	"testing"
)

var resultDec512 dec512.Dec512
var resultDec1024 Dec1024

// Same representable financial operands and equivalent rounding contract.
// This measures primitive overhead, not Iacon throughput.
func BenchmarkDec512Vs1024(b *testing.B) {
	x512 := dec512.FromString("123456.78")
	y512 := dec512.FromString("987.65")
	x1024 := FromDec512(x512)
	y1024 := FromDec512(y512)
	b.Run("Dec512/Add", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resultDec512 = x512.Add(y512)
		}
	})
	b.Run("Dec1024/Add", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resultDec1024 = x1024.Add(y1024)
		}
	})
	b.Run("Dec512/Mul", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resultDec512 = x512.Mul(y512)
		}
	})
	b.Run("Dec1024/Mul", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resultDec1024 = x1024.Mul(y1024)
		}
	})
	b.Run("Dec512/MulRound", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resultDec512, _ = x512.MulRoundWithStatus(y512, 2, RoundingHalfEven)
		}
	})
	b.Run("Dec1024/MulRound", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resultDec1024, _ = x1024.MulRoundWithStatus(y1024, 2, RoundingHalfEven)
		}
	})
	b.Run("Dec512/DivRound", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resultDec512, _ = x512.DivRoundWithStatus(y512, 2, RoundingHalfEven)
		}
	})
	b.Run("Dec1024/DivRound", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resultDec1024, _ = x1024.DivRoundWithStatus(y1024, 2, RoundingHalfEven)
		}
	})
}
