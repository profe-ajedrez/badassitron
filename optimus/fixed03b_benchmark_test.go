package optimus

import (
	"strings"
	"testing"
)

var fixed03bBenchSink Optimus
var fixed03bBenchStatus bool

func BenchmarkFixed03BRescale(b *testing.B) {
	a128 := FromString("1234567890123456.1234567890123456789")
	a256 := FromString("0." + strings.Repeat("7", 70))
	a512 := FromString("0." + strings.Repeat("8", 150))
	cases := []struct {
		name string
		d    Optimus
		p    uint8
		m    RoundingMode
	}{
		{"same128", a128, a128.Precision(), RoundingHalfEven},
		{"same512", a512, a512.Precision(), RoundingHalfEven},
		{"down1", FromString("12.345"), 2, RoundingHalfEven},
		{"down2", FromString("1234.56789"), 3, RoundingHalfEven},
		{"down6", FromString("1234.123456789"), 3, RoundingHalfEven},
		{"down19", FromString("12." + strings.Repeat("9", 19)), 0, RoundingHalfEven},
		{"down20", FromString("12." + strings.Repeat("9", 20)), 0, RoundingHalfEven},
		{"down38wide", a256, 32, RoundingHalfEven},
		{"down76wide", a512, 74, RoundingHalfEven},
		{"up2", FromString("1.25"), 4, RoundingHalfEven},
		{"up19", FromString("1.25"), 21, RoundingHalfEven},
		{"overflow512", a512, 154, RoundingHalfEven},
		{"zero", FromString("0.000000"), 2, RoundingHalfEven},
		{"tieEven", FromString("2.500"), 0, RoundingHalfEven},
		{"tieAway", FromString("-2.500"), 0, RoundingHalfAwayFromZero},
		{"floorNeg", FromString("-2.511"), 0, RoundingFloor},
		{"ceilPos", FromString("2.511"), 0, RoundingCeil},
		{"invalidMode", a128, 2, RoundingMode(255)},
		{"invalidScale", a128, 255, RoundingHalfEven},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				fixed03bBenchSink, fixed03bBenchStatus = tc.d.RescaleRoundWithStatus(tc.p, tc.m)
			}
		})
	}
}
