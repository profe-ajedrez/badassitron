package uint128_test

import (
	"math/rand"
	"testing"

	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	bu "github.com/profe-ajedrez/badassitron/dec128/uint128"
)

// Pin the COMPLETE set of states returned by Jokruger v1.4.0's unsigned
// decimal parser. The release facade uses a short state encoding that must be
// reviewed before upgrading Jokruger. Parsing itself remains 100% upstream.
func TestM7EParserStateDomainAndZeroOnError(t *testing.T) {
	cases := []struct {
		input string
		state state.State
		err   errors.Error
	}{
		{"", state.OK, errors.None},
		{"0", state.OK, errors.None},
		{"1234567890", state.OK, errors.None},
		{"340282366920938463463374607431768211455", state.OK, errors.None},
		{"340282366920938463463374607431768211456", state.Overflow, errors.Overflow},
		{"x", state.InvalidFormat, errors.InvalidFormat},
		{"123456789x", state.InvalidFormat, errors.InvalidFormat},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			up, st := ju.FromString(tc.input)
			if st != tc.state {
				t.Fatalf("upstream parser status changed: got %v, want %v", st, tc.state)
			}
			got, err := bu.FromString(tc.input)
			gotBytes, errBytes := bu.FromString([]byte(tc.input))
			if err != tc.err || got != bu.Uint128(up) || errBytes != err || gotBytes != got {
				t.Fatalf("parse %q: got %v, %v, expected %v, %v", tc.input, got, err, up, tc.err)
			}
			if err != errors.None && got != bu.Zero {
				t.Fatal("parser error must produce the historical zero result")
			}
		})
	}
	for _, st := range []struct{ upstream, local uint8 }{{0, 0}, {5, 3}, {9, 7}} {
		if got := st.upstream - 2*((st.upstream+3)>>3); got != st.local {
			t.Fatalf("state encoding %d -> %d, want %d", st.upstream, got, st.local)
		}
	}

	// Exercise both success and failure paths, beyond the existing 35k codec
	// differential corpus (which also includes leading-zero and wide values).
	rng := rand.New(rand.NewSource(0x4d37455))
	for i := 0; i < 12000; i++ {
		s := "1234567890"
		switch i % 6 {
		case 0:
			s = "123x456789"
		case 1:
			s = "340282366920938463463374607431768211456"
		case 2:
			s = "000000000000" // valid leading zeros
		case 3:
			s = "1000000000000000000000000000000000000000"
		case 4:
			s = string([]byte{'0' + byte(rng.Intn(10)), '0' + byte(rng.Intn(10))})
		}
		up, st := ju.FromString(s)
		var expected errors.Error
		switch st {
		case state.OK:
			expected = errors.None
		case state.Overflow:
			expected = errors.Overflow
		case state.InvalidFormat:
			expected = errors.InvalidFormat
		default:
			t.Fatalf("unexpected upstream parser state %v for %q", st, s)
		}
		if expected != errors.None && up != ju.Zero {
			t.Fatalf("upstream changed its zero-on-error contract for %q", s)
		}
		got, err := bu.FromString(s)
		if got != bu.Uint128(up) || err != expected {
			t.Fatalf("case %d: facade=%v/%d upstream=%v/%v", i, got, err, up, st)
		}
	}
}
