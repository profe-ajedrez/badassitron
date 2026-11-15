package dec256

import (
	"github.com/profe-ajedrez/badassitron/dec128"
	"testing"
)

func TestConversionDec128RoundTrip(t *testing.T) {
	src := dec128.FromString("-123456789.12345000")
	wide := FromDec128(src)
	if !wide.Equal(FromString(src.String())) {
		t.Fatalf("widen changed value: %v", wide)
	}
	narrow, e := wide.Dec128Exact()
	if e != nil || !narrow.Equal(src) {
		t.Fatal(narrow, e)
	}
	// Canonical trailing zero reduction enables exact narrowing.
	reducible := FromString("1.200000000000000000000000000000")
	if _, e := reducible.Dec128Exact(); e != nil {
		t.Fatal(e)
	}
	if _, e := FromString("0.000000000000000000000001").Dec128Exact(); e == nil {
		t.Fatal("lossy narrowing was accepted")
	}
}
