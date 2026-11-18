package uint128_test

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// This oracle is deliberately independent of both the historical and upstream
// algorithms. It runs under the default and optional badassitron_m3 builds.
func TestMigrationM3Unsigned128AgainstBigInt(t *testing.T) {
	rng := rand.New(rand.NewSource(0x4d335f313238))
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	boundary := []uint128.Uint128{
		uint128.Zero, uint128.One, uint128.Max, uint128.Max64,
		{Lo: 0, Hi: 1}, {Lo: 0, Hi: 1 << 63},
		{Lo: 1, Hi: 1}, {Lo: ^uint64(0) - 1, Hi: ^uint64(0)},
	}
	toBig := func(u uint128.Uint128) *big.Int {
		return new(big.Int).Or(new(big.Int).Lsh(new(big.Int).SetUint64(u.Hi), 64), new(big.Int).SetUint64(u.Lo))
	}
	check := func(op string, i int, got uint128.Uint128, code errors.Error, exact *big.Int, wantError errors.Error) {
		t.Helper()
		if code != wantError {
			t.Fatalf("%s case %d: code=%v want=%v", op, i, code, wantError)
		}
		if wantError != errors.None {
			if got != uint128.Zero {
				t.Fatalf("%s case %d: error result=%v (want zero)", op, i, got)
			}
			return
		}
		if got.BigInt().Cmp(exact) != 0 {
			t.Fatalf("%s case %d: got=%v want=%s", op, i, got, exact)
		}
	}
	for i := 0; i < 5000; i++ {
		a := uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		b := uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		if i < len(boundary)*len(boundary) {
			a = boundary[i/len(boundary)]
			b = boundary[i%len(boundary)]
		}
		A, B := toBig(a), toBig(b)
		sum := new(big.Int).Add(A, B)
		want := errors.None
		if sum.Cmp(max) > 0 {
			want = errors.Overflow
		}
		u, c := a.Add(b)
		check("Add", i, u, c, sum, want)
		diff := new(big.Int).Sub(A, B)
		want = errors.None
		if diff.Sign() < 0 {
			want = errors.Underflow
		}
		u, c = a.Sub(b)
		check("Sub", i, u, c, diff, want)
		prod := new(big.Int).Mul(A, B)
		want = errors.None
		if prod.Cmp(max) > 0 {
			want = errors.Overflow
		}
		u, c = a.Mul(b)
		check("Mul", i, u, c, prod, want)
		low, high := a.MulCarry(b)
		p := new(big.Int).Or(toBig(low), new(big.Int).Lsh(toBig(high), 128))
		if p.Cmp(prod) != 0 {
			t.Fatalf("MulCarry case %d got=%s want=%s", i, p, prod)
		}
		if B.Sign() != 0 {
			q, r := new(big.Int).QuoRem(A, B, new(big.Int))
			gotQ, gotR, c := a.QuoRem(b)
			if c != errors.None || gotQ.BigInt().Cmp(q) != 0 || gotR.BigInt().Cmp(r) != 0 {
				t.Fatalf("QuoRem case %d: %v %v %v, want %s %s", i, gotQ, gotR, c, q, r)
			}
		} else {
			q, r, c := a.QuoRem(b)
			if q != uint128.Zero || r != uint128.Zero || c != errors.DivisionByZero {
				t.Fatalf("division by zero case %d", i)
			}
		}
		for _, n := range []uint{0, 1, 63, 64, 65, 127, 128, 129} {
			wantLeft := new(big.Int).And(new(big.Int).Lsh(A, n), max)
			wantRight := new(big.Int).Rsh(A, n)
			if a.Lsh(n).BigInt().Cmp(wantLeft) != 0 || a.Rsh(n).BigInt().Cmp(wantRight) != 0 {
				t.Fatalf("shift case %d n=%d a=%v", i, n, a)
			}
		}
	}
}

func TestMigrationM3WideDivisionAgainstBigInt(t *testing.T) {
	rng := rand.New(rand.NewSource(0x4d335f323536))
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	for i := 0; i < 4000; i++ {
		lo := uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		hi := uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		v := uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		switch i % 13 {
		case 0:
			hi = uint128.Zero
		case 1:
			v = uint128.One
		case 2:
			v = uint128.Max
		case 3:
			hi = uint128.Zero
			v = uint128.Max64
		case 4:
			hi = uint128.One
			v = uint128.Uint128{Lo: 0, Hi: 1}
		case 5:
			hi = uint128.Max
			v = uint128.Max
		case 6:
			hi = uint128.Uint128{Lo: 0, Hi: 1}
			v = uint128.Uint128{Lo: 0, Hi: 1}
		}
		numerator := new(big.Int).Or(new(big.Int).Lsh(hi.BigInt(), 128), lo.BigInt())
		q, r := new(big.Int).QuoRem(numerator, v.BigInt(), new(big.Int))
		wantCode := errors.None
		if q.Cmp(max) > 0 {
			wantCode = errors.Overflow
		}
		gotQ, gotR, code := uint128.QuoRem256By128(lo, hi, v)
		if code != wantCode || (code == errors.None && (gotQ.BigInt().Cmp(q) != 0 || gotR.BigInt().Cmp(r) != 0)) || (code != errors.None && (gotQ != uint128.Zero || gotR != uint128.Zero)) {
			t.Fatalf("wide division case %d: got=(%v,%v,%v) expected=(%s,%s,%v)", i, gotQ, gotR, code, q, r, wantCode)
		}
	}
}

func TestMigrationM3Div192By64AgainstBigInt(t *testing.T) {
	rng := rand.New(rand.NewSource(0x4d335f313932))
	for i := 0; i < 3000; i++ {
		u := uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		d := rng.Uint64() | 1
		carry := rng.Uint64()
		if i%5 != 0 {
			carry %= d
		}
		q, r, code := uint128.QuoRem192By64(u, carry, d)
		if carry >= d {
			if code != errors.Overflow || q != uint128.Zero || r != 0 {
				t.Fatalf("192/64 overflow case %d: (%v,%d,%v)", i, q, r, code)
			}
			continue
		}
		numerator := new(big.Int).Or(new(big.Int).Lsh(new(big.Int).SetUint64(carry), 128), u.BigInt())
		expectedQ, expectedR := new(big.Int).QuoRem(numerator, new(big.Int).SetUint64(d), new(big.Int))
		if code != errors.None || q.BigInt().Cmp(expectedQ) != 0 || r != expectedR.Uint64() {
			t.Fatalf("192/64 case %d: got (%v,%d,%v), want (%s,%s)", i, q, r, code, expectedQ, expectedR)
		}
	}
}

func TestMigrationM3WideDivisionErrorPrecedence(t *testing.T) {
	cases := []struct {
		low, high, divisor uint128.Uint128
		code               errors.Error
	}{
		{uint128.Zero, uint128.Zero, uint128.Zero, errors.DivisionByZero},
		{uint128.One, uint128.Zero, uint128.Zero, errors.DivisionByZero},
		{uint128.Zero, uint128.One, uint128.Zero, errors.Overflow},
		{uint128.Max, uint128.Max, uint128.One, errors.Overflow},
		{uint128.Zero, uint128.Max, uint128.Max, errors.Overflow},
	}
	for i, tc := range cases {
		q, r, code := uint128.QuoRem256By128(tc.low, tc.high, tc.divisor)
		if code != tc.code || q != uint128.Zero || r != uint128.Zero {
			t.Fatalf("case %d got %v %v %v; want zero, zero, %v", i, q, r, code, tc.code)
		}
	}
	for _, high := range []uint64{0, 1, ^uint64(0)} {
		q, r, code := uint128.QuoRem192By64(uint128.One, high, 0)
		if code != errors.Overflow || q != uint128.Zero || r != 0 {
			t.Fatalf("192/64 zero divisor with high=%d: %v %v %v", high, q, r, code)
		}
	}
}

func TestMigrationM3HistoricalConversionContract(t *testing.T) {
	// Retain the historical string and binary representations in both builds.
	values := []uint128.Uint128{uint128.Zero, uint128.One, uint128.Max64, uint128.Max}
	for _, u := range values {
		text, err := u.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		var roundtrip uint128.Uint128
		if err = roundtrip.UnmarshalText(text); err != nil || roundtrip != u {
			t.Fatalf("text %v: %v %v", u, roundtrip, err)
		}
		if uint128.FromBytes(u.Bytes()) != u || uint128.FromBytesBigEndian(u.BytesBigEndian()) != u {
			t.Fatalf("binary: %v", u)
		}
	}
}

var m3BenchmarkSink uint128.Uint128

func BenchmarkMigrationM3Arithmetic(b *testing.B) {
	var left, right [32]uint128.Uint128
	for j := 0; j < len(left); j++ {
		left[j] = uint128.Uint128{Lo: 0xa7839b428dc5 + uint64(j)*31, Hi: 0x1200 + uint64(j)}
		right[j] = uint128.Uint128{Lo: 0x79123 + uint64(j)*13, Hi: 3}
	}
	b.Run("Add", func(b *testing.B) {
		b.ReportAllocs()
		var r uint128.Uint128
		for i := 0; i < b.N; i++ {
			r, _ = left[i&31].Add(right[i&31])
		}
		m3BenchmarkSink = r
	})
	b.Run("MulExact", func(b *testing.B) {
		b.ReportAllocs()
		var r uint128.Uint128
		for i := 0; i < b.N; i++ {
			r, _ = left[i&31].Mul64(17)
		}
		m3BenchmarkSink = r
	})
	b.Run("MulOverflow", func(b *testing.B) {
		b.ReportAllocs()
		var r uint128.Uint128
		for i := 0; i < b.N; i++ {
			r, _ = left[i&31].Mul(right[i&31])
		}
		m3BenchmarkSink = r
	})
	b.Run("QuoRem", func(b *testing.B) {
		b.ReportAllocs()
		var r uint128.Uint128
		for i := 0; i < b.N; i++ {
			r, _, _ = left[i&31].QuoRem(right[i&31])
		}
		m3BenchmarkSink = r
	})
	b.Run("WideQuoRem", func(b *testing.B) {
		b.ReportAllocs()
		var r uint128.Uint128
		for i := 0; i < b.N; i++ {
			r, _, _ = uint128.QuoRem256By128(left[i&31], uint128.One, right[i&31])
		}
		m3BenchmarkSink = r
	})
}

// TestMigrationM3MulSuccessAndOverflow specifically covers both branches of
// Uint128.Mul. Generic uniformly random 128-bit operands mostly overflow and
// can mask a regression in successful full-width multiplication.
func TestMigrationM3MulSuccessAndOverflow(t *testing.T) {
	rng := rand.New(rand.NewSource(0x4d334d554c))
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	exactCount, overflowCount := 0, 0
	for i := 0; i < 6000; i++ {
		var left, right uint128.Uint128
		switch i % 6 {
		case 0: // Arbitrary 64 x 64 always fits into 128 bits.
			left, right = uint128.FromUint64(rng.Uint64()), uint128.FromUint64(rng.Uint64())
		case 1: // Full-width operands with a small multiplier.
			left, right = uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64() >> 2}, uint128.FromUint64(uint64(1+i%3))
		case 2: // Two nonzero upper limbs necessarily overflow.
			left = uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64() | 1}
			right = uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64() | 1}
		case 3: // Boundary full width, identity.
			left, right = uint128.Max, uint128.One
		case 4: // Boundary full width, overflow.
			left, right = uint128.Max, uint128.FromUint64(2)
		default: // Zero must not overflow, even against the maximum.
			left, right = uint128.Zero, uint128.Max
		}
		exact := new(big.Int).Mul(left.BigInt(), right.BigInt())
		got, code := left.Mul(right)
		if exact.Cmp(max) > 0 {
			overflowCount++
			if code != errors.Overflow || got != uint128.Zero {
				t.Fatalf("overflow #%d: %v * %v = %v (%v)", i, left, right, got, code)
			}
		} else {
			exactCount++
			if code != errors.None || got.BigInt().Cmp(exact) != 0 {
				t.Fatalf("exact #%d: %v * %v = %v (%v), want %s", i, left, right, got, code, exact)
			}
		}
	}
	if exactCount == 0 || overflowCount == 0 {
		t.Fatalf("missing branch: exact=%d overflow=%d", exactCount, overflowCount)
	}
}
