package dec128

import (
	"testing"
	"unsafe"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func TestV4StateLayout(t *testing.T) {
	if got := unsafe.Sizeof(Dec128{}); got != 24 {
		t.Fatalf("Dec128 is %d bytes, want 24", got)
	}
	for code := 0; code <= 255; code++ {
		reason := errors.Error(code)
		for _, neg := range []bool{false, true} {
			d := Dec128{coef: uint128.Uint128{Lo: 777, Hi: 9}, exp: 7, state: combinedState(neg, reason)}
			if got := d.errorCode(); got != reason {
				t.Fatalf("error=%d, got %d", reason, got)
			}
			if got := d.negative(); got != neg {
				t.Fatalf("neg=%v, got %v", neg, got)
			}
			if got := d.IsNaN(); got != (reason != errors.None) {
				t.Fatalf("isNaN mismatch")
			}
		}
	}
	// All exported errors.Error values must round trip through NaN, including
	// values reserved for future error codes. A uint8 packed state silently
	// dropped bit 7 for codes >= 128, turning some NaNs into valid values.
	for code := 1; code <= 255; code++ {
		reason := errors.Error(code)
		d := NaN(reason)
		if !d.IsNaN() || d.errorCode() != reason {
			t.Fatalf("NaN(%d) was truncated: state=%d decoded=%d", code, d.state, d.errorCode())
		}
		if got := d.Add(One); got.errorCode() != reason {
			t.Fatalf("Add must propagate error %d, got %d", code, got.errorCode())
		}
		if got := One.Sub(d); got.errorCode() != reason {
			t.Fatalf("Sub must propagate error %d, got %d", code, got.errorCode())
		}
		if got := d.Mul(One); got.errorCode() != reason {
			t.Fatalf("Mul must propagate error %d, got %d", code, got.errorCode())
		}
		if got := d.Canonical(); got.errorCode() != reason {
			t.Fatalf("Canonical must preserve error %d, got %d", code, got.errorCode())
		}
	}
	for _, prec := range []uint8{1, 2, 9, 19} {
		z := New(uint128.Zero, prec, true)
		if !z.negative() || !z.IsZero() || z.IsNegative() || z.IsNaN() {
			t.Fatalf("signed zero lost: %+v", z)
		}
		if z.Compare(New(uint128.Zero, prec, false)) != 0 {
			t.Fatalf("signed zero compare")
		}
	}
}
