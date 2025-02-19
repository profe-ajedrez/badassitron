package unit

import (
	"fmt"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
)

func TestDecimalUint128Encoding(t *testing.T) {
	dec128.SetDefaultPrecision(19)

	type testCase struct {
		i string
		p uint8
		s string
	}

	testCases := [...]testCase{
		{"0", 6, "0"},
		{"1", 6, "1"},
		{"1.1", 6, "1.1"},
		{"1.01", 6, "1.01"},
		{"123.456", 6, "123.456"},
		{"1234567890.1234567890", 6, "1234567890.123456"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("TestDecimalFromUint128(%v)", tc), func(t *testing.T) {
			d := dec128.FromString(tc.i)
			u, err := d.EncodeToUint128(tc.p)
			if err != nil {
				t.Errorf("error: %v", err)
			}
			s := dec128.New(u, tc.p, false).String()
			if s != tc.s {
				t.Errorf("expected: %v, got: %v", tc.s, s)
			}
		})
	}
}
