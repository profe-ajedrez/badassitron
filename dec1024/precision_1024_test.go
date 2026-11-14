package dec1024

import (
	"math/big"
	"strings"
	"testing"
	"unsafe"

	"github.com/profe-ajedrez/badassitron/dec1024/uint1024"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec512"
)

func TestExtendedPrecisionAndCoefficientBoundaries(t *testing.T) {
	if MaxPrecision != 308 {
		t.Fatalf("MaxPrecision=%d", MaxPrecision)
	}
	if size := unsafe.Sizeof(Dec1024{}); size != 136 {
		t.Fatalf("Dec1024 size=%d", size)
	}
	highest := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 1024), big.NewInt(1))
	maximum := FromString(highest.String())
	if maximum.IsNaN() || maximum.Coefficient() != uint1024.Max {
		t.Fatal("max unsigned coefficient")
	}
	if !FromString(new(big.Int).Add(highest, big.NewInt(1)).String()).IsNaN() {
		t.Fatal("accepted overflowing coefficient")
	}

	tinyText := "0." + strings.Repeat("0", 307) + "1"
	tiny := FromString(tinyText)
	if tiny.IsNaN() || tiny.Precision() != 308 || tiny.Coefficient() != uint1024.One {
		t.Fatalf("tiny=%s scale=%d", tiny, tiny.Precision())
	}
	if tiny.StringFixed() != tinyText {
		t.Fatal("lost tiny stored scale")
	}
	if got := tiny.RescaleRound(307, RoundingAwayFromZero).StringFixed(); got != "0."+strings.Repeat("0", 306)+"1" {
		t.Fatalf("rounded tiny=%s", got)
	}
	if got := tiny.RescaleRound(307, RoundingTowardZero); !got.IsZero() {
		t.Fatalf("trunc tiny=%v", got)
	}
	if d := FromString("0." + strings.Repeat("0", 308) + "1"); !d.IsNaN() || d.ErrorDetails() != errors.PrecisionOutOfRange.Value() {
		t.Fatal("accepted scale 309")
	}
	if d := New(uint1024.One, 309, false); !d.IsNaN() {
		t.Fatal("accepted invalid constructed scale")
	}

	// Both operands fit in 1024 bits. Their exact product is 2000 bits,
	// but the normalized mathematical result is precisely 1.
	large := "1" + strings.Repeat("0", 300)
	unity := FromString(large[:1] + "." + large[1:]) // 1.00...00 (scale 300)
	if unity.IsNaN() || unity.Precision() != 300 {
		t.Fatal("bad fixture")
	}
	if d := unity.Mul(unity); d.IsNaN() || !d.Equal(One) {
		t.Fatalf("2000-bit intermediate result=%v", d)
	}
	if d, ix := unity.MulRoundWithStatus(unity, 2, RoundingHalfEven); d.IsNaN() || d.StringFixed() != "1.00" || ix {
		t.Fatalf("rounded large product=%v ix=%v", d, ix)
	}
}

func TestWidenNarrowDec512Exact(t *testing.T) {
	source := dec512.FromString("-1234567890.0000000400")
	wide := FromDec512(source)
	if wide.IsNaN() || wide.StringFixed() != source.StringFixed() {
		t.Fatal("widen lost representation")
	}
	narrow, e := wide.Dec512Exact()
	if e != nil || !narrow.Equal(source) {
		t.Fatalf("narrow=%v error=%v", narrow, e)
	}
	// Canonicalized narrowing is allowed when the stored scale exceeds Dec512's limit.
	more := New(uint1024.FromUint64(100), 155, false)
	narrower, e := more.Dec512Exact()
	if e != nil || narrower.Precision() != 153 || !FromDec512(narrower).Equal(more) {
		t.Fatalf("narrow canonical=%v error=%v", narrower, e)
	}
	if _, e := FromString("0." + strings.Repeat("0", 307) + "1").Dec512Exact(); e == nil {
		t.Fatal("unexpected narrow tiny")
	}
	if _, e := New(uint1024.Max, 0, false).Dec512Exact(); e == nil {
		t.Fatal("unexpected narrow max")
	}
}

func BenchmarkDec1024Basic(b *testing.B) {
	a := FromString("123456789012345678901234567890.123456789")
	c := FromString("987654321098765432109876543210.987654321")
	b.Run("add", func(b *testing.B) {
		var out Dec1024
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			out = a.Add(c)
		}
		if out.IsNaN() {
			b.Fatal(out)
		}
	})
	b.Run("mul", func(b *testing.B) {
		var out Dec1024
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			out = a.Mul(c)
		}
		if out.IsNaN() {
			b.Fatal(out)
		}
	})
	b.Run("mulround", func(b *testing.B) {
		var out Dec1024
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			out, _ = a.MulRoundWithStatus(c, 2, RoundingHalfEven)
		}
		if out.IsNaN() {
			b.Fatal(out)
		}
	})
	b.Run("divround", func(b *testing.B) {
		var out Dec1024
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			out, _ = a.DivRoundWithStatus(c, 2, RoundingHalfEven)
		}
		if out.IsNaN() {
			b.Fatal(out)
		}
	})
}
