package dec512

import (
	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec256"
	"testing"
)

func TestConversionRoundTrips(t *testing.T) {
	a := dec128.FromString("-17.1234000")
	if b, e := FromDec128(a).Dec128Exact(); e != nil || !b.Equal(a) {
		t.Fatal(b, e)
	}
	c := dec256.FromString("123456789.123456789012345678901234567890")
	if d, e := FromDec256(c).Dec256Exact(); e != nil || !d.Equal(c) {
		t.Fatal(d, e)
	}
	if _, e := FromString("0.000000000000000000000000000000000000000000000000000000000000000000000000000001").Dec256Exact(); e == nil {
		t.Fatal("accepted lossy narrowing")
	}
	if _, e := FromString("0.000000000000000000000000000000000000000000000000000000000000000000000000000001").Dec128Exact(); e == nil {
		t.Fatal("accepted lossy narrowing to 128")
	}
}
