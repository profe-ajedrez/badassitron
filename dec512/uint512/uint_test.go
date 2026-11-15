package uint512

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec256/uint256"
)

var benchmarkSink Uint512

func oracleFromBig(t *testing.T, x *big.Int) Uint512 {
	t.Helper()
	y, err := FromBigInt(x)
	if err != errors.None {
		t.Fatalf("FromBigInt(%v): %v", x, err)
	}
	return y
}
func oracleEquals(t *testing.T, kind string, got Uint512, want *big.Int) {
	t.Helper()
	if got.BigInt().Cmp(want) != 0 {
		t.Fatalf("%s: got %s, want %s", kind, got.String(), want.String())
	}
}

func TestBoundaryAndConstruction(t *testing.T) {
	limit := new(big.Int).Lsh(big.NewInt(1), 512)
	max := new(big.Int).Sub(new(big.Int).Set(limit), big.NewInt(1))
	if Max.BigInt().Cmp(max) != 0 {
		t.Fatalf("Max does not match 2^512-1")
	}
	if Max.String() != max.String() || len(Max.String()) != MaxStrLen {
		t.Fatal("Max.String mismatch")
	}
	if !Zero.IsZero() || Zero.String() != "0" || One.String() != "1" {
		t.Fatal("zero or one")
	}
	if Zero.BitLen() != 0 || Max.BitLen() != 512 {
		t.Fatal("bit lengths")
	}
	if Max.LeadingZeroBitsCount() != 0 || Zero.LeadingZeroBitsCount() != 512 || Zero.TrailingZeroBitsCount() != 512 {
		t.Fatal("zero bits")
	}
	if Max.NonZeroBitsCount() != 512 || One.TrailingZeroBitsCount() != 0 {
		t.Fatal("ones count")
	}
	if _, e := FromBigInt(new(big.Int).Set(limit)); e != errors.Overflow {
		t.Fatal("big overflow")
	}
	if _, e := FromBigInt(big.NewInt(-1)); e != errors.Negative {
		t.Fatal("negative big")
	}
	if _, e := FromBigInt(nil); e != errors.InvalidFormat {
		t.Fatal("nil big")
	}
	if _, e := FromString(limit.String()); e != errors.Overflow {
		t.Fatal("string overflow")
	}
	if _, e := FromString("12.3"); e != errors.InvalidFormat {
		t.Fatal("invalid string")
	}
	if _, e := FromString("-1"); e != errors.InvalidFormat {
		t.Fatal("negative string")
	}
	if v, e := FromString(""); e != errors.None || !v.IsZero() {
		t.Fatal("empty string")
	}
	if v, e := FromString("00010"); e != errors.None || v.String() != "10" {
		t.Fatal("leading zeros")
	}
	if _, e := Max.Add(One); e != errors.Overflow {
		t.Fatal("add overflow")
	}
	if _, e := Zero.Sub(One); e != errors.Underflow {
		t.Fatal("sub underflow")
	}
	if _, e := Max.Mul(FromUint64(2)); e != errors.Overflow {
		t.Fatal("mul overflow")
	}
	if _, _, e := Max.QuoRem(Zero); e != errors.DivisionByZero {
		t.Fatal("divide by zero")
	}
	if _, _, e := Max.QuoRem64(0); e != errors.DivisionByZero {
		t.Fatal("divide by zero 64")
	}
	if _, e := FromString(strings.Repeat("0", 300) + "123"); e != errors.None {
		t.Fatal("leading zeros overflow")
	}
	for i, p := range Pow10Uint512 {
		expected := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(i)), nil)
		if p.BigInt().Cmp(expected) != 0 {
			t.Fatalf("power of ten %d", i)
		}
	}
	if _, e := Max.Uint64(); e != errors.Overflow {
		t.Fatal("Uint64 overflow")
	}
	if val, e := FromUint64(123).Uint64(); e != errors.None || val != 123 {
		t.Fatal("Uint64 conversion")
	}
}

func TestRandomAgainstBigInt(t *testing.T) {
	rnd := rand.New(rand.NewSource(5122026))
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	for i := 0; i < 3500; i++ {
		var ab, bb [64]byte
		_, _ = rnd.Read(ab[:])
		_, _ = rnd.Read(bb[:])
		a := FromBytes(ab)
		b := FromBytes(bb)
		x, y := a.BigInt(), b.BigInt()
		if a.Compare(b) != x.Cmp(y) || a.Equal(b) != (x.Cmp(y) == 0) {
			t.Fatalf("compare %d", i)
		}

		want := new(big.Int).Add(x, y)
		sum, e := a.Add(b)
		if want.Cmp(max) > 0 {
			if e != errors.Overflow {
				t.Fatalf("add overflow %d", i)
			}
		} else if e != errors.None {
			t.Fatalf("add err %d: %v", i, e)
		} else {
			oracleEquals(t, "add", sum, want)
		}

		want = new(big.Int).Sub(x, y)
		diff, e := a.Sub(b)
		if want.Sign() < 0 {
			if e != errors.Underflow {
				t.Fatalf("sub underflow %d", i)
			}
		} else if e != errors.None {
			t.Fatalf("sub err %d: %v", i, e)
		} else {
			oracleEquals(t, "sub", diff, want)
		}

		want = new(big.Int).Mul(x, y)
		lo, hi := a.MulCarry(b)
		full := new(big.Int).Lsh(hi.BigInt(), 512)
		full.Add(full, lo.BigInt())
		if full.Cmp(want) != 0 {
			t.Fatalf("mulCarry %d", i)
		}
		mul, e := a.Mul(b)
		if want.Cmp(max) > 0 {
			if e != errors.Overflow {
				t.Fatalf("mul overflow %d", i)
			}
		} else if e != errors.None {
			t.Fatalf("mul err %d: %v", i, e)
		} else {
			oracleEquals(t, "mul", mul, want)
		}

		if !b.IsZero() {
			q, r, e := a.QuoRem(b)
			if e != errors.None {
				t.Fatalf("div err %d: %v", i, e)
			}
			wq, wr := new(big.Int).QuoRem(x, y, new(big.Int))
			oracleEquals(t, "quotient", q, wq)
			oracleEquals(t, "remainder", r, wr)
			d, _ := a.Div(b)
			m, _ := a.Mod(b)
			if !q.Equal(d) || !r.Equal(m) {
				t.Fatalf("div/mod inconsistency %d", i)
			}
		}
		d := rnd.Uint64()
		if d == 0 {
			d = 1
		}
		q, r, e := a.QuoRem64(d)
		if e != errors.None {
			t.Fatalf("div64 err %d", i)
		}
		wq, wr := new(big.Int).QuoRem(x, new(big.Int).SetUint64(d), new(big.Int))
		oracleEquals(t, "quotient64", q, wq)
		if r != wr.Uint64() {
			t.Fatalf("remainder64 %d", i)
		}

		shift := uint(rnd.Intn(512 + 128))
		wl := new(big.Int).And(new(big.Int).Lsh(x, shift), max)
		wr = new(big.Int).Rsh(x, shift)
		oracleEquals(t, "lsh", a.Lsh(shift), wl)
		oracleEquals(t, "rsh", a.Rsh(shift), wr)
		oracleEquals(t, "and", a.And(b), new(big.Int).And(x, y))
		oracleEquals(t, "or", a.Or(b), new(big.Int).Or(x, y))
		oracleEquals(t, "xor", a.Xor(b), new(big.Int).Xor(x, y))
		if a.NonZeroBitsCount() != countBigBits(x) {
			t.Fatalf("ones count %d", i)
		}
		if !a.RotateBitsLeft(512).Equal(a) || !a.RotateBitsRight(512).Equal(a) {
			t.Fatalf("rotate full width %d", i)
		}
		if !a.RotateBitsLeft(17).RotateBitsRight(17).Equal(a) {
			t.Fatalf("rotate inverse %d", i)
		}
		if !a.ReverseBits().ReverseBits().Equal(a) || !a.ReverseBytes().ReverseBytes().Equal(a) {
			t.Fatalf("reverse involution %d", i)
		}
		if a.String() != x.String() {
			t.Fatalf("string %d", i)
		}
		parsed, e := FromString(a.String())
		if e != errors.None || !parsed.Equal(a) {
			t.Fatalf("parse roundtrip %d", i)
		}
		if !FromBytes(a.Bytes()).Equal(a) || !FromBytesBigEndian(a.BytesBigEndian()).Equal(a) {
			t.Fatalf("bytes %d", i)
		}
		txt, _ := a.MarshalText()
		var unmarshaled Uint512
		if err := unmarshaled.UnmarshalText(txt); err != nil || !unmarshaled.Equal(a) {
			t.Fatalf("text roundtrip %d", i)
		}
	}
}
func countBigBits(x *big.Int) int {
	n := 0
	for _, b := range x.Bytes() {
		for b > 0 {
			n += int(b & 1)
			b >>= 1
		}
	}
	return n
}

func TestDivisionDirected(t *testing.T) {
	cases := []struct{ u, v string }{
		{Max.String(), "3"}, {Max.String(), "18446744073709551616"}, {Max.String(), Max.String()},
		{"1", Max.String()}, {"0", Max.String()}, {Max.String(), "340282366920938463463374607431768211455"},
		{Max.String(), "170141183460469231731687303715884105727"},
		{Max.String(), "340282366920938463463374607431768211457"},
	}
	for _, c := range cases {
		u, ue := FromString(c.u)
		v, ve := FromString(c.v)
		if ue != errors.None || ve != errors.None {
			t.Fatal("bad test fixture", c)
		}
		q, r, e := u.QuoRem(v)
		if e != errors.None {
			t.Fatalf("division failed %s/%s", c.u, c.v)
		}
		x, y := u.BigInt(), v.BigInt()
		eq, er := new(big.Int).QuoRem(x, y, new(big.Int))
		oracleEquals(t, "direct quotient", q, eq)
		oracleEquals(t, "direct remainder", r, er)
	}
	for _, n := range []uint{0, 1, 63, 64, 65, 256, 511, 512, 513} {
		x := Max.Lsh(n)
		expected := new(big.Int).Lsh(Max.BigInt(), n)
		expected.And(expected, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1)))
		oracleEquals(t, "direct lsh", x, expected)
	}
}

func TestBytesTextAndScanner(t *testing.T) {
	x := Max
	if !bytes.Equal(x.AppendBytes(nil), func() []byte { b := x.Bytes(); return b[:] }()) {
		t.Fatal("append bytes")
	}
	if !bytes.Equal(x.AppendBytesBigEndian(nil), func() []byte { b := x.BytesBigEndian(); return b[:] }()) {
		t.Fatal("append BE bytes")
	}
	var b [64]byte
	if e := x.PutBytes(b[:64-1]); e != errors.NotEnoughBytes {
		t.Fatal("put short")
	}
	if e := x.PutBytesBigEndian(b[:64-1]); e != errors.NotEnoughBytes {
		t.Fatal("put BE short")
	}
	var into Uint512
	_, err := fmt.Sscan("123456789", &into)
	if err != nil || into.String() != "123456789" {
		t.Fatalf("scan err: %v (%v)", err, into.String())
	}
	_, err = fmt.Sscan("-1", &into)
	if err == nil {
		t.Fatal("negative scan")
	}
	if got := hex.EncodeToString(func() []byte { b := x.BytesBigEndian(); return b[:] }()); len(got) != 128 {
		t.Fatal("hex length")
	}
	// Parse invalid input must not mutate destination.
	old := into
	if e := into.UnmarshalText([]byte("bad")); e == nil || into != old {
		t.Fatal("unmarshal mutation")
	}
	// BigInt conversion must not mutate caller-owned big.Int.
	source := new(big.Int).SetUint64(123456789)
	_, ec := FromBigInt(source)
	if ec != errors.None || source.Uint64() != 123456789 {
		t.Fatal("big.Int mutated")
	}
}

func BenchmarkWideAdd(b *testing.B) {
	x, y := Max.Rsh(1), FromUint64(27)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkSink, _ = x.Add(y)
	}
}
func BenchmarkWideMul(b *testing.B) {
	x, y := Max.Rsh(4), FromUint64(7)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkSink, _ = x.Mul(y)
	}
}
func BenchmarkWideDivSingleLimb(b *testing.B) {
	x, y := Max, FromUint64(999999937)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkSink, _ = x.Div(y)
	}
}
func BenchmarkWideDivMultiLimb(b *testing.B) {
	x, y := Max, Max.Rsh(96)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkSink, _ = x.Div(y)
	}
}

// TestDivisionVariedWidths exercises multiword quotients, rather than only
// same-size random operands (whose quotient is usually zero or one).
func TestDivisionVariedWidths(t *testing.T) {
	rng := rand.New(rand.NewSource(512771))
	for i := 0; i < 5000; i++ {
		divisorBits := 1 + rng.Intn(512)
		numeratorBits := 1 + rng.Intn(512)
		// Generate independent nonzero positive integers of those bit sizes.
		numerator := new(big.Int).Lsh(big.NewInt(1), uint(numeratorBits-1))
		if numeratorBits > 1 {
			random := new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(numeratorBits-1)))
			numerator.Add(numerator, random)
		}
		divisor := new(big.Int).Lsh(big.NewInt(1), uint(divisorBits-1))
		if divisorBits > 1 {
			random := new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(divisorBits-1)))
			divisor.Add(divisor, random)
		}
		u, v := oracleFromBig(t, numerator), oracleFromBig(t, divisor)
		q, r, e := u.QuoRem(v)
		if e != errors.None {
			t.Fatalf("division %d: %v", i, e)
		}
		wq, wr := new(big.Int).QuoRem(numerator, divisor, new(big.Int))
		if q.BigInt().Cmp(wq) != 0 || r.BigInt().Cmp(wr) != 0 {
			t.Fatalf("division %d, bits=(%d,%d): got q=%s r=%s; want q=%s r=%s", i, numeratorBits, divisorBits, q, r, wq, wr)
		}
	}
}

func FuzzWideDivision(f *testing.F) {
	f.Add(uint64(1), uint64(2), uint64(3), uint64(4))
	f.Add(uint64(0), uint64(0), uint64(0), uint64(0))
	f.Add(uint64(18446744073709551615), uint64(1), uint64(18446744073709551615), uint64(0))
	f.Fuzz(func(t *testing.T, a, b, c, d uint64) {
		x := FromUint64(a).Lsh(uint(b % 512))
		v := FromUint64(c).Lsh(uint(d % 512))
		q, r, e := x.QuoRem(v)
		if v.IsZero() {
			if e != errors.DivisionByZero {
				t.Fatalf("zero divisor: %v", e)
			}
			return
		}
		if e != errors.None {
			t.Fatalf("division error: %v", e)
		}
		xbig, vbig := x.BigInt(), v.BigInt()
		wq, wr := new(big.Int).QuoRem(xbig, vbig, new(big.Int))
		if q.BigInt().Cmp(wq) != 0 || r.BigInt().Cmp(wr) != 0 {
			t.Fatal("division differs from big.Int")
		}
	})
}

func TestWidening(t *testing.T) {
	original := uint256.Max
	widened := FromUint256(original)
	recovered, e := widened.Uint256()
	if e != errors.None || recovered != original {
		t.Fatal("uint256->uint512->uint256")
	}
	if _, e := Max.Uint256(); e != errors.Overflow {
		t.Fatal("narrow overflow")
	}
}
