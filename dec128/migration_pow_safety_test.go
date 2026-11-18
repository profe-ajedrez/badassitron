package dec128

import (
	"math"
	"testing"
)

func TestMigrationPowMinIntTerminates(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  string
	}{
		{"1", "1"}, {"-1", "1"},
	} {
		got := FromString(tc.input).PowInt(math.MinInt)
		if got.IsNaN() || got.String() != tc.want {
			t.Fatalf("PowInt(%s, MinInt) = %s (%v), want %s", tc.input, got.String(), got.ErrorDetails(), tc.want)
		}
	}
	// The historical fixed-width algorithm can signal overflow for large
	// magnitudes; the safety requirement is termination without a panic.
	_ = FromString("2").PowInt(math.MinInt)
}
