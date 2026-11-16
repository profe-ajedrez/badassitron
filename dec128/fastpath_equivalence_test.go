package dec128

import (
	"math/rand"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// These reference implementations intentionally preserve the unmodified
// public-method control flow. They are independent of the fast paths.
func referenceAdd(a, b Dec128) Dec128 {
	if a.errorCode() != errors.None {
		return a
	}
	if b.errorCode() != errors.None {
		return b
	}
	r, ok := a.tryAdd(b)
	if ok {
		return r
	}
	r, ok = a.Canonical().tryAdd(b.Canonical())
	if ok {
		return r
	}
	return NaN(errors.Overflow)
}

func referenceSub(a, b Dec128) Dec128 {
	if a.errorCode() != errors.None {
		return a
	}
	if b.errorCode() != errors.None {
		return b
	}
	r, ok := a.trySub(b)
	if ok {
		return r
	}
	r, ok = a.Canonical().trySub(b.Canonical())
	if ok {
		return r
	}
	return NaN(errors.Overflow)
}

func referenceMul(a, b Dec128) Dec128 {
	if a.errorCode() != errors.None {
		return a
	}
	if b.errorCode() != errors.None {
		return b
	}
	if a.IsZero() || b.IsZero() {
		return Zero
	}
	r, ok := a.tryMul(b)
	if ok {
		return r
	}
	r, ok = a.Canonical().tryMul(b.Canonical())
	if ok {
		return r
	}
	return NaN(errors.Overflow)
}

func referenceCompare(a, b Dec128) int {
	if a.errorCode() != errors.None && b.errorCode() != errors.None {
		return 0
	}
	if a.errorCode() != errors.None {
		return -1
	}
	if b.errorCode() != errors.None {
		return 1
	}
	if a.coef.IsZero() {
		if b.coef.IsZero() {
			return 0
		}
		if b.negative() {
			return 1
		}
		return -1
	}
	if b.coef.IsZero() {
		if a.negative() {
			return -1
		}
		return 1
	}
	if a.negative() != b.negative() {
		if a.negative() {
			return -1
		}
		return 1
	}
	cmp := compareMagnitude(a, b)
	if a.negative() {
		return -cmp
	}
	return cmp
}

func TestFastPathDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(20261115))
	special := []uint128.Uint128{
		{}, {Lo: 1}, {Lo: 10}, {Lo: 100}, {Lo: 100000000},
		{Lo: ^uint64(0)}, {Lo: 0, Hi: 1}, {Lo: 1, Hi: 1},
		{Lo: ^uint64(0), Hi: ^uint64(0)},
		{Lo: ^uint64(0) - 1, Hi: ^uint64(0)},
		{Lo: 0, Hi: ^uint64(0)},
		{Lo: 1, Hi: 1 << 63},
	}
	vals := make([]Dec128, 0, 22000)
	for _, x := range special {
		for _, exp := range []uint8{0, 1, 2, 9, 18, 19} {
			for _, neg := range []bool{false, true} {
				vals = append(vals, Dec128{coef: x, exp: exp, state: signState(neg)})
			}
		}
	}
	for _, err := range []errors.Error{errors.NotANumber, errors.Overflow, errors.Underflow, errors.PrecisionOutOfRange} {
		vals = append(vals, Dec128{coef: special[3], exp: 4, state: combinedState(true, err)})
	}
	for i := 0; i < 22000; i++ {
		var c uint128.Uint128
		switch i % 5 {
		case 0:
			c = special[rng.Intn(len(special))]
		case 1:
			c = uint128.Uint128{Lo: rng.Uint64()}
		case 2:
			c = uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}
		case 3:
			c = uint128.Uint128{Lo: uint64(rng.Int63n(1000000000000000))}
		default:
			c = uint128.Uint128{Lo: uint64(rng.Int63n(100000000)), Hi: uint64(rng.Int63n(100000))}
		}
		vals = append(vals, Dec128{coef: c, exp: uint8(rng.Intn(20)), state: signState(rng.Intn(2) == 1)})
	}
	check := func(a, b Dec128) {
		t.Helper()
		if got, want := a.Add(b), referenceAdd(a, b); got != want {
			t.Fatalf("Add(%+v,%+v) got %+v, want %+v", a, b, got, want)
		}
		if got, want := a.Sub(b), referenceSub(a, b); got != want {
			t.Fatalf("Sub(%+v,%+v) got %+v, want %+v", a, b, got, want)
		}
		if got, want := a.Mul(b), referenceMul(a, b); got != want {
			t.Fatalf("Mul(%+v,%+v) got %+v, want %+v", a, b, got, want)
		}
		if got, want := a.Compare(b), referenceCompare(a, b); got != want {
			t.Fatalf("Compare(%+v,%+v) got %d, want %d", a, b, got, want)
		}
	}
	// Complete targeted cross product, including signed zeros and NaNs.
	for i := 0; i < 144; i++ {
		for j := 0; j < 144; j++ {
			check(vals[i], vals[j])
		}
	}
	// More than 40,000 reproducible randomized pairs; half of them share
	// exponent and sign, targeting the proposed fast paths.
	for i := 0; i < 40000; i++ {
		a := vals[rng.Intn(len(vals))]
		b := vals[rng.Intn(len(vals))]
		if i%2 == 0 {
			b.exp = a.exp
			b.state = (b.state &^ uint16(1)) | (a.state & 1)
		}
		check(a, b)
	}
}
