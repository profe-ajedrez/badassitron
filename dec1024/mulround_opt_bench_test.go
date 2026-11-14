package dec1024

import (
    "strings"
    "testing"
    "github.com/profe-ajedrez/badassitron/dec1024/uint1024"
    "github.com/profe-ajedrez/badassitron/dec512/uint512"
)

var mulRoundOptSink Dec1024
var mulRoundOptInexact bool

func BenchmarkMulRoundOpt(b *testing.B) {
    one300 := FromString("1." + strings.Repeat("0", 300))
    max308 := New(uint1024.Max, 308, false)
    half512 := New(uint1024.FromUint512(uint512.Max), 22, false)
    smallA, smallB := FromString("1234.56789"), FromString("98.7654321")
    tests := []struct { name string; a, y Dec1024; precision uint16; mode RoundingMode }{
        {"small-exact", FromString("12.34"), FromString("56.78"), 4, RoundingHalfEven},
        {"small-round-2", smallA, smallB, 2, RoundingHalfEven},
        {"small-round-8", smallA, smallB, 8, RoundingHalfEven},
        {"medium-512bit-round-2", half512, half512, 2, RoundingHalfEven},
        {"wide-2000bit-round-2", one300, one300, 2, RoundingHalfEven},
        {"wide-2000bit-round-308", one300, one300, 308, RoundingHalfEven},
        {"wide-max-round-0", max308, max308, 0, RoundingHalfEven},
        {"wide-max-round-308", max308, max308, 308, RoundingHalfEven},
        {"wide-max-round-ceil", max308, max308, 0, RoundingCeil},
    }
    for _, tc := range tests {
        b.Run(tc.name, func(b *testing.B) {
            b.ReportAllocs()
            b.ResetTimer()
            for i:=0; i<b.N; i++ { mulRoundOptSink, mulRoundOptInexact = tc.a.MulRoundWithStatus(tc.y, tc.precision, tc.mode) }
        })
    }
}
