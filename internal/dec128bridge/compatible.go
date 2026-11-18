package dec128bridge

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	jk "github.com/jokruger/dec128"
	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
)

// Compatible is a migration-only representation that preserves stored signed
// zero. Parsing, formatting and SQL no longer invoke the historical engine.
// The value remains private so upstream normalization cannot erase metadata.
type Compatible struct {
	value        jk.Dec128
	zeroNegative bool
	zeroScale    uint8
}

func WrapLegacy(v legacy.Dec128) Compatible {
	out := Compatible{value: ToUpstream(v)}
	if v.IsZero() {
		out.zeroNegative, out.zeroScale = v.StoredNegative(), v.Precision()
	}
	return out
}

func NewCompatible(c ju.Uint128, scale uint8, negative bool) Compatible {
	out := Compatible{value: jk.New(c, scale, negative)}
	if c.IsZero() && !out.value.IsNaN() && scale != 0 {
		out.zeroNegative, out.zeroScale = negative, scale
	}
	return out
}

// Legacy exists solely for migration interoperability and differential tests.
func (v Compatible) Legacy() legacy.Dec128 {
	if v.value.IsZero() {
		return legacyFiniteZero(v.zeroScale, v.zeroNegative)
	}
	return FromUpstream(v.value)
}
func (v Compatible) IsNaN() bool  { return v.value.IsNaN() }
func (v Compatible) IsZero() bool { return v.value.IsZero() }

// Precision returns the stored fractional scale, including for signed zeros.
func (v Compatible) Precision() uint8 {
	if v.value.IsZero() {
		return v.zeroScale
	}
	return v.value.Scale()
}

// StoredNegative exposes the finite sign bit, not just the numerical sign.
func (v Compatible) StoredNegative() bool {
	if v.value.IsZero() {
		return v.zeroNegative
	}
	return !v.value.IsNaN() && v.value.IsNegative()
}

func (v Compatible) Coefficient() ju.Uint128 { return v.value.Coefficient() }

// ErrorDetails maps the upstream state to the historical error vocabulary.
func (v Compatible) ErrorDetails() error {
	if !v.IsNaN() {
		return nil
	}
	return fromState(v.value.ErrorDetails()).Value()
}

func (v Compatible) StringFixed() string {
	if v.IsNaN() {
		return "NaN"
	}
	if v.value.IsZero() {
		if v.zeroScale == 0 {
			return "0"
		}
		return "0." + strings.Repeat("0", int(v.zeroScale))
	}
	return v.value.StringFixed()
}
func (v Compatible) String() string {
	if v.IsNaN() {
		return "NaN"
	}
	if v.value.IsZero() {
		return "0"
	}
	return v.value.String()
}
func (v Compatible) MarshalText() ([]byte, error) { return []byte(v.String()), nil }
func (v Compatible) MarshalJSON() ([]byte, error) { return json.Marshal(v.String()) }
func (v Compatible) Value() (driver.Value, error) { return v.String(), nil }

// legacyFiniteZero constructs signed zero without invoking legacy arithmetic.
// The old New constructor intentionally normalizes negative zero at scale 0.
func legacyFiniteZero(scale uint8, negative bool) legacy.Dec128 {
	return legacy.NewStoredZero(scale, negative)
}

// parseCompatible is independent of the legacy decimal parser. It uses
// upstream's exported unsigned decimal parser and constructs with jk.New.
// Historical empty input is zero; historical 5. is accepted only on its
// short-number path. Scientific notation remains unsupported.
func parseCompatible(src string) Compatible {
	if src == "" {
		return NewCompatible(ju.Zero, 0, false)
	}
	neg := false
	s := src
	if s[0] == '-' || s[0] == '+' {
		neg = s[0] == '-'
		s = s[1:]
	}
	if s == "" {
		return Compatible{value: jk.NaN(state.InvalidFormat)}
	}
	// For short inputs, validate and accumulate the coefficient in one pass.
	// For long inputs, let upstream validate each digit while parsing its
	// original substring, avoiding an additional full-input validation pass.
	short := len(src) <= ju.MaxSafeStrLen64
	if short {
		dot := -1
		var small uint64
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c == '.' {
				if dot >= 0 {
					return Compatible{value: jk.NaN(state.InvalidFormat)}
				}
				dot = i
				continue
			}
			if c < '0' || c > '9' {
				return Compatible{value: jk.NaN(state.InvalidFormat)}
			}
			small = small*10 + uint64(c-'0')
		}
		if s == "." || dot == len(s)-1 && len(src) > ju.MaxSafeStrLen64 {
			return Compatible{value: jk.NaN(state.InvalidFormat)}
		}
		scale := 0
		if dot >= 0 {
			scale = len(s) - dot - 1
		}
		if scale > int(jk.MaxScale) {
			return Compatible{value: jk.NaN(state.ScaleOutOfRange)}
		}
		return NewCompatible(ju.Uint128{Lo: small}, uint8(scale), neg)
	}

	dot := strings.IndexByte(s, '.')
	if dot == len(s)-1 {
		// Historical long-number parsing rejects trailing decimal points.
		return Compatible{value: jk.NaN(state.InvalidFormat)}
	}
	scale := 0
	if dot >= 0 {
		scale = len(s) - dot - 1
	}
	if scale > int(jk.MaxScale) {
		// The historical long-number path checks fractional scale before
		// parsing any digits, including potentially invalid ones.
		return Compatible{value: jk.NaN(state.ScaleOutOfRange)}
	}
	var c ju.Uint128
	var st state.State
	if dot < 0 {
		c, st = ju.FromString(s)
	} else {
		var whole, fraction ju.Uint128
		whole, st = ju.FromString(s[:dot])
		if st == state.OK {
			fraction, st = ju.FromString(s[dot+1:])
		}
		if st == state.OK {
			c, st = whole.Mul64(ju.Pow10Uint64[scale])
		}
		if st == state.OK {
			c, st = c.Add64(fraction.Lo)
		}
	}
	if st != state.OK {
		return Compatible{value: jk.NaN(st)}
	}
	out := NewCompatible(c, uint8(scale), neg)
	if c.IsZero() && neg && scale == 0 {
		// The legacy long-integer parser returns the raw negative sign of
		// an all-zero coefficient. The short parser normalizes it instead.
		out.zeroNegative = true
	}
	return out
}

func ParseCompatible(src string) Compatible { return parseCompatible(src) }

func (v *Compatible) UnmarshalText(data []byte) error {
	if v == nil {
		return fmt.Errorf("dec128bridge: nil text destination")
	}
	parsed := parseCompatible(string(data))
	if parsed.IsNaN() {
		return fromState(parsed.value.ErrorDetails()).Value()
	}
	*v = parsed
	return nil
}
func (v *Compatible) UnmarshalJSON(data []byte) error {
	if v == nil {
		return fmt.Errorf("dec128bridge: nil json destination")
	}
	if len(data) == 0 || string(data) == "null" {
		*v = NewCompatible(ju.Zero, 0, false)
		return nil
	}
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		// Preserve legacy behavior: strip enclosing quotes, not JSON escapes.
		data = data[1 : len(data)-1]
	}
	return v.UnmarshalText(data)
}
func (v *Compatible) Scan(src any) error {
	if v == nil {
		return fmt.Errorf("dec128: nil scan destination")
	}
	var s string
	switch x := src.(type) {
	case nil:
		s = ""
	case string:
		s = x
	case []byte:
		s = string(x)
	case int:
		s = strconv.FormatInt(int64(x), 10)
	case int64:
		s = strconv.FormatInt(x, 10)
	case float32:
		s = strconv.FormatFloat(float64(x), 'f', -1, 32)
	case float64:
		s = strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return fmt.Errorf("can't scan %T to Dec128: %T is not supported", src, src)
	}
	return v.UnmarshalText([]byte(s))
}
