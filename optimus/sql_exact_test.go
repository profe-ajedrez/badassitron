package optimus

import (
	"database/sql"
	"database/sql/driver"
	"math"
	"math/big"
	"strings"
	"testing"
)

var _ sql.Scanner = (*Optimus)(nil)
var _ driver.Valuer = Optimus{}
var _ sql.Scanner = (*SQLDecimal)(nil)
var _ driver.Valuer = SQLDecimal{}

func TestSQLExactRoundTripAndBoundaries(t *testing.T) {
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	tooLarge := new(big.Int).Add(max, big.NewInt(1))
	cases := []struct {
		value string
		exact bool
	}{
		{"0", true},
		{"-0.000", true},
		{"+0.0100", true},
		{"12345678901234567890.12345678901234567890", true},
		{"-" + max.String(), true},
		{tooLarge.String(), false},
		{"0." + strings.Repeat("0", 153) + "1", true},
		{"0." + strings.Repeat("0", 154) + "1", false},
		{"1." + strings.Repeat("0", 154) + "1", false},
		{"1." + strings.Repeat("0", 154) + "0", true},
		{"", false},
		{"NaN", false},
		{"1e3", false},
		{" 1.2", false},
		{"1.2 ", false},
		{"1/2", false},
	}
	for _, tc := range cases {
		t.Run(tc.value[:min(len(tc.value), 30)], func(t *testing.T) {
			d := SQLDecimal{Decimal: FromInt(17), Valid: true}
			err := d.Scan([]byte(tc.value))
			if (err == nil) != tc.exact {
				t.Fatalf("input %q: error %v; want success %v", tc.value, err, tc.exact)
			}
			if !tc.exact {
				if !d.Valid || !d.Decimal.Equal(FromInt(17)) {
					t.Fatalf("failed scan mutated destination: %+v", d)
				}
				return
			}
			out, err := d.Value()
			if err != nil {
				t.Fatal(err)
			}
			inRat, ok := new(big.Rat).SetString(tc.value)
			if !ok {
				t.Fatalf("bad golden decimal %q", tc.value)
			}
			outRat, ok := new(big.Rat).SetString(out.(string))
			if !ok || inRat.Cmp(outRat) != 0 {
				t.Fatalf("SQL roundtrip %q => %q changed value", tc.value, out)
			}
		})
	}
}

func TestSQLExactNullableErrorsAndTypes(t *testing.T) {
	d := SQLDecimal{Decimal: FromInt(23), Valid: true}
	if err := d.Scan(nil); err != nil || d.Valid || !d.Decimal.IsZero() {
		t.Fatalf("SQL NULL: %+v, %v", d, err)
	}
	value, err := d.Value()
	if err != nil || value != nil {
		t.Fatalf("NULL Value: %v, %v", value, err)
	}
	if err := d.Scan(int64(math.MinInt64)); err != nil || !d.Valid || d.Decimal.String() != "-9223372036854775808" {
		t.Fatalf("int64: %+v, %v", d, err)
	}
	old := d
	for _, source := range []any{float32(0.1), float64(0.1), true, struct{}{}} {
		if err := d.Scan(source); err == nil || d != old {
			t.Fatalf("%T: invalid source accepted or destination mutated: %v", source, err)
		}
	}
	if err := (*SQLDecimal)(nil).Scan("1"); err == nil {
		t.Fatal("nil destination accepted")
	}
	bad := SQLDecimal{Decimal: NaN(0), Valid: true}
	if v, err := bad.Value(); err == nil || v != nil {
		t.Fatalf("NaN produced SQL driver value: %v, %v", v, err)
	}
	// Legacy direct SQL behavior stays available, without changing signatures.
	legacy := FromInt(42)
	if v, err := legacy.Value(); err != nil || v != "42" {
		t.Fatalf("legacy value: %v, %v", v, err)
	}
	if err := legacy.Scan("43.5"); err != nil || legacy.String() != "43.5" {
		t.Fatalf("legacy scan: %s, %v", legacy, err)
	}
}

func FuzzSQLExactNeverAltersValue(f *testing.F) {
	for _, s := range []string{"123.4500", "0." + strings.Repeat("0", 154) + "1", "-" + strings.Repeat("9", 155), "NaN", "", "0.000", "001.2", "3.14159"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 1024 {
			return
		}
		var d SQLDecimal
		if err := d.Scan(value); err != nil {
			return
		}
		if !d.Valid || d.Decimal.IsNaN() {
			t.Fatalf("accepted invalid decimal %q", value)
		}
		expected, ok := new(big.Rat).SetString(value)
		if !ok {
			t.Fatalf("accepted invalid syntax %q", value)
		}
		got, err := d.Value()
		if err != nil {
			t.Fatal(err)
		}
		actual, ok := new(big.Rat).SetString(got.(string))
		if !ok || actual.Cmp(expected) != 0 {
			t.Fatalf("decimal value changed from %q to %s", value, got)
		}
	})
}
