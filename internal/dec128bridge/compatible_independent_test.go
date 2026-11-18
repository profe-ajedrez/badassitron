package dec128bridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func TestIndependentCompatibleCodecsDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(20261117))
	inputs := []string{"", "-0", "-0.0000000000000000000", "+0.00", ".5", "5.", "1e20", "+", "-", ".", "--2", "1..2", "340282366920938463463374607431768211455", "340282366920938463463374607431768211456", "00000000000000000000000000000000000000000000001", "12345678901234567890.", "-12345678901234567890.", "0.00000000000000000000", "NaN"}
	for i := 0; i < 20000; i++ {
		n := rng.Intn(49)
		var b strings.Builder
		if rng.Intn(4) == 0 {
			b.WriteByte('-')
		} else if rng.Intn(11) == 0 {
			b.WriteByte('+')
		}
		dot := rng.Intn(n + 2)
		for j := 0; j < n; j++ {
			if j == dot {
				b.WriteByte('.')
			}
			b.WriteByte(byte('0' + rng.Intn(10)))
		}
		if dot == n {
			b.WriteByte('.')
		}
		inputs = append(inputs, b.String())
	}
	for i, input := range inputs {
		a := legacy.FromString(input)
		b := ParseCompatible(input)
		if a.IsNaN() != b.IsNaN() || a.StringFixed() != b.StringFixed() || (a.IsNaN() && a.ErrorDetails() != b.Legacy().ErrorDetails()) || (!a.IsNaN() && (a.Precision() != b.Legacy().Precision() || a.StoredNegative() != b.Legacy().StoredNegative())) {
			t.Fatalf("case %d input=%q old=%s scale=%d error=%v; bridge=%s scale=%d error=%v", i, input, a.StringFixed(), a.Precision(), a.ErrorDetails(), b.StringFixed(), b.Legacy().Precision(), b.Legacy().ErrorDetails())
		}
		at, _ := a.MarshalText()
		bt, _ := b.MarshalText()
		if !bytes.Equal(at, bt) {
			t.Fatalf("text case %d %q %q %q", i, input, at, bt)
		}
		aj, _ := json.Marshal(a)
		bj, _ := json.Marshal(b)
		if !bytes.Equal(aj, bj) {
			t.Fatalf("json case %d %q %s %s", i, input, aj, bj)
		}
		av, _ := a.Value()
		bv, _ := b.Value()
		if av != bv {
			t.Fatalf("sql value case %d", i)
		}
	}
}
func TestIndependentCompatibleScanDifferential(t *testing.T) {
	for _, x := range []any{nil, int(0), int(-23), int64(-9223372036854775808), int64(9223372036854775807), float32(1.234567), float32(-0.00000001), float64(1.234567890123), float64(1e20), "-0.001", []byte("5.5"), "bad", []byte("bad"), true} {
		a := legacy.Zero
		ae := legacy.ScanInto(&a, x)
		b := ParseCompatible("17.25")
		be := b.Scan(x)
		if (ae == nil) != (be == nil) {
			t.Fatalf("scan %T %v: %v vs %v", x, x, ae, be)
		}
		if ae == nil && (a.StringFixed() != b.StringFixed() || a.StoredNegative() != b.Legacy().StoredNegative()) {
			t.Fatalf("scan %T %v: %s vs %s", x, x, a.StringFixed(), b.StringFixed())
		}
		if be != nil && b.StringFixed() != "17.25" {
			t.Fatalf("scan mutated on error %v", x)
		}
	}
}
func TestIndependentCompatibleInputRollback(t *testing.T) {
	for _, invalid := range []string{"bad", "1e20", "1..2", strings.Repeat("9", 42)} {
		v := ParseCompatible("-0.001")
		if err := v.UnmarshalText([]byte(invalid)); err == nil {
			t.Fatalf("expected error: %q", invalid)
		}
		if v.StringFixed() != "-0.001" || !v.Legacy().StoredNegative() {
			t.Fatalf("mutated after %q", invalid)
		}
	}
	for _, src := range []any{float64(12.5), float32(0.25), int64(-10), strconv.FormatInt(-125, 10), []byte("-0.00")} {
		v := ParseCompatible("2")
		if err := v.Scan(src); err != nil {
			t.Fatalf("%s: %v", fmt.Sprint(src), err)
		}
	}
}
