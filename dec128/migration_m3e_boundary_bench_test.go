package dec128

import (
	"testing"

	ju "github.com/jokruger/dec128/uint128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

var migrationM3EBoundarySink Dec128

func BenchmarkMigrationM3ECoefficientBoundary(b *testing.B) {
	var coefficients [32]ju.Uint128
	for i := range coefficients {
		coefficients[i] = ju.Uint128{Lo: uint64(i+1)*13579, Hi: uint64(i % 5)}
	}
	b.Run("Internal", func(b *testing.B) {
		var out Dec128
		for i := 0; i < b.N; i++ {
			out = fromParts(coefficients[i&31], uint8(i%19), i&1 != 0)
		}
		migrationM3EBoundarySink = out
	})
	b.Run("PublicNew", func(b *testing.B) {
		var out Dec128
		for i := 0; i < b.N; i++ {
			out = New(lu.Uint128(coefficients[i&31]), uint8(i%19), i&1 != 0)
		}
		migrationM3EBoundarySink = out
	})
}
