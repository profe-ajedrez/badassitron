package uint128

import (
	"github.com/jokruger/dec128/state"
	ju "github.com/jokruger/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
)

// FromString parses decimal digits using the imported Jokruger backend,
// translating only the public error code. Frozen vectors enforce legacy semantics.
func FromString[S string | []byte](s S) (Uint128, errors.Error) {
	n, status := ju.FromString(s)
	switch status {
	case state.OK:
		return Uint128(n), errors.None
	case state.Overflow:
		return Zero, errors.Overflow
	case state.InvalidFormat:
		return Zero, errors.InvalidFormat
	default:
		// Explicitly preserve rejection for unexpected future upstream states.
		return Zero, fromState(status)
	}
}
