package uint128_test

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	bu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func TestM3FCodecParity(t *testing.T) {
	status := func(s state.State) errors.Error {
		switch s {
		case state.OK:
			return errors.None
		case state.Overflow:
			return errors.Overflow
		case state.InvalidFormat:
			return errors.InvalidFormat
		default:
			t.Fatalf("unexpected upstream status %v", s)
			return errors.NotANumber
		}
	}
	rng := rand.New(rand.NewSource(20261117))
	values := []string{"", "0", "0000", "+1", "-1", " 1", "1 ", "18446744073709551615", "340282366920938463463374607431768211455", "340282366920938463463374607431768211456", "000000000000000000000000000000000000000000000000000000000000000000000000", "00000000000000000000000000000000000000000000000000000000000000000000000A"}
	for i := 0; i < 35000; i++ {
		n := rng.Intn(95)
		var b strings.Builder
		for k := 0; k < n; k++ {
			v := rng.Intn(26)
			if v < 21 {
				b.WriteByte(byte('0' + rng.Intn(10)))
			} else {
				b.WriteByte("ab+-x"[v-21])
			}
		}
		values = append(values, b.String())
	}
	// Valid unsigned coefficients and their zero-padded forms are important:
	// malformed random input alone mostly measures rejected conversions.
	for i := 0; i < 24000; i++ {
		u := bu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		values = append(values, u.String(), strings.Repeat("0", i%17)+u.String(), strconv.FormatUint(uint64(i)*100003, 10))
	}
	for i, s := range values {
		got, err := bu.FromString(s)
		want, st := ju.FromString(s)
		if got != bu.Uint128(want) || err != status(st) {
			t.Fatalf("parse case %d %q got=%+v err=%v up=%+v state=%v", i, s, got, err, want, st)
		}
		gotBytes, errBytes := bu.FromString([]byte(s))
		if gotBytes != got || errBytes != err {
			t.Fatalf("byte variant case %d", i)
		}
		if err != errors.None {
			continue
		}
		var lbuf, ubuf [39]byte
		lg := got.StringToBuf(lbuf[:])
		ug := ju.Uint128(got).StringToBuf(ubuf[:])
		if string(lg) != string(ug) || got.String() != ju.Uint128(got).String() {
			t.Fatalf("format case %d got %q want %q", i, lg, ug)
		}
	}
}

// The upstream wide-division error ordering is part of the compatibility
// contract: with a nonzero upper half and a zero divisor it reports overflow,
// but with a zero upper half it reports division by zero.
func TestM3FWideDivisionErrorPrecedence(t *testing.T) {
	cases := []struct {
		lo, carry, div bu.Uint128
		want           errors.Error
	}{
		{bu.One, bu.Zero, bu.Zero, errors.DivisionByZero},
		{bu.One, bu.One, bu.Zero, errors.Overflow},
		{bu.One, bu.One, bu.One, errors.Overflow},
		{bu.One, bu.Uint128{Lo: 2}, bu.Uint128{Lo: 3}, errors.None},
		{bu.One, bu.Uint128{Hi: 1}, bu.Uint128{Hi: 1}, errors.Overflow},
	}
	for i, tc := range cases {
		q, r, e := bu.QuoRem256By128(tc.lo, tc.carry, tc.div)
		uQ, uR, s := ju.QuoRem256By128(ju.Uint128(tc.lo), ju.Uint128(tc.carry), ju.Uint128(tc.div))
		if q != bu.Uint128(uQ) || r != bu.Uint128(uR) || e != tc.want {
			t.Fatalf("case %d got %v,%v,%v upstream %v,%v,%v", i, q, r, e, uQ, uR, s)
		}
	}
}
