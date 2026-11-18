package uint128_test

import (
	"math/rand"
	"testing"

	"github.com/jokruger/dec128/state"
	upstream "github.com/jokruger/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	facade "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// TestMigrationM3DivisionStatus is intentionally compiled in both build modes.
// It tests the M3 adapter's operation-specific error translation against the
// exact states emitted by the pinned Jokruger v1.4.0 unsigned operations.
func TestMigrationM3DivisionStatus(t *testing.T) {
	toUp := func(v facade.Uint128) upstream.Uint128 { return upstream.Uint128(v) }
	toFacade := func(v upstream.Uint128) facade.Uint128 { return facade.Uint128(v) }
	mapStatus := func(s state.State) errors.Error {
		switch s {
		case state.OK:
			return errors.None
		case state.DivisionByZero:
			return errors.DivisionByZero
		case state.Overflow:
			return errors.Overflow
		default:
			t.Fatalf("upstream returned unexpected unsigned division state %v", s)
			return errors.NotANumber
		}
	}
	rng := rand.New(rand.NewSource(0x4d33445f444956))
	values := []facade.Uint128{
		facade.Zero,
		facade.One,
		facade.Max64,
		facade.Max,
		{Lo: 7, Hi: 1},
		{Lo: ^uint64(0), Hi: 1 << 63},
	}
	for i := 0; i < 2000; i++ {
		a := facade.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		b := facade.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		if i%7 == 0 {
			b = facade.Zero
		}
		if i%7 == 1 {
			b.Hi = 0
		}
		if i%7 == 2 {
			a, b = values[i%len(values)], values[(i+2)%len(values)]
		}
		if i%7 == 3 {
			a, b = facade.Max, facade.One
		}
		ua, ub := toUp(a), toUp(b)

		q, r, e := a.QuoRem(b)
		uq, ur, us := ua.QuoRem(ub)
		if q != toFacade(uq) || r != toFacade(ur) || e != mapStatus(us) {
			t.Fatalf("QuoRem case %d: %v/%v: got %v,%v,%v upstream %v,%v,%v", i, a, b, q, r, e, uq, ur, us)
		}
		fq, fe := a.Div(b)
		uqOnly, usOnly := ua.Div(ub)
		if fq != toFacade(uqOnly) || fe != mapStatus(usOnly) {
			t.Fatalf("Div case %d: %v,%v vs %v,%v", i, fq, fe, uqOnly, usOnly)
		}
		fr, fre := a.Mod(b)
		urOnly, urState := ua.Mod(ub)
		if fr != toFacade(urOnly) || fre != mapStatus(urState) {
			t.Fatalf("Mod case %d: %v,%v vs %v,%v", i, fr, fre, urOnly, urState)
		}

		divisor64 := b.Lo
		if i%5 == 0 {
			divisor64 = 0
		}
		q64, r64, e64 := a.QuoRem64(divisor64)
		uq64, ur64, us64 := ua.QuoRem64(divisor64)
		if q64 != toFacade(uq64) || r64 != ur64 || e64 != mapStatus(us64) {
			t.Fatalf("QuoRem64 case %d: %v,%d,%v vs %v,%d,%v", i, q64, r64, e64, uq64, ur64, us64)
		}
		fq64, fe64 := a.Div64(divisor64)
		uq64Only, uq64State := ua.Div64(divisor64)
		if fq64 != toFacade(uq64Only) || fe64 != mapStatus(uq64State) {
			t.Fatalf("Div64 case %d: %v,%v vs %v,%v", i, fq64, fe64, uq64Only, uq64State)
		}
		fr64, fre64 := a.Mod64(divisor64)
		ur64Only, ur64State := ua.Mod64(divisor64)
		if fr64 != ur64Only || fre64 != mapStatus(ur64State) {
			t.Fatalf("Mod64 case %d: %v,%v vs %v,%v", i, fr64, fre64, ur64Only, ur64State)
		}

		carry := facade.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		if i%5 == 0 {
			carry = facade.Zero
		}
		if i%5 == 1 {
			carry = b
		}
		if i%5 == 2 {
			carry = facade.One
		}
		wq, wr, we := facade.QuoRem256By128(a, carry, b)
		uwq, uwr, uws := upstream.QuoRem256By128(ua, toUp(carry), ub)
		if wq != toFacade(uwq) || wr != toFacade(uwr) || we != mapStatus(uws) {
			t.Fatalf("wide case %d: %v,%v,%v vs %v,%v,%v", i, wq, wr, we, uwq, uwr, uws)
		}
		carry64 := uint64(0)
		if i%3 == 0 {
			carry64 = divisor64
		}
		if i%3 == 1 {
			carry64 = divisor64 / 2
		}
		wq64, wr64, we64 := facade.QuoRem192By64(a, carry64, divisor64)
		uwq64, uwr64, uws64 := upstream.QuoRem192By64(ua, carry64, divisor64)
		if wq64 != toFacade(uwq64) || wr64 != uwr64 || we64 != mapStatus(uws64) {
			t.Fatalf("wide64 case %d: %v,%d,%v vs %v,%d,%v", i, wq64, wr64, we64, uwq64, uwr64, uws64)
		}
	}
}
