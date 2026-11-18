package uint128_test

import (
	upstream "github.com/jokruger/dec128/uint128"
	facade "github.com/profe-ajedrez/badassitron/dec128/uint128"
	"testing"
)

// This benchmark is built unchanged under both build modes: the default facade
// is historical, while badassitron_m3 enables the imported backend adapter.
// UpstreamDirect always calls the imported primitive without the facade.
var threeWayFacadeSink facade.Uint128
var threeWayUpstreamSink upstream.Uint128
var threeWayFacadeRemainderSink facade.Uint128
var threeWayUpstreamRemainderSink upstream.Uint128
var threeWayStatusSink uint64

func BenchmarkMigrationM3ThreeWay(b *testing.B) {
	var fa, fb [32]facade.Uint128
	var ua, ub [32]upstream.Uint128
	for i := range fa {
		fa[i] = facade.Uint128{Lo: 0xa7839b428dc5 + uint64(i)*31, Hi: 0x1200 + uint64(i)}
		fb[i] = facade.Uint128{Lo: 0x79123 + uint64(i)*13, Hi: 3}
		ua[i] = upstream.Uint128{Lo: fa[i].Lo, Hi: fa[i].Hi}
		ub[i] = upstream.Uint128{Lo: fb[i].Lo, Hi: fb[i].Hi}
	}
	b.Run("Add", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			b.ReportAllocs()
			var out facade.Uint128
			var remainder facade.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				r, s := fa[j].Add(fb[j])
				out = r
				code = uint8(s)
			}
			threeWayFacadeSink = out
			threeWayFacadeRemainderSink = remainder
			threeWayStatusSink = uint64(code)
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			b.ReportAllocs()
			var out upstream.Uint128
			var remainder upstream.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				r, s := ua[j].Add(ub[j])
				out = r
				code = uint8(s)
			}
			threeWayUpstreamSink = out
			threeWayUpstreamRemainderSink = remainder
			threeWayStatusSink = uint64(code)
		})
	})
	b.Run("MulExact", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			b.ReportAllocs()
			var out facade.Uint128
			var remainder facade.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				r, s := fa[j].Mul64(17)
				out = r
				code = uint8(s)
			}
			threeWayFacadeSink = out
			threeWayFacadeRemainderSink = remainder
			threeWayStatusSink = uint64(code)
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			b.ReportAllocs()
			var out upstream.Uint128
			var remainder upstream.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				r, s := ua[j].Mul64(17)
				out = r
				code = uint8(s)
			}
			threeWayUpstreamSink = out
			threeWayUpstreamRemainderSink = remainder
			threeWayStatusSink = uint64(code)
		})
	})
	// These cases exercise the successful Uint128.Mul path, not Mul64.
	// The existing MulOverflow case only exercises the error path.
	for _, tc := range []struct {
		name        string
		left, right func(i int) facade.Uint128
	}{
		{
			name: "MulNarrowExact",
			left: func(i int) facade.Uint128 {
				return facade.Uint128{Lo: (1 << 32) + uint64(i), Hi: 0}
			},
			right: func(i int) facade.Uint128 { return facade.Uint128{Lo: 17 + uint64(i&3)} },
		},
		{
			name: "MulHighExact",
			left: func(i int) facade.Uint128 {
				return facade.Uint128{Lo: 0xa7839b428dc5 + uint64(i), Hi: 1 + uint64(i&3)}
			},
			right: func(i int) facade.Uint128 { return facade.Uint128{Lo: 2 + uint64(i&3)} },
		},
	} {
		var left, right [32]facade.Uint128
		var upLeft, upRight [32]upstream.Uint128
		for i := range left {
			left[i], right[i] = tc.left(i), tc.right(i)
			upLeft[i] = upstream.Uint128(left[i])
			upRight[i] = upstream.Uint128(right[i])
		}
		b.Run(tc.name, func(b *testing.B) {
			b.Run("Facade", func(b *testing.B) {
				b.ReportAllocs()
				var result facade.Uint128
				var code uint8
				for i := 0; i < b.N; i++ {
					j := i & 31
					r, s := left[j].Mul(right[j])
					result, code = r, uint8(s)
				}
				threeWayFacadeSink, threeWayStatusSink = result, uint64(code)
			})
			b.Run("UpstreamDirect", func(b *testing.B) {
				b.ReportAllocs()
				var result upstream.Uint128
				var code uint8
				for i := 0; i < b.N; i++ {
					j := i & 31
					r, s := upLeft[j].Mul(upRight[j])
					result, code = r, uint8(s)
				}
				threeWayUpstreamSink, threeWayStatusSink = result, uint64(code)
			})
		})
	}
	b.Run("MulOverflow", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			b.ReportAllocs()
			var out facade.Uint128
			var remainder facade.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				r, s := fa[j].Mul(fb[j])
				out = r
				code = uint8(s)
			}
			threeWayFacadeSink = out
			threeWayFacadeRemainderSink = remainder
			threeWayStatusSink = uint64(code)
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			b.ReportAllocs()
			var out upstream.Uint128
			var remainder upstream.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				r, s := ua[j].Mul(ub[j])
				out = r
				code = uint8(s)
			}
			threeWayUpstreamSink = out
			threeWayUpstreamRemainderSink = remainder
			threeWayStatusSink = uint64(code)
		})
	})
	b.Run("QuoRem", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			b.ReportAllocs()
			var out facade.Uint128
			var remainder facade.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				q, r, s := fa[j].QuoRem(fb[j])
				out = q
				remainder = r
				code = uint8(s)
			}
			threeWayFacadeSink = out
			threeWayFacadeRemainderSink = remainder
			threeWayStatusSink = uint64(code)
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			b.ReportAllocs()
			var out upstream.Uint128
			var remainder upstream.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				q, r, s := ua[j].QuoRem(ub[j])
				out = q
				remainder = r
				code = uint8(s)
			}
			threeWayUpstreamSink = out
			threeWayUpstreamRemainderSink = remainder
			threeWayStatusSink = uint64(code)
		})
	})
	b.Run("WideQuoRem", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			b.ReportAllocs()
			var out facade.Uint128
			var remainder facade.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				q, r, s := facade.QuoRem256By128(fa[j], facade.One, fb[j])
				out = q
				remainder = r
				code = uint8(s)
			}
			threeWayFacadeSink = out
			threeWayFacadeRemainderSink = remainder
			threeWayStatusSink = uint64(code)
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			b.ReportAllocs()
			var out upstream.Uint128
			var remainder upstream.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				q, r, s := upstream.QuoRem256By128(ua[j], upstream.One, ub[j])
				out = q
				remainder = r
				code = uint8(s)
			}
			threeWayUpstreamSink = out
			threeWayUpstreamRemainderSink = remainder
			threeWayStatusSink = uint64(code)
		})
	})

	// Error paths must be benchmarked independently from successful division.
	b.Run("AddOverflow", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			b.ReportAllocs()
			var out facade.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				r, s := facade.Max.Add(fb[i&31])
				out, code = r, uint8(s)
			}
			threeWayFacadeSink, threeWayStatusSink = out, uint64(code)
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			b.ReportAllocs()
			var out upstream.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				r, s := upstream.Max.Add(ub[i&31])
				out, code = r, uint8(s)
			}
			threeWayUpstreamSink, threeWayStatusSink = out, uint64(code)
		})
	})
	b.Run("QuoRem64", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			b.ReportAllocs()
			var out facade.Uint128
			var rem uint64
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				q, r, s := fa[j].QuoRem64(fb[j].Lo)
				out, rem, code = q, r, uint8(s)
			}
			threeWayFacadeSink, threeWayStatusSink = out, uint64(code)^rem
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			b.ReportAllocs()
			var out upstream.Uint128
			var rem uint64
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				q, r, s := ua[j].QuoRem64(ub[j].Lo)
				out, rem, code = q, r, uint8(s)
			}
			threeWayUpstreamSink, threeWayStatusSink = out, uint64(code)^rem
		})
	})
	b.Run("QuoRemZero", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			b.ReportAllocs()
			var out, remainder facade.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				q, r, s := fa[i&31].QuoRem(facade.Zero)
				out, remainder, code = q, r, uint8(s)
			}
			threeWayFacadeSink, threeWayFacadeRemainderSink, threeWayStatusSink = out, remainder, uint64(code)
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			b.ReportAllocs()
			var out, remainder upstream.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				q, r, s := ua[i&31].QuoRem(upstream.Zero)
				out, remainder, code = q, r, uint8(s)
			}
			threeWayUpstreamSink, threeWayUpstreamRemainderSink, threeWayStatusSink = out, remainder, uint64(code)
		})
	})
	b.Run("WideOverflow", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			b.ReportAllocs()
			var out, remainder facade.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				q, r, s := facade.QuoRem256By128(fa[j], fb[j], fb[j])
				out, remainder, code = q, r, uint8(s)
			}
			threeWayFacadeSink, threeWayFacadeRemainderSink, threeWayStatusSink = out, remainder, uint64(code)
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			b.ReportAllocs()
			var out, remainder upstream.Uint128
			var code uint8
			for i := 0; i < b.N; i++ {
				j := i & 31
				q, r, s := upstream.QuoRem256By128(ua[j], ub[j], ub[j])
				out, remainder, code = q, r, uint8(s)
			}
			threeWayUpstreamSink, threeWayUpstreamRemainderSink, threeWayStatusSink = out, remainder, uint64(code)
		})
	})
}
