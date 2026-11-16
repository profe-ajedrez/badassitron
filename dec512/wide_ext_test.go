package dec512

import (
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

func fromWideBig(t *testing.T, v *big.Int, scale int, neg bool) Dec512 {
	t.Helper()
	c, e := uint512.FromBigInt(v)
	if e != 0 {
		t.Fatalf("wide int from big: %v", e)
	}
	return New(c, uint8(scale), neg)
}
func TestFullWidthDifferential(t *testing.T) {
	rnd := rand.New(rand.NewSource(2026111202 + 512))
	for i := 0; i < 1500; i++ {
		aBytes := make([]byte, 512/8)
		bBytes := make([]byte, 512/8)
		rnd.Read(aBytes)
		rnd.Read(bBytes)
		ca := new(big.Int).SetBytes(aBytes)
		cb := new(big.Int).SetBytes(bBytes)
		if cb.Sign() == 0 {
			cb.SetInt64(1)
		}
		sa := rnd.Intn(int(MaxPrecision) + 1)
		sb := rnd.Intn(int(MaxPrecision) + 1)
		a, b := fromWideBig(t, ca, sa, rnd.Intn(2) == 1), fromWideBig(t, cb, sb, rnd.Intn(2) == 1)
		if got, ref := a.Compare(b), rat(a).Cmp(rat(b)); got != ref {
			t.Fatalf("compare mismatch %s %s got=%d ref=%d", a, b, got, ref)
		}
		if got := a.Add(b); !got.IsNaN() {
			want := new(big.Rat).Add(rat(a), rat(b))
			if rat(got).Cmp(want) != 0 {
				t.Fatalf("full add mismatch %v %v -> %v != %s", a, b, got, want)
			}
		}
		if got := a.Sub(b); !got.IsNaN() {
			want := new(big.Rat).Sub(rat(a), rat(b))
			if rat(got).Cmp(want) != 0 {
				t.Fatalf("full sub mismatch %v %v -> %v != %s", a, b, got, want)
			}
		}
		if got := a.Mul(b); !got.IsNaN() {
			want := new(big.Rat).Mul(rat(a), rat(b))
			if rat(got).Cmp(want) != 0 {
				t.Fatalf("full mul mismatch %v %v -> %v != %s", a, b, got, want)
			}
		}
		target := rnd.Intn(1 + int(MaxPrecision))
		mode := RoundingMode(rnd.Intn(7))
		if got, ix := a.DivRoundWithStatus(b, uint8(target), mode); !got.IsNaN() {
			want, wantIx := scaledRound(new(big.Rat).Quo(rat(a), rat(b)), target, mode)
			if valueAtScale(got, target).Cmp(want) != 0 || ix != wantIx {
				t.Fatalf("full div mismatch a=%s b=%s target=%d mode=%v got=%s, ix=%v want=%s, ix=%v", a, b, target, mode, got, ix, want, wantIx)
			}
		}
		if got, ix := a.RescaleRoundWithStatus(uint8(target), mode); !got.IsNaN() {
			want, wantIx := scaledRound(rat(a), target, mode)
			if valueAtScale(got, target).Cmp(want) != 0 || ix != wantIx {
				t.Fatalf("full rescale mismatch %v target=%d mode=%v got=%s ix=%v want=%s ix=%v", a, target, mode, got, ix, want, wantIx)
			}
		}
		mod := a.Mod(b)
		if mod.IsNaN() {
			t.Fatalf("mod unexpectedly NaN: %v", mod.ErrorDetails())
		}
		if rat(mod.Abs()).Cmp(new(big.Rat).Abs(rat(b))) >= 0 {
			t.Fatalf("mod magnitude %v vs %v", mod, b)
		}
		if !mod.IsZero() && mod.Sign() != a.Sign() {
			t.Fatalf("mod sign")
		}
		// Independent modulo via integer alignment; bigint is test oracle only.
		scale := max(sa, sb)
		u, v := new(big.Int).Set(ca), new(big.Int).Set(cb)
		if sa < scale {
			u.Mul(u, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale-sa)), nil))
		}
		if sb < scale {
			v.Mul(v, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale-sb)), nil))
		}
		r := new(big.Int).Rem(u, v)
		if a.IsNegative() {
			r.Neg(r)
		}
		wanted := new(big.Rat).SetFrac(r, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil))
		if rat(mod).Cmp(wanted) != 0 {
			t.Fatalf("full mod mismatch %v %% %v: got=%v want=%v", a, b, mod, wanted)
		}
	}
}

func TestExactWidthAndScaleCanonicalFallback(t *testing.T) {
	max := New(uint512.Max, 1, false)
	// Sum of equal coefficients overflows, but trailing decimal zero can
	// be removed, resulting in a representable value at scale zero.
	value := max.Add(max)
	// Max=2^512-1 ends in a different digit by bit width; test only
	// mathematically sound equality for any accepted result.
	if !value.IsNaN() && rat(value).Cmp(new(big.Rat).Mul(rat(max), big.NewRat(2, 1))) != 0 {
		t.Fatal("incorrect canonical overflow fallback", value)
	}
	x := New(uint512.Max, 1, false)
	ten := FromInt(10)
	prod := x.Mul(ten)
	if prod.IsNaN() || !prod.Equal(New(uint512.Max, 0, false)) {
		t.Fatal("full-width product can cancel exact powers of ten", prod)
	}
	// A number with maximum scale and tiny coefficient remains representable.
	tiny := FromString("0." + strings.Repeat("0", int(MaxPrecision)-1) + "1")
	if tiny.IsNaN() || tiny.Mul(ten).IsNaN() {
		t.Fatal("high-scale exact multiplication")
	}
	// Integer maximum is representable, but it cannot be made bigger.
	if !New(uint512.Max, 0, false).Add(One).IsNaN() {
		t.Fatal("max overflow")
	}
}

func TestConcurrentPrecision(t *testing.T) {
	// The setter is process-wide by API compatibility; checks verify atomic
	// synchronization, not deterministic outcomes of concurrent divisions.
	const workers = 4
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for n := 0; n < 400; n++ {
				if i == 0 {
					SetDefaultPrecision(uint8(n % 20))
				} else {
					z := FromInt(1).Div(FromInt(3))
					if z.IsNaN() {
						t.Errorf("unexpected NaN during precision race")
					}
				}
			}
		}(i)
	}
	wg.Wait()
	SetDefaultPrecision(19)
}

func FuzzWideArithmetic(f *testing.F) {
	f.Add("1.1", "2.3", uint8(2))
	f.Add("-3.14", "1.23", uint8(5))
	f.Add("0.000000001", "9999999", uint8(19))
	f.Fuzz(func(t *testing.T, x, y string, p uint8) {
		if len(x) > 200 || len(y) > 200 {
			return
		}
		a, b := FromString(x), FromString(y)
		if a.IsNaN() || b.IsNaN() {
			return
		}
		if z := a.Add(b); !z.IsNaN() {
			want := new(big.Rat).Add(rat(a), rat(b))
			if rat(z).Cmp(want) != 0 {
				t.Fatal("add", a, b, z, want)
			}
		}
		if z := a.Mul(b); !z.IsNaN() {
			want := new(big.Rat).Mul(rat(a), rat(b))
			if rat(z).Cmp(want) != 0 {
				t.Fatal("mul", a, b, z, want)
			}
		}
		if !b.IsZero() {
			p %= MaxPrecision + 1
			z, _ := a.DivRoundWithStatus(b, p, RoundingHalfEven)
			if !z.IsNaN() {
				want, _ := scaledRound(new(big.Rat).Quo(rat(a), rat(b)), int(p), RoundingHalfEven)
				if valueAtScale(z, int(p)).Cmp(want) != 0 {
					t.Fatal("divround", strconv.Itoa(int(p)), a, b, z, want)
				}
			}
		}
	})
}
