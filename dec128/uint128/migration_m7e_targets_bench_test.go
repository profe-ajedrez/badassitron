package uint128_test

import (
	"testing"

	ju "github.com/jokruger/dec128/uint128"
	bu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// BenchmarkReleaseM7ETargets complements (not replaces) the 14 stable
// facade/upstream comparisons. Its arrays mix success and exceptional cases
// so an optimization for a constant zero divisor cannot silently degrade the
// common nonzero-divisor path.
func BenchmarkReleaseM7ETargets(b *testing.B) {
	var a, divisor, carry [32]bu.Uint128
	var ua, ud, uc [32]ju.Uint128
	var texts [32]string
	for i := range a {
		a[i] = bu.Uint128{Lo: 1200003 + uint64(i), Hi: 501 + uint64(i)}
		divisor[i] = bu.Uint128{Lo: 2003 + uint64(i), Hi: 17}
		if i%3 == 0 {
			carry[i] = bu.Uint128{Lo: 1}
		}
		if i%16 == 0 {
			divisor[i] = bu.Zero
			carry[i] = bu.Zero
		} else if i%16 == 1 {
			carry[i] = divisor[i] // wide quotient overflow
		}
		texts[i] = "1234567890"
		if i%16 == 0 {
			texts[i] = "123456789x" // short invalid-format path
		}
		ua[i], ud[i], uc[i] = ju.Uint128(a[i]), ju.Uint128(divisor[i]), ju.Uint128(carry[i])
	}
	b.Run("QuoRemMixed", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			var q, r bu.Uint128
			var st uint8
			for i := 0; i < b.N; i++ {
				q0, r0, code := a[i&31].QuoRem(divisor[i&31])
				q, r, st = q0, r0, uint8(code)
			}
			threeWayFacadeSink, threeWayFacadeRemainderSink, threeWayStatusSink = q, r, uint64(st)
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			var q, r ju.Uint128
			var st uint8
			for i := 0; i < b.N; i++ {
				q0, r0, code := ua[i&31].QuoRem(ud[i&31])
				q, r, st = q0, r0, uint8(code)
			}
			threeWayUpstreamSink, threeWayUpstreamRemainderSink, threeWayStatusSink = q, r, uint64(st)
		})
	})
	b.Run("WideQuoRemMixed", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			var q, r bu.Uint128
			var st uint8
			for i := 0; i < b.N; i++ {
				q0, r0, code := bu.QuoRem256By128(a[i&31], carry[i&31], divisor[i&31])
				q, r, st = q0, r0, uint8(code)
			}
			threeWayFacadeSink, threeWayFacadeRemainderSink, threeWayStatusSink = q, r, uint64(st)
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			var q, r ju.Uint128
			var st uint8
			for i := 0; i < b.N; i++ {
				q0, r0, code := ju.QuoRem256By128(ua[i&31], uc[i&31], ud[i&31])
				q, r, st = q0, r0, uint8(code)
			}
			threeWayUpstreamSink, threeWayUpstreamRemainderSink, threeWayStatusSink = q, r, uint64(st)
		})
	})
	b.Run("ParseShortMixed", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			var n bu.Uint128
			var st uint8
			for i := 0; i < b.N; i++ {
				v, code := bu.FromString(texts[i&31])
				n, st = v, uint8(code)
			}
			m3fCodecFacadeSink, m3fCodecStateSink = n, st
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			var n ju.Uint128
			var st uint8
			for i := 0; i < b.N; i++ {
				v, code := ju.FromString(texts[i&31])
				n, st = v, uint8(code)
			}
			m3fCodecUpstreamSink, m3fCodecStateSink = n, st
		})
	})
}
