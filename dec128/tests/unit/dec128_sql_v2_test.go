package unit

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
)

func TestV2ScanIntoAndSQLDestination(t *testing.T) {
	tests := []struct {
		name string
		src  any
		want string
	}{
		{"string", "123.456789012345", "123.456789012345"},
		{"bytes", []byte("-0.010500"), "-0.0105"},
		{"int", int(-42), "-42"},
		{"int64", int64(985), "985"},
		{"float32", float32(1.234567), "1.234567"},
		{"float64", float64(1.25), "1.25"},
		{"null", nil, "0"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			value := dec128.FromString("77")
			if err := dec128.ScanInto(&value, tc.src); err != nil || value.String() != tc.want {
				t.Fatalf("ScanInto(%T) => %s, %v, want %s", tc.src, value, err, tc.want)
			}
			value = dec128.FromString("77")
			scanner := dec128.ScanDestination(&value)
			var _ sql.Scanner = scanner
			if err := scanner.Scan(tc.src); err != nil || value.String() != tc.want {
				t.Fatalf("SQL scanner(%T) => %s, %v, want %s", tc.src, value, err, tc.want)
			}
		})
	}
}

func TestV2ScanPreservesDestinationOnError(t *testing.T) {
	original := dec128.FromString("19.25")
	for _, v := range []any{"wrong", []byte("1.12345678901234567890"), struct{}{}, true} {
		x := original
		if err := dec128.ScanInto(&x, v); err == nil || !x.Equal(original) {
			t.Fatalf("ScanInto(%T) err=%v, value=%s: want unchanged", v, err, x)
		}
	}
	if err := dec128.ScanInto(nil, "1"); err == nil || !strings.Contains(err.Error(), "nil") {
		t.Fatalf("nil scan target must report error; got %v", err)
	}
	if err := dec128.ScanDestination(nil).Scan("1"); err == nil {
		t.Fatal("nil SQL destination must report error")
	}
}
