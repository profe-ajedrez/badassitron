package optimus

import (
	"fmt"
	"github.com/profe-ajedrez/badassitron/internal/decimalwide"
	"math/rand"
	"testing"
)

func legacyRescaleOracle(d Optimus, p uint8, m RoundingMode) (Optimus, bool) {
	if d.IsNaN() {
		return d, false
	}
	n, inx := decimalwide.Quantize(d.core(), int(p), 512, int(MaxPrecision), m)
	return pack(n), inx
}
func TestNativeRescaleRandomDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(1717))
	seeds := []string{"0", "-0.000", "10.25", "-10.25", "99999999999999999999999999999999999999", "0.123456789012345678901234567890", "-1234567890123456789012345678901234567890123456789012", "9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999"}
	for i := 0; i < 8000; i++ {
		s := seeds[rng.Intn(len(seeds))]
		if i >= len(seeds)*16 {
			digits := 1 + rng.Intn(90)
			x := make([]byte, 0, digits+3)
			if rng.Intn(2) == 0 {
				x = append(x, '-')
			}
			for k := 0; k < digits; k++ {
				v := byte('0' + rng.Intn(10))
				if k == 0 && v == '0' {
					v = '1'
				}
				x = append(x, v)
			}
			if rng.Intn(3) != 0 {
				pos := 1 + rng.Intn(digits)
				offset := 0
				if x[0] == '-' {
					offset = 1
				}
				at := offset + pos
				x = append(x[:at], append([]byte{'.'}, x[at:]...)...)
			}
			s = string(x)
		}
		d := FromString(s)
		if d.IsNaN() {
			continue
		}
		scales := []uint8{0, 1, 2, 10, 19, 28, 38, 76, 154, 155, 255, d.exp}
		p := scales[rng.Intn(len(scales))]
		m := RoundingMode(rng.Intn(10))
		want, wi := legacyRescaleOracle(d, p, m)
		got, gi := d.RescaleRoundWithStatus(p, m)
		if got.IsNaN() != want.IsNaN() || got.Precision() != want.Precision() || got.Width() != want.Width() || got.StringFixed() != want.StringFixed() || gi != wi || (got.IsNaN() && fmt.Sprint(got.ErrorDetails()) != fmt.Sprint(want.ErrorDetails())) {
			t.Fatalf("case %d input %q p=%d m=%d: want %q (err=%v, inexact=%v) got %q (err=%v, inexact=%v)", i, s, p, m, want.StringFixed(), want.ErrorDetails(), wi, got.StringFixed(), got.ErrorDetails(), gi)
		}
	}
}
