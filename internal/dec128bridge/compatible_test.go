package dec128bridge

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"testing"

	ju "github.com/jokruger/dec128/uint128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func TestCompatiblePreservesSignedZeroMetadata(t *testing.T) {
	for scale := uint8(0); scale <= legacy.MaxPrecision; scale++ {
		for _, negative := range []bool{false, true} {
			before := legacy.New(lu.Zero, scale, negative)
			via := WrapLegacy(before)
			after := via.Legacy()
			if before.Precision() != after.Precision() || before.StoredNegative() != after.StoredNegative() || !after.IsZero() {
				t.Fatalf("scale=%d neg=%t: %#v to %#v", scale, negative, before, after)
			}
			built := NewCompatible(ju.Zero, scale, negative).Legacy()
			if built.Precision() != before.Precision() || built.StoredNegative() != before.StoredNegative() {
				t.Fatalf("constructor lost negative zero at scale %d neg=%t", scale, negative)
			}
			if via.String() != before.String() || via.StringFixed() != before.StringFixed() {
				t.Fatalf("zero formatting differs scale=%d neg=%t", scale, negative)
			}
		}
	}
}

func TestCompatibleRoundtripFiniteAndErrors(t *testing.T) {
	rng := rand.New(rand.NewSource(20261117))
	for i := 0; i < 20000; i++ {
		before := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		after := WrapLegacy(before).Legacy()
		if after.Coefficient() != before.Coefficient() || after.Precision() != before.Precision() || after.StoredNegative() != before.StoredNegative() || after.ErrorDetails() != before.ErrorDetails() {
			t.Fatalf("roundtrip %d: %#v %#v", i, before, after)
		}
	}
	for _, reason := range []le.Error{le.NotANumber, le.DivisionByZero, le.Overflow, le.Underflow, le.InvalidFormat, le.PrecisionOutOfRange, le.SqrtNegative} {
		before := legacy.NaN(reason)
		after := WrapLegacy(before).Legacy()
		if !after.IsNaN() || after.ErrorDetails() != before.ErrorDetails() {
			t.Fatalf("nan %v: %v", reason, after.ErrorDetails())
		}
	}
}

func TestCompatibleCodecsAndSQL(t *testing.T) {
	for _, input := range []string{"", "0", "-0.000", ".5", "5.0", "-123.456789", "1e20", "bad", "340282366920938463463374607431768211455"} {
		before := legacy.FromString(input)
		v := ParseCompatible(input)
		if v.IsNaN() != before.IsNaN() || v.Legacy().StringFixed() != before.StringFixed() || v.Legacy().ErrorDetails() != before.ErrorDetails() {
			t.Fatalf("parser %q mismatch", input)
		}
		rawText, err := v.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		expectedText, err := before.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(rawText, expectedText) {
			t.Fatalf("text %q: %q != %q", input, rawText, expectedText)
		}
		rawJSON, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		expectedJSON, err := json.Marshal(before)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(rawJSON, expectedJSON) {
			t.Fatalf("json %q: %s != %s", input, rawJSON, expectedJSON)
		}
		a, ea := v.Value()
		b, eb := before.Value()
		if ea != eb || a != b {
			t.Fatalf("sql value %q: %v %v", input, a, b)
		}
		if !before.IsNaN() {
			var parsed Compatible
			if err := parsed.UnmarshalJSON(rawJSON); err != nil {
				t.Fatal(err)
			}
			if parsed.String() != v.String() {
				t.Fatalf("json roundtrip %q", input)
			}
			if err := parsed.UnmarshalText(rawText); err != nil {
				t.Fatal(err)
			}
			if parsed.String() != v.String() {
				t.Fatalf("text roundtrip %q", input)
			}
			if err := parsed.Scan(before.StringFixed()); err != nil {
				t.Fatal(err)
			}
			if !parsed.Legacy().Equal(before) {
				t.Fatalf("SQL roundtrip %q", input)
			}
		}
	}
	v := ParseCompatible("12.50")
	if err := v.Scan("invalid"); err == nil {
		t.Fatal("Scan must reject invalid input")
	}
	if v.String() != "12.5" {
		t.Fatal("failed scan modified receiver")
	}
	var nilDst *Compatible
	if err := nilDst.Scan("1"); err == nil {
		t.Fatal("nil scan must return an error")
	}
}

func TestCompatibleSQLInputTypesAndFailureAtomicity(t *testing.T) {
	cases := []struct {
		src  any
		want string
	}{
		{nil, "0"}, {"12.50", "12.5"}, {[]byte("-13.25"), "-13.25"},
		{int(17), "17"}, {int64(-18), "-18"}, {float32(1.25), "1.25"}, {float64(-0.125), "-0.125"},
	}
	for _, tc := range cases {
		v := ParseCompatible("999")
		if err := v.Scan(tc.src); err != nil {
			t.Fatalf("Scan(%T): %v", tc.src, err)
		}
		if got := v.String(); got != tc.want {
			t.Fatalf("Scan(%T): got %s want %s", tc.src, got, tc.want)
		}
	}
	v := ParseCompatible("123.40")
	if err := v.Scan(struct{}{}); err == nil {
		t.Fatal("unsupported SQL type unexpectedly accepted")
	}
	if got := v.StringFixed(); got != "123.40" {
		t.Fatalf("failed Scan mutated destination: %s", got)
	}
	if err := v.UnmarshalJSON([]byte(`"nonsense"`)); err == nil {
		t.Fatal("invalid JSON decimal accepted")
	}
	if got := v.StringFixed(); got != "123.40" {
		t.Fatalf("failed JSON mutated destination: %s", got)
	}
	if err := v.UnmarshalText([]byte("nonsense")); err == nil {
		t.Fatal("invalid text decimal accepted")
	}
	if got := v.StringFixed(); got != "123.40" {
		t.Fatalf("failed text mutated destination: %s", got)
	}
}
