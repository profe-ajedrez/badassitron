package uint1024

import (
	"math/big"
	"testing"
	"unsafe"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

func Test1024WideBoundaries(t *testing.T) {
	if size := unsafe.Sizeof(Uint1024{}); size != 128 {
		t.Fatalf("size %d", size)
	}
	if len(Pow10Uint1024) != 309 {
		t.Fatal("powers incomplete")
	}
	if Pow10Uint1024[308].BitLen() > 1024 {
		t.Fatal("last power too large")
	}
	if x, e := FromBigInt(new(big.Int).Lsh(big.NewInt(1), 1024)); e != errors.Overflow || x != Zero {
		t.Fatalf("overflow: %v", e)
	}
	original := uint512.Max
	grown := FromUint512(original)
	if low, e := grown.Uint512(); e != errors.None || low != original {
		t.Fatalf("widen/contract: %v", e)
	}
	if _, e := Max.Uint512(); e != errors.Overflow {
		t.Fatalf("narrow: %v", e)
	}
	hi := One.Lsh(1023)
	if hi.Hi.IsZero() || !hi.Lo.IsZero() {
		t.Fatal("high limb propagation")
	}
	if hi.Rsh(1023) != One {
		t.Fatal("high shift return")
	}
	lo, high := Max.MulCarry(Max)
	full := new(big.Int).Lsh(high.BigInt(), 1024)
	full.Add(full, lo.BigInt())
	expected := new(big.Int).Mul(Max.BigInt(), Max.BigInt())
	if full.Cmp(expected) != 0 {
		t.Fatal("2048-bit product corruption")
	}
}

func BenchmarkUint1024Operations(b *testing.B) {
	a := Max.Rsh(5)
	c := Max.Rsh(7)
	b.Run("add", func(b *testing.B) {
		var v Uint1024
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			v, _ = a.Add(c)
		}
		if v.IsZero() {
			b.Fatal(v)
		}
	})
	b.Run("mulcarry", func(b *testing.B) {
		var lo, hi Uint1024
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			lo, hi = a.MulCarry(c)
		}
		if lo.IsZero() && hi.IsZero() {
			b.Fatal(lo)
		}
	})
	b.Run("quorem", func(b *testing.B) {
		var q, r Uint1024
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			q, r, _ = a.QuoRem(c)
		}
		if q.IsZero() && r.IsZero() {
			b.Fatal(q)
		}
	})
}
