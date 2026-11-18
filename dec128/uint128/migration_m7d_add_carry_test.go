package uint128_test

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	local "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// Add uses Jokruger's exported AddCarry to remove a status-code conversion.
// Check its carry, zero-on-overflow, and public errors against Jokruger's Add.
func TestM7DAddCarryCompatibility(t *testing.T) {
	rng := rand.New(rand.NewSource(0x7d4addca))
	edge := []ju.Uint128{
		{}, {Lo: 1}, {Lo: 2}, {Lo: ^uint64(0)}, {Hi: 1},
		{Lo: ^uint64(0), Hi: ^uint64(0)},
		{Lo: 0, Hi: ^uint64(0)},
		{Lo: ^uint64(0), Hi: ^uint64(0) - 1},
		{Lo: 1, Hi: ^uint64(0)},
	}
	for i := 0; i < 24000; i++ {
		a := ju.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		b := ju.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		if i%3 == 0 {
			a = edge[(i/3)%len(edge)]
		}
		if i%5 == 0 {
			b = edge[(i/5)%len(edge)]
		}
		if i%11 == 0 {
			a.Hi = 0
		}
		if i%13 == 0 {
			b.Hi = 0
		}
		la, lb := local.Uint128(a), local.Uint128(b)
		result, code := la.Add(lb)
		ref, status := a.Add(b)
		carryValue, carry := a.AddCarry(b)
		if carry > 1 {
			t.Fatalf("case %d: undocumented carry=%d", i, carry)
		}
		expectedError := errors.None
		switch status {
		case state.OK:
			if carry != 0 || carryValue != ref {
				t.Fatalf("case %d: exact carry/add mismatch", i)
			}
		case state.Overflow:
			expectedError = errors.Overflow
			if carry != 1 || ref != ju.Zero {
				t.Fatalf("case %d: overflow carry/add mismatch", i)
			}
		default:
			t.Fatalf("case %d: unexpected upstream status %d", i, status)
		}
		if result != local.Uint128(ref) || code != expectedError {
			t.Fatalf("case %d: result=%v code=%d, expected=%v code=%d", i, result, code, ref, expectedError)
		}
	}
}

// Independently verify addition and its error boundary without depending on
// Jokruger's status mapping or implementation details.
func TestM7DAddIndependentMathBig(t *testing.T) {
	rng := rand.New(rand.NewSource(0x7d4f1b19))
	toBig := func(u local.Uint128) *big.Int {
		hi := new(big.Int).Lsh(new(big.Int).SetUint64(u.Hi), 64)
		return hi.Add(hi, new(big.Int).SetUint64(u.Lo))
	}
	for i := 0; i < 4096; i++ {
		a := local.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		b := local.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		if i%7 == 0 {
			a = local.Max
		}
		if i%11 == 0 {
			b = local.One
		}
		got, status := a.Add(b)
		sum := new(big.Int).Add(toBig(a), toBig(b))
		if sum.BitLen() > 128 {
			if got != local.Zero || status != errors.Overflow {
				t.Fatalf("case %d: incorrect overflow result/code", i)
			}
		} else if gotBig := toBig(got); status != errors.None || gotBig.Cmp(sum) != 0 {
			t.Fatalf("case %d: exact sum=%s status=%d, expected=%s", i, gotBig, status, sum)
		}
	}
}

// The imported wide divider has precedence rules that an apparently harmless
// local early check can change. Pin the observable legacy mapping of its output.
func TestM7DWideDivisionErrorPrecedence(t *testing.T) {
	cases := []struct {
		name                string
		low, carry, divisor local.Uint128
		want                errors.Error
	}{
		{"zero carry and zero divisor", local.One, local.Zero, local.Zero, errors.DivisionByZero},
		{"low carry and zero divisor", local.One, local.One, local.Zero, errors.Overflow},
		{"high carry and zero divisor", local.One, local.Uint128{Hi: 1}, local.Zero, errors.Overflow},
		{"equal carry and divisor", local.One, local.Uint128{Lo: 3, Hi: 1}, local.Uint128{Lo: 3, Hi: 1}, errors.Overflow},
		{"zero carry and nonzero divisor", local.One, local.Zero, local.One, errors.None},
		{"lesser carry than divisor", local.One, local.One, local.Uint128{Lo: 3, Hi: 1}, errors.None},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q, r, s := local.QuoRem256By128(tc.low, tc.carry, tc.divisor)
			uq, ur, us := ju.QuoRem256By128(ju.Uint128(tc.low), ju.Uint128(tc.carry), ju.Uint128(tc.divisor))
			if q != local.Uint128(uq) || r != local.Uint128(ur) || s != tc.want {
				t.Fatalf("facade q=%v r=%v s=%d; upstream q=%v r=%v s=%d, expected status=%d", q, r, s, uq, ur, us, tc.want)
			}
		})
	}
}
