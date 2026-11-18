package dec128bridge

import (
	"encoding/json"
	"math/big"
	"math/rand"
	"strings"
	"testing"

	legacy "github.com/profe-ajedrez/badassitron/dec128"
)

func TestMigrationParserLongInputBoundaries(t *testing.T) {
	inputs := []string{
		"123456789012345678.1234567890123456789",
		"-123456789012345678.1234567890123456789",
		"+123456789012345678.1234567890123456789",
		"340282366920938463463374607431768211455",
		"340282366920938463463374607431768211456",
		"0000000000000000000000000000000000000000000000000",
		"-0000000000000000000",
		"-00000000000000000000",
		"-0000000000000000000000000000000000000000000000.0000000000000000000",
		"0000000000000000000000000000000000000000000000.1000000000000000000",
		"0.0000000000000000000",
		"-0.0000000000000000000",
		"0.00000000000000000000",
		"-0.00000000000000000000",
		"0000000000000000000.",
		"00000000000000000000.",
		"-00000000000000000000.",
		"00000000000000000000.1.2",
		"00000000000000000000.123456789012345678x",
		"340282366920938463463374607431768211456x",
		"x340282366920938463463374607431768211456",
		"+000000000000000000000000000000000001.0000000000000000000",
		"--00000000000000000000000000000000001.0",
	}
	for _, in := range inputs {
		old, got := legacy.FromString(in), ParseCompatible(in)
		if old.IsNaN() != got.IsNaN() || (old.Coefficient().Lo != got.Coefficient().Lo || old.Coefficient().Hi != got.Coefficient().Hi) || old.Precision() != got.Precision() || old.StoredNegative() != got.StoredNegative() || old.ErrorDetails() != got.ErrorDetails() {
			t.Errorf("parse %q legacy=%#v (%v), bridge=%#v (%v)", in, old, old.ErrorDetails(), got, got.ErrorDetails())
		}
	}
}

func TestMigrationParserDeterministicBoundaryDifferential(t *testing.T) {
	r := rand.New(rand.NewSource(2026111709))
	for i := 0; i < 30000; i++ {
		n := 17 + r.Intn(55)
		var sb strings.Builder
		if r.Intn(3) == 0 {
			if r.Intn(2) == 0 {
				sb.WriteByte('-')
			} else {
				sb.WriteByte('+')
			}
		}
		place := r.Intn(n + 2)
		for j := 0; j < n; j++ {
			if j == place {
				sb.WriteByte('.')
			}
			if r.Intn(8) == 0 {
				sb.WriteByte('0')
			} else {
				sb.WriteByte(byte('0' + r.Intn(10)))
			}
		}
		if place == n {
			sb.WriteByte('.')
		}
		input := sb.String()
		old, got := legacy.FromString(input), ParseCompatible(input)
		if old.IsNaN() != got.IsNaN() || (old.Coefficient().Lo != got.Coefficient().Lo || old.Coefficient().Hi != got.Coefficient().Hi) || old.Precision() != got.Precision() || old.StoredNegative() != got.StoredNegative() || old.ErrorDetails() != got.ErrorDetails() {
			t.Fatalf("case=%d input=%q legacy=%#v (%v) bridge=%#v (%v)", i, input, old, old.ErrorDetails(), got, got.ErrorDetails())
		}
	}
}

func TestMigrationParserBigIntFiniteReference(t *testing.T) {
	r := rand.New(rand.NewSource(2026111710))
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	for i := 0; i < 10000; i++ {
		integerDigits := 1 + r.Intn(39)
		scale := r.Intn(20)
		var sb strings.Builder
		neg := r.Intn(2) == 0
		if neg {
			sb.WriteByte('-')
		}
		coefficientText := make([]byte, integerDigits+scale)
		for j := range coefficientText {
			coefficientText[j] = byte('0' + r.Intn(10))
		}
		sb.Write(coefficientText[:integerDigits])
		if scale > 0 {
			sb.WriteByte('.')
			sb.Write(coefficientText[integerDigits:])
		}
		input := sb.String()
		got := ParseCompatible(input)
		bigCoef, ok := new(big.Int).SetString(string(coefficientText), 10)
		if !ok {
			t.Fatal("bad test fixture")
		}
		if bigCoef.Cmp(max) > 0 {
			if !got.IsNaN() {
				t.Fatalf("%q should overflow", input)
			}
			continue
		}
		if got.IsNaN() {
			t.Fatalf("%q unexpectedly invalid", input)
		}
		v := got
		if new(big.Int).Add(new(big.Int).Lsh(new(big.Int).SetUint64(v.Coefficient().Hi), 64), new(big.Int).SetUint64(v.Coefficient().Lo)).Cmp(bigCoef) != 0 || int(v.Precision()) != scale || v.StoredNegative() != (neg && (bigCoef.Sign() != 0 || scale != 0)) {
			t.Fatalf("%q parsed %#v", input, v)
		}
	}
}

func TestMigrationWrappedNegativeZeroScaleZero(t *testing.T) {
	historical := legacy.FromString("-0.000").Rescale(0)
	if !historical.StoredNegative() {
		t.Fatal("fixture should be negative zero")
	}
	via := WrapLegacy(historical)
	back := via.Legacy()
	if back.Precision() != 0 || !back.StoredNegative() || !back.IsZero() || back != historical {
		t.Fatalf("negative zero scale zero was lost: %#v %#v", historical, back)
	}
	for _, encoded := range []func() ([]byte, error){via.MarshalText, via.MarshalJSON} {
		data, err := encoded()
		if err != nil {
			t.Fatal(err)
		}
		if len(data) == 0 {
			t.Fatal("empty output")
		}
	}
	j, err := json.Marshal(via)
	if err != nil || string(j) != "\"0\"" {
		t.Fatalf("json zero representation %s %v", j, err)
	}
}
