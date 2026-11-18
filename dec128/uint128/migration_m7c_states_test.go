package uint128_test

import (
	"math/rand"
	"testing"

	"github.com/jokruger/dec128/state"
	up "github.com/jokruger/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	local "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// The release adapter's branchless error conversion is deliberately specific
// to the published Jokruger v1.4.0 state codes. Fail rather than silently
// returning a wrong legacy error if either set of constants changes.
func TestReleaseM7CStateEncoding(t *testing.T) {
	for _, tc := range []struct {
		name      string
		got, want uint8
	}{
		{"upstream OK", uint8(state.OK), 0},
		{"upstream Overflow", uint8(state.Overflow), 5},
		{"upstream Underflow", uint8(state.Underflow), 6},
		{"upstream DivisionByZero", uint8(state.DivisionByZero), 4},
		{"legacy None", uint8(errors.None), 0},
		{"legacy Overflow", uint8(errors.Overflow), 3},
		{"legacy Underflow", uint8(errors.Underflow), 4},
		{"legacy DivisionByZero", uint8(errors.DivisionByZero), 2},
	} {
		if tc.got != tc.want {
			t.Fatalf("%s: got %d, want %d; review optimized state mapping", tc.name, tc.got, tc.want)
		}
	}
}

func TestReleaseM7CImportedArithmeticAndErrors(t *testing.T) {
	rng := rand.New(rand.NewSource(0x4d37435f53544154))
	toLocal := func(x up.Uint128) local.Uint128 { return local.Uint128(x) }
	check := func(kind string, i int, actual local.Uint128, code errors.Error, reference up.Uint128, want errors.Error) {
		t.Helper()
		if actual != toLocal(reference) || code != want {
			t.Fatalf("%s #%d: result=%v error=%d; upstream result=%v mapped error=%d", kind, i, actual, code, reference, want)
		}
	}
	mapped := func(s state.State) errors.Error {
		switch s {
		case state.OK:
			return errors.None
		case state.Overflow:
			return errors.Overflow
		case state.Underflow:
			return errors.Underflow
		case state.DivisionByZero:
			return errors.DivisionByZero
		default:
			t.Fatalf("unexpected upstream unsigned state: %d", s)
			return errors.NotANumber
		}
	}
	values := []up.Uint128{
		{}, {Lo: 1}, {Lo: 2}, {Lo: ^uint64(0)}, {Hi: 1},
		{Lo: ^uint64(0), Hi: ^uint64(0)}, {Lo: 0, Hi: ^uint64(0)},
		{Lo: ^uint64(0), Hi: ^uint64(0) - 1},
	}
	for i := 0; i < 12000; i++ {
		a := up.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		b := up.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		if i%3 == 0 {
			a = values[(i/3)%len(values)]
		}
		if i%5 == 0 {
			b = values[(i/5)%len(values)]
		}
		if i%7 == 0 {
			b.Hi = 0
		}
		if i%11 == 0 {
			a.Hi = 0
		}
		if i%13 == 0 {
			b = up.Zero
		}
		la, lb := toLocal(a), toLocal(b)

		ref, s := a.Add(b)
		got, e := la.Add(lb)
		check("Add", i, got, e, ref, mapped(s))
		ref, s = a.Add64(b.Lo)
		got, e = la.Add64(b.Lo)
		check("Add64", i, got, e, ref, mapped(s))
		ref, s = a.Sub(b)
		got, e = la.Sub(lb)
		check("Sub", i, got, e, ref, mapped(s))
		ref, s = a.Sub64(b.Lo)
		got, e = la.Sub64(b.Lo)
		check("Sub64", i, got, e, ref, mapped(s))
		ref, s = a.Mul(b)
		got, e = la.Mul(lb)
		check("Mul", i, got, e, ref, mapped(s))
		ref, s = a.Mul64(b.Lo)
		got, e = la.Mul64(b.Lo)
		check("Mul64", i, got, e, ref, mapped(s))

		ref, s = a.Div(b)
		got, e = la.Div(lb)
		check("Div", i, got, e, ref, mapped(s))
		ref, s = a.Mod(b)
		got, e = la.Mod(lb)
		check("Mod", i, got, e, ref, mapped(s))
		refQ, refR, refState := a.QuoRem(b)
		q, r, code := la.QuoRem(lb)
		check("QuoRem quotient", i, q, code, refQ, mapped(refState))
		if r != toLocal(refR) {
			t.Fatalf("QuoRem remainder #%d: %v, expected %v", i, r, refR)
		}
		refQ, refRem64, refState := a.QuoRem64(b.Lo)
		q, rem64, code := la.QuoRem64(b.Lo)
		check("QuoRem64 quotient", i, q, code, refQ, mapped(refState))
		if rem64 != refRem64 {
			t.Fatalf("QuoRem64 remainder #%d: %d, expected %d", i, rem64, refRem64)
		}

		carry := up.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		if i%4 == 0 {
			carry = up.Zero
		}
		if i%6 == 0 {
			carry = b
		}
		wq, wr, ws := up.QuoRem256By128(a, carry, b)
		lq, lr, le := local.QuoRem256By128(la, toLocal(carry), lb)
		check("QuoRem256By128 quotient", i, lq, le, wq, mapped(ws))
		if lr != toLocal(wr) {
			t.Fatalf("QuoRem256By128 remainder #%d: %v, expected %v", i, lr, wr)
		}
		wq, wr64, ws := up.QuoRem192By64(a, carry.Lo, b.Lo)
		lq, lr64, le := local.QuoRem192By64(la, carry.Lo, b.Lo)
		check("QuoRem192By64 quotient", i, lq, le, wq, mapped(ws))
		if lr64 != wr64 {
			t.Fatalf("QuoRem192By64 remainder #%d: %d, expected %d", i, lr64, wr64)
		}
	}
}
