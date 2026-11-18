package dec128bridge

import (
	"encoding/json"
	"math/rand"
	"testing"

	jk "github.com/jokruger/dec128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// The public behavior is validated independently of its future implementation.
// These tests are deliberately in the experimental bridge package.
func TestMigrationConstructorsAndCodecs(t *testing.T) {
	cases := []string{"", "0", "-0", "-0.000", ".5", "5.", "5.0", "-123.1234567890123456789", "340282366920938463463374607431768211455", "340282366920938463463374607431768211456", "not-a-number", "1.2.3", "1e20"}
	for _, s := range cases {
		d := legacy.FromString(s)
		if !d.IsNaN() {
			text, err := d.MarshalText()
			if err != nil {
				t.Fatalf("%q text: %v", s, err)
			}
			var out legacy.Dec128
			if err = out.UnmarshalText(text); err != nil {
				t.Fatalf("%q text decode: %v", s, err)
			}
			if out.String() != d.String() {
				t.Fatalf("%q text roundtrip: %s vs %s", s, d, out)
			}
			encoded, err := json.Marshal(d)
			if err != nil {
				t.Fatalf("%q json: %v", s, err)
			}
			var decoded legacy.Dec128
			if err = json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("%q json decode: %v", s, err)
			}
			if decoded.String() != d.String() {
				t.Fatalf("%q json roundtrip: %s vs %s", s, d, decoded)
			}
			dst := legacy.FromInt(123)
			if err = legacy.ScanInto(&dst, d.StringFixed()); err != nil {
				t.Fatalf("%q SQL ScanInto: %v", s, err)
			}
			if !dst.Equal(d) {
				t.Fatalf("%q SQL mismatch: %s vs %s", s, d, dst)
			}
		}
	}
	// Error classifications are part of the old contract, not the upstream enum.
	if got := legacy.FromString("invalid"); got.ErrorDetails() != le.InvalidFormat.Value() {
		t.Fatalf("invalid input code: %v", got.ErrorDetails())
	}
	if got := legacy.New(lu.Uint128{Lo: 1}, 20, false); got.ErrorDetails() != le.PrecisionOutOfRange.Value() {
		t.Fatalf("out of range code: %v", got.ErrorDetails())
	}
}

func TestBridgeRoundTripFiniteAndErrorCodes(t *testing.T) {
	rng := rand.New(rand.NewSource(2026111708))
	for i := 0; i < 10000; i++ {
		d := legacy.New(lu.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 0)
		got := FromUpstream(ToUpstream(d))
		if !sameStrict(d, got) {
			t.Fatalf("roundtrip[%d] original=%#v result=%#v", i, d, got)
		}
	}
	for _, e := range []le.Error{le.NotANumber, le.DivisionByZero, le.Overflow, le.Underflow, le.InvalidFormat, le.PrecisionOutOfRange, le.RescaleToLessPrecision, le.SqrtNegative} {
		d := legacy.NaN(e)
		got := FromUpstream(ToUpstream(d))
		if !sameStrict(d, got) {
			t.Fatalf("error %v: %v vs %v", e, d.ErrorDetails(), got.ErrorDetails())
		}
	}
}

// This demonstrates a known limitation rather than silently claiming equivalence:
// the historical IsNegative intentionally hides the sign of signed zeros.
func TestSignedZeroRequiresExplicitPolicy(t *testing.T) {
	d := legacy.New(lu.Zero, 3, true)
	if !d.IsZero() || d.IsNegative() {
		t.Fatal("historical signed zero contract changed")
	}
	u := ToUpstream(d)
	if !u.IsZero() {
		t.Fatalf("upstream conversion changed zero: %v", u)
	}
	// Preserve the numerical value; representation-bit identity is NOT proven.
	if !FromUpstream(u).IsZero() {
		t.Fatal("upstream zero failed")
	}
}

func TestUpstreamParsingIsNotTheCompatibilityContract(t *testing.T) {
	// The wrapper must keep Badassitron's parser independent of upstream.
	input := "1e20"
	_ = jk.FromString(input)
	if !legacy.FromString(input).IsNaN() {
		t.Fatal("legacy parser unexpectedly accepted exponent notation")
	}
}
