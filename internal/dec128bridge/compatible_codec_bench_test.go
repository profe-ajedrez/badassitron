package dec128bridge

import (
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	"testing"
)

var codecLegacySink legacy.Dec128
var codecBridgeSink Compatible
var codecStringSink string

func BenchmarkMigrationCodecParse(b *testing.B) {
	const input = "123456789012345678.1234567890123456789"
	b.Run("legacy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			codecLegacySink = legacy.FromString(input)
		}
	})
	b.Run("bridge", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			codecBridgeSink = ParseCompatible(input)
		}
	})
}
func BenchmarkMigrationCodecString(b *testing.B) {
	const input = "123456789012345678.1234567890123456789"
	l := legacy.FromString(input)
	w := ParseCompatible(input)
	b.Run("legacy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			codecStringSink = l.String()
		}
	})
	b.Run("bridge", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			codecStringSink = w.String()
		}
	})
}

// Shapes are deliberately diverse: a win on one long fractional input is
// insufficient to establish a parser-wide performance improvement.
func BenchmarkMigrationParserShapes(b *testing.B) {
	cases := []struct {
		name string
		text string
	}{
		{"short_integer", "123456789012345678"},
		{"short_fraction", "-12345.67890"},
		{"long_fraction", "123456789012345678.1234567890123456789"},
		{"long_integer", "340282366920938463463374607431768211455"},
		{"long_zero", "-000000000000000000000000000000000000000000000"},
		{"overlong_scale", "0.000000000000000000000000000000000000000000"},
	}
	for _, tc := range cases {
		b.Run(tc.name+"/legacy", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				codecLegacySink = legacy.FromString(tc.text)
			}
		})
		b.Run(tc.name+"/bridge", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				codecBridgeSink = ParseCompatible(tc.text)
			}
		})
	}
}
