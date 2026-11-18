package dec128

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	jk "github.com/jokruger/dec128"
	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	le "github.com/profe-ajedrez/badassitron/dec128/errors"
)

func NewFromString(s string) (Dec128, error) {
	out := FromString(s)
	if out.IsNaN() {
		return Dec128{}, fmt.Errorf("invalid decimal number")
	}
	return out, nil
}
func NewFromInt(i int) Dec128                      { return FromInt(i) }
func FromString[S string | []byte](input S) Dec128 { return parse(string(input)) }
func parse(src string) Dec128 {
	if src == "" {
		return Zero
	}
	neg := false
	s := src
	if s[0] == '-' || s[0] == '+' {
		neg = s[0] == '-'
		s = s[1:]
	}
	if s == "" {
		return NaN(le.InvalidFormat)
	}
	short := len(src) <= ju.MaxSafeStrLen64
	if short {
		var small uint64
		var scale uint8
		seenDot := false
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c == '.' {
				if seenDot || len(s) == 1 {
					return NaN(le.InvalidFormat)
				}
				seenDot = true
				scale = uint8(len(s) - i - 1)
				continue
			}
			if c < '0' || c > '9' {
				return NaN(le.InvalidFormat)
			}
			small = small*10 + uint64(c-'0')
		}
		// <=19 source bytes guarantees a valid scale and uint64 factor.
		if small == 0 && scale == 0 {
			return Zero
		}
		return Dec128{coef: ju.Uint128{Lo: small}, scale: scale, state: signState(neg)}
	}
	dot := strings.IndexByte(s, '.')
	if dot == len(s)-1 {
		return NaN(le.InvalidFormat)
	}
	scale := 0
	if dot >= 0 {
		scale = len(s) - dot - 1
	}
	if scale > int(jk.MaxScale) {
		return NaN(le.PrecisionOutOfRange)
	}
	var coef ju.Uint128
	var st state.State
	if dot < 0 {
		coef, st = ju.FromString(s)
	} else {
		var whole, fraction ju.Uint128
		whole, st = ju.FromString(s[:dot])
		if st == state.OK {
			fraction, st = ju.FromString(s[dot+1:])
		}
		if st == state.OK {
			coef, st = whole.Mul64(ju.Pow10Uint64[scale])
		}
		if st == state.OK {
			coef, st = coef.Add64(fraction.Lo)
		}
	}
	if st != state.OK {
		return NaN(fromState(st.Error()))
	}
	out := fromParts(coef, uint8(scale), neg)
	if coef.IsZero() && neg && scale == 0 {
		out.state |= 1
	}
	return out
}
func (d *Dec128) UnmarshalText(b []byte) error {
	if d == nil {
		return fmt.Errorf("dec128: nil text destination")
	}
	out := FromString(b)
	if out.IsNaN() {
		return out.ErrorDetails()
	}
	*d = out
	return nil
}
func (d *Dec128) UnmarshalJSON(b []byte) error {
	if d == nil {
		return fmt.Errorf("dec128: nil json destination")
	}
	if len(b) == 0 || string(b) == "null" {
		*d = Zero
		return nil
	}
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	return d.UnmarshalText(b)
}
func (d Dec128) Scan(v any) error { // retained value receiver for interface compatibility
	return ScanInto(&d, v)
}

// ScanInto does not allocate an intermediate []byte or decimal text for
// string/byte/integer sources. Parse into a local value first so invalid
// database input never mutates the destination.
func ScanInto(dst *Dec128, src any) error {
	if dst == nil {
		return fmt.Errorf("dec128: nil scan destination")
	}
	var parsed Dec128
	switch v := src.(type) {
	case nil:
		parsed = Zero
	case string:
		parsed = FromString(v)
	case []byte:
		parsed = FromString(v)
	case int:
		parsed = FromInt(v)
	case int64:
		parsed = FromInt64(v)
	case float32:
		parsed = FromString(strconv.FormatFloat(float64(v), 'f', -1, 32))
	case float64:
		parsed = FromFloat64(v)
	default:
		return fmt.Errorf("can't scan %T to Dec128: %T is not supported", src, src)
	}
	if parsed.IsNaN() {
		return parsed.ErrorDetails()
	}
	*dst = parsed
	return nil
}

type scanDestination struct{ dst *Dec128 }

func (s scanDestination) Scan(v any) error    { return ScanInto(s.dst, v) }
func ScanDestination(dst *Dec128) sql.Scanner { return scanDestination{dst} }
