// Package externalclient exercises only exported APIs, allowing this fixture
// to be copied into an independent module for an actual third-party build.
package externalclient_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec256"
	"github.com/profe-ajedrez/badassitron/dec512"
	"github.com/profe-ajedrez/badassitron/optimus"
)

var (
	_ fmt.Stringer             = dec128.Dec128{}
	_ encoding.TextMarshaler   = dec128.Dec128{}
	_ encoding.TextUnmarshaler = (*dec128.Dec128)(nil)
	_ json.Marshaler           = dec128.Dec128{}
	_ json.Unmarshaler         = (*dec128.Dec128)(nil)
	_ driver.Valuer            = dec128.Dec128{}
	_ sql.Scanner              = dec128.Dec128{} // Historical value-receiver contract.
	_ interface {
		Add(dec128.Dec128) dec128.Dec128
	} = dec128.Dec128{}
	_ interface {
		MulRoundWithStatus(dec128.Dec128, uint8, dec128.RoundingMode) (dec128.Dec128, bool)
	} = dec128.Dec128{}
)

func TestExternalDecimalContract(t *testing.T) {
	v := dec128.New(uint128.Uint128{Lo: 12340}, 3, false)
	if v.StringFixed() != "12.340" || v.String() != "12.34" || v.Precision() != 3 {
		t.Fatalf("stored scale mismatch: %v, %d", v, v.Precision())
	}
	if fromBytes := dec128.FromString([]byte("12.340")); !fromBytes.Equal(v) {
		t.Fatalf("generic bytes input mismatch: %v", fromBytes)
	}
	if got := v.Add(dec128.FromString("0.660")); got.StringFixed() != "13.000" {
		t.Fatalf("addition mismatch: %s", got.StringFixed())
	}
	signedZero := dec128.NewStoredZero(3, true)
	if !signedZero.IsZero() || !signedZero.StoredNegative() || signedZero.Precision() != 3 {
		t.Fatal("negative zero was not retained")
	}
	if got := signedZero.Rescale(0); !got.StoredNegative() || got.Precision() != 0 {
		t.Fatal("rescale lost stored negative zero")
	}
	n := dec128.NaN(errors.Overflow)
	if got := n.Mul(dec128.One); !got.IsNaN() || got.ErrorDetails() != errors.Overflow.Value() {
		t.Fatalf("error propagation mismatch: %s", got.String())
	}
	if dec128.RoundingHalfEven != dec256.RoundingHalfEven {
		t.Fatal("dec256 rounding mode diverged")
	}
	if dec128.RoundingHalfEven != dec512.RoundingHalfEven {
		t.Fatal("dec512 rounding mode diverged")
	}
	if dec128.RoundingHalfEven != optimus.RoundingHalfEven {
		t.Fatal("optimus rounding mode diverged")
	}
}

func TestExternalCodecAndSQL(t *testing.T) {
	value := dec128.FromString("123.450")
	raw, err := json.Marshal(value)
	if err != nil || string(raw) != `"123.45"` {
		t.Fatalf("JSON marshaling: %s %v", raw, err)
	}
	var dst dec128.Dec128
	if err := json.Unmarshal(raw, &dst); err != nil || dst.String() != "123.45" {
		t.Fatalf("JSON parsing: %s %v", dst, err)
	}
	if err := dec128.ScanDestination(&dst).Scan("44.20"); err != nil || dst.StringFixed() != "44.20" {
		t.Fatalf("SQL scanner: %s %v", dst.StringFixed(), err)
	}
	if v, err := dst.Value(); err != nil || v != "44.2" {
		t.Fatalf("driver value: %v %v", v, err)
	}
}

func TestExternalWideConsumers(t *testing.T) {
	a := "12345678901234567890.12"
	for label, got := range map[string]string{
		"256":     dec256.FromString(a).String(),
		"512":     dec512.FromString(a).String(),
		"optimus": optimus.FromString(a).String(),
		"128":     dec128.FromString(a).String(),
	} {
		if got != a {
			t.Fatalf("%s parsed %s, want %s", label, got, a)
		}
	}
	// This multiplication requires more than a 128-bit coefficient.
	x := "99999999999999999999999999999"
	if got := dec128.FromString(x).Mul(dec128.FromString(x)); !got.IsNaN() {
		t.Fatalf("dec128 should not silently approximate multiplication: %s", got)
	}
	if got := dec256.FromString(x).Mul(dec256.FromString(x)); got.IsNaN() {
		t.Fatalf("dec256 lost wide multiply: %v", got.ErrorDetails())
	}
	if got := dec512.FromString(x).Mul(dec512.FromString(x)); got.IsNaN() {
		t.Fatalf("dec512 lost wide multiply: %v", got.ErrorDetails())
	}
}

// M3 public uint128 must preserve the historical type and error-code signatures,
// even though its optional arithmetic implementation imports upstream.
var (
	_ fmt.Scanner              = (*uint128.Uint128)(nil)
	_ fmt.Stringer             = uint128.Uint128{}
	_ encoding.TextMarshaler   = uint128.Uint128{}
	_ encoding.TextUnmarshaler = (*uint128.Uint128)(nil)
	_ interface {
		Add(uint128.Uint128) (uint128.Uint128, errors.Error)
		MulCarry(uint128.Uint128) (uint128.Uint128, uint128.Uint128)
		QuoRem(uint128.Uint128) (uint128.Uint128, uint128.Uint128, errors.Error)
	} = uint128.Uint128{}
)

func TestExternalUint128Contract(t *testing.T) {
	a := uint128.Uint128{Lo: 0, Hi: 1}
	if got, code := a.Sub(uint128.One); code != errors.None || got.Lo != ^uint64(0) || got.Hi != 0 {
		t.Fatalf("unsigned subtraction: %v %v", got, code)
	}
	if got, code := uint128.Max.Add(uint128.One); code != errors.Overflow || got != uint128.Zero {
		t.Fatalf("overflow error contract: %v %v", got, code)
	}
	if q, r, code := a.QuoRem(uint128.Zero); code != errors.DivisionByZero || q != uint128.Zero || r != uint128.Zero {
		t.Fatalf("division-by-zero contract: %v %v %v", q, r, code)
	}
	if lo, hi := a.MulCarry(a); lo != uint128.Zero || hi != uint128.One {
		t.Fatalf("wide product halves: %v %v", lo, hi)
	}
	if b := a.BytesBigEndian(); uint128.FromBytesBigEndian(b) != a {
		t.Fatal("binary roundtrip changed")
	}
}
