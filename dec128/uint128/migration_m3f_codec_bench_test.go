package uint128_test

import (
	"testing"

	upstream "github.com/jokruger/dec128/uint128"
	facade "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

var m3fCodecFacadeSink facade.Uint128
var m3fCodecUpstreamSink upstream.Uint128
var m3fCodecTextSink string
var m3fCodecStateSink uint8

// BenchmarkMigrationM3FCodec measures the compatibility facade and direct
// Jokruger using identical operands and the same output visibility rules.
// The historical facade is available under badassitron_uint128_legacy.
func BenchmarkMigrationM3FCodec(b *testing.B) {
	inputs := [32]string{}
	for i := range inputs {
		inputs[i] = "340282366920938463463374607431768211455"
		if i%2 == 0 {
			inputs[i] = "17293822569102704650112939346821"
		}
	}
	for _, tc := range []struct {
		name string
		in   [32]string
	}{
		{"ParseLong", inputs},
		{"ParseShort", func() [32]string {
			var v [32]string
			for i := range v {
				v[i] = "1234567890"
			}
			return v
		}()},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.Run("Facade", func(b *testing.B) {
				b.ReportAllocs()
				var last facade.Uint128
				var code uint8
				for i := 0; i < b.N; i++ {
					v, e := facade.FromString(tc.in[i&31])
					last, code = v, uint8(e)
				}
				m3fCodecFacadeSink, m3fCodecStateSink = last, code
			})
			b.Run("UpstreamDirect", func(b *testing.B) {
				b.ReportAllocs()
				var last upstream.Uint128
				var code uint8
				for i := 0; i < b.N; i++ {
					v, e := upstream.FromString(tc.in[i&31])
					last, code = v, uint8(e)
				}
				m3fCodecUpstreamSink, m3fCodecStateSink = last, code
			})
		})
	}
	vals := [32]facade.Uint128{}
	up := [32]upstream.Uint128{}
	for i := range vals {
		vals[i] = facade.Uint128{Lo: 0xffffffeeee444111 + uint64(i), Hi: 0xcafedeadf0030000 + uint64(i)}
		up[i] = upstream.Uint128(vals[i])
	}
	b.Run("StringToBufWide", func(b *testing.B) {
		b.Run("Facade", func(b *testing.B) {
			b.ReportAllocs()
			buf := [39]byte{}
			var out []byte
			for i := 0; i < b.N; i++ {
				out = vals[i&31].StringToBuf(buf[:])
			}
			m3fCodecTextSink = string(out)
		})
		b.Run("UpstreamDirect", func(b *testing.B) {
			b.ReportAllocs()
			buf := [39]byte{}
			var out []byte
			for i := 0; i < b.N; i++ {
				out = up[i&31].StringToBuf(buf[:])
			}
			m3fCodecTextSink = string(out)
		})
	})
}
