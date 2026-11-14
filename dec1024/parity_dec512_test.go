package dec1024

import (
	"math/rand"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec512"
)

// Values that fit the narrower backend must produce identical visible results
// and status flags across all seven rounding modes.
func TestDec512ParityForSharedDomain(t *testing.T) {
	rnd := rand.New(rand.NewSource(20261114))
	for i := 0; i < 800; i++ {
		a := dec512.DecodeFromUint64(uint64(rnd.Int63n(10000000000000)), uint8(rnd.Intn(6)))
		b := dec512.DecodeFromUint64(uint64(rnd.Int63n(10000000000000)+1), uint8(rnd.Intn(6)))
		if rnd.Intn(2) == 0 {
			a = a.Neg()
		}
		if rnd.Intn(2) == 0 {
			b = b.Neg()
		}
		x, y := FromDec512(a), FromDec512(b)
		for _, tc := range []struct {
			name string
			old  dec512.Dec512
			next Dec1024
		}{
			{"add", a.Add(b), x.Add(y)}, {"sub", a.Sub(b), x.Sub(y)},
			{"mul", a.Mul(b), x.Mul(y)},
		} {
			if tc.old.IsNaN() != tc.next.IsNaN() || (!tc.old.IsNaN() && tc.old.StringFixed() != tc.next.StringFixed()) {
				t.Fatalf("%s parity[%d]: 512=%s 1024=%s", tc.name, i, tc.old.StringFixed(), tc.next.StringFixed())
			}
		}
		scale := uint8(rnd.Intn(8))
		for mode := RoundingTowardZero; mode <= RoundingHalfEven; mode++ {
			o, oi := a.MulRoundWithStatus(b, scale, mode)
			n, ni := x.MulRoundWithStatus(y, uint16(scale), mode)
			if oi != ni || o.IsNaN() != n.IsNaN() || (!o.IsNaN() && o.StringFixed() != n.StringFixed()) {
				t.Fatalf("mulround parity[%d] scale=%d mode=%d: 512=%s,%v 1024=%s,%v", i, scale, mode, o, oi, n, ni)
			}
			o, oi = a.DivRoundWithStatus(b, scale, mode)
			n, ni = x.DivRoundWithStatus(y, uint16(scale), mode)
			if oi != ni || o.IsNaN() != n.IsNaN() || (!o.IsNaN() && o.StringFixed() != n.StringFixed()) {
				t.Fatalf("divround parity[%d] scale=%d mode=%d: 512=%s,%v 1024=%s,%v", i, scale, mode, o, oi, n, ni)
			}
		}
	}
}
