package optimus

import (
	"bytes"
	"encoding/json"
	"math/big"
	"strings"
	"testing"
)

func TestPhaseCRepresentationsAgreeAcrossCoefficientTiers(t *testing.T) {
	values := []string{
		"0", "-0.000000", "1.000000000000000", "18446744073709551615",
		"18446744073709551616", "340282366920938463463374607431768211456",
		"123456789012345678901234567890123456789012345678901234567890.0001",
	}
	for _, source := range values {
		d := FromString(source)
		if d.IsNaN() {
			t.Fatalf("unrepresentable fixture %q", source)
		}
		normalized := FromString(d.String())
		if !d.Equal(normalized) || d.Compare(normalized) != 0 {
			t.Fatalf("representation mismatch %q", source)
		}
		if d.Canonical().Compare(d) != 0 {
			t.Fatalf("canonical changes value %q", source)
		}
		raw, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		var decoded Optimus
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.Compare(d) != 0 {
			t.Fatalf("JSON changed value %q => %q", source, raw)
		}
		binaryText, err := d.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		var again Optimus
		if err := again.UnmarshalText(binaryText); err != nil || again.Compare(d) != 0 {
			t.Fatalf("text changed value %q: %v", source, err)
		}
	}
}

func TestPhaseCSerializationFailureDoesNotMutateDestination(t *testing.T) {
	original := FromInt(197)
	for _, raw := range [][]byte{[]byte("invalid"), []byte("1e4"), []byte("."), []byte(strings.Repeat("9", 10001))} {
		d := original
		if err := d.UnmarshalText(raw); err == nil || !d.Equal(original) {
			t.Fatalf("text %q changed destination or accepted invalid input", raw[:min(len(raw), 20)])
		}
		d = original
		quoted, _ := json.Marshal(string(raw))
		if err := json.Unmarshal(quoted, &d); err == nil || !d.Equal(original) {
			t.Fatalf("JSON %q changed destination or accepted invalid input", raw[:min(len(raw), 20)])
		}
	}
}

func TestPhaseCSQLIndependentOfStoredScale(t *testing.T) {
	for _, str := range []string{"123.4500", "000123.45", "+123.45000000000000000"} {
		var d SQLDecimal
		if err := d.Scan(str); err != nil {
			t.Fatal(err)
		}
		val, err := d.Value()
		if err != nil {
			t.Fatal(err)
		}
		left, _ := new(big.Rat).SetString(str)
		right, _ := new(big.Rat).SetString(val.(string))
		if left.Cmp(right) != 0 {
			t.Fatalf("value changed: %q -> %q", str, val)
		}
		if !bytes.Equal([]byte(d.Decimal.String()), []byte(val.(string))) {
			t.Fatal("Value differs from decimal representation")
		}
	}
}
