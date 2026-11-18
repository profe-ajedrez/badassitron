package dec128bridge

import (
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	"testing"
)

func TestDivisionNegativeZeroAtScaleZero(t *testing.T) {
	a, b := legacy.FromString("-1"), legacy.FromString("1000")
	legacy.SetDefaultPrecision(0)
	defer legacy.SetDefaultPrecision(19)
	want := a.Div(b)
	got := DivAtPrecision(a, b, 0)
	if !want.IsZero() || !want.StoredNegative() || want.Precision() != 0 {
		t.Fatalf("legacy contract changed: %v scale=%d sign=%v", want, want.Precision(), want.StoredNegative())
	}
	if !got.IsZero() || !got.StoredNegative() || got.Precision() != want.Precision() {
		t.Fatalf("bridge did not retain stored zero: got=%v precision=%d negative=%v", got, got.Precision(), got.StoredNegative())
	}
}
